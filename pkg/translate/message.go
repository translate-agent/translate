package translate

import (
	"context"
	"errors"
	"fmt"

	"go.expect.digital/translate/pkg/model"
	"golang.org/x/text/language"

	"github.com/google/uuid"
	translatev1 "go.expect.digital/translate/pkg/pb/translate/v1"
	"go.expect.digital/translate/pkg/repo/common"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ----------------------ListMessages-------------------------------

type listMessagesParams struct {
	serviceID uuid.UUID
}

func parseListMessagesRequestParams(req *translatev1.ListMessagesRequest) (*listMessagesParams, error) {
	serviceID, err := uuidFromProto(req.GetServiceId())
	if err != nil {
		return nil, fmt.Errorf("parse service_id: %w", err)
	}

	return &listMessagesParams{serviceID: serviceID}, nil
}

func (l *listMessagesParams) validate() error {
	if l.serviceID == uuid.Nil {
		return errors.New("'service_id' is required")
	}

	return nil
}

func (t *TranslateServiceServer) ListMessages(
	ctx context.Context,
	req *translatev1.ListMessagesRequest,
) (*translatev1.ListMessagesResponse, error) {
	params, err := parseListMessagesRequestParams(req)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	if err = params.validate(); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	messages, err := t.repo.LoadMessages(ctx, params.serviceID, common.LoadMessagesOpts{})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "")
	}

	return &translatev1.ListMessagesResponse{Messages: messagesSliceToProto(messages)}, nil
}

// ----------------------UpdateMessages-------------------------------

type updateMessagesParams struct {
	messages  *model.Messages
	serviceID uuid.UUID
}

func parseUpdateMessagesRequestParams(req *translatev1.UpdateMessagesRequest) (*updateMessagesParams, error) {
	serviceID, err := uuidFromProto(req.GetServiceId())
	if err != nil {
		return nil, fmt.Errorf("parse service_id: %w", err)
	}

	requestLanguageTag, err := language.Parse(req.GetMessages().Language)
	if err != nil {
		return nil, fmt.Errorf("parse language: %w", err)
	}

	messages := &model.Messages{
		Language: requestLanguageTag,
		Messages: make([]model.Message, len(req.GetMessages().GetMessages())),
	}

	for i, pbMsg := range req.GetMessages().GetMessages() {
		messages.Messages[i] = model.Message{
			ID:          pbMsg.GetId(),
			Message:     pbMsg.GetMessage(),
			Description: pbMsg.GetDescription(),
			Fuzzy:       pbMsg.GetFuzzy(),
		}
	}

	return &updateMessagesParams{serviceID: serviceID, messages: messages}, nil
}

func (u *updateMessagesParams) validate() error {
	if u.serviceID == uuid.Nil {
		return errors.New("'service_id' is required")
	}

	return nil
}

func (t *TranslateServiceServer) UpdateMessages(
	ctx context.Context,
	req *translatev1.UpdateMessagesRequest,
) (*translatev1.UpdateMessagesResponse, error) {
	params, err := parseUpdateMessagesRequestParams(req)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	if err = params.validate(); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	err = t.repo.SaveMessages(ctx, params.serviceID, params.messages)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update messages: %v", err)
	}

	// Construct and return the response
	response := &translatev1.UpdateMessagesResponse{
		Messages: messagesToProto(params.messages),
	}

	return response, nil
}

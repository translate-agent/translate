package translate

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	translatev1 "go.expect.digital/translate/pkg/pb/translate/v1"
	"go.expect.digital/translate/pkg/repo/common"
	"golang.org/x/text/language"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ----------------------ListMessages-------------------------------

type listMessagesParams struct {
	languageTags []language.Tag
	serviceID    uuid.UUID
}

func parseListMessagesRequestParams(req *translatev1.ListMessagesRequest) (*listMessagesParams, error) {
	serviceID, err := uuidFromProto(req.GetServiceId())
	if err != nil {
		return nil, fmt.Errorf("parse service_id: %w", err)
	}

	if len(req.GetLanguages()) == 1 && strings.Contains(req.GetLanguages()[0], ",") {
		req.Languages = strings.Split(req.GetLanguages()[0], ",")
	}

	langTags, err := langTagsFromProto(req.GetLanguages())
	if err != nil {
		return nil, fmt.Errorf("parse languages: %w", err)
	}

	return &listMessagesParams{serviceID: serviceID, languageTags: langTags}, nil
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

	messages, err := t.repo.LoadMessages(ctx, params.serviceID,
		common.LoadMessagesOpts{FilterLanguages: params.languageTags})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "")
	}

	return &translatev1.ListMessagesResponse{Messages: messagesSliceToProto(messages)}, nil
}

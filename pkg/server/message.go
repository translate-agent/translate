package server

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"go.expect.digital/translate/pkg/model"
	translatev1 "go.expect.digital/translate/pkg/pb/translate/v1"
	"go.expect.digital/translate/pkg/repo/common"
	"golang.org/x/exp/slices"
	"golang.org/x/text/language"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ----------------------CreateMessages-------------------------------

type createMessagesParams struct {
	messages  *model.Messages
	serviceID uuid.UUID
}

func parseCreateMessagesRequestParams(req *translatev1.CreateMessagesRequest) (*createMessagesParams, error) {
	var (
		p   = &createMessagesParams{}
		err error
	)

	if p.serviceID, err = uuidFromProto(req.ServiceId); err != nil {
		return nil, fmt.Errorf("parse service_id: %w", err)
	}

	if p.messages, err = messagesFromProto(req.Messages); err != nil {
		return nil, fmt.Errorf("parse messages: %w", err)
	}

	return p, nil
}

func (c *createMessagesParams) validate() error {
	if c.serviceID == uuid.Nil {
		return errors.New("'service_id' is required")
	}

	if c.messages == nil {
		return errors.New("'messages' is required")
	}

	if c.messages.Language == language.Und {
		return fmt.Errorf("invalid messages: %w", errors.New("'language' is required"))
	}

	return nil
}

func (t *TranslateServiceServer) CreateMessages(
	ctx context.Context,
	req *translatev1.CreateMessagesRequest,
) (*translatev1.Messages, error) {
	params, err := parseCreateMessagesRequestParams(req)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	if err = params.validate(); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	msgs, err := t.repo.LoadMessages(ctx, params.serviceID,
		common.LoadMessagesOpts{FilterLanguages: []language.Tag{params.messages.Language}})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "")
	}

	if len(msgs) > 0 {
		return nil, status.Errorf(codes.AlreadyExists, "messages already exist for language: '%s'", params.messages.Language)
	}

	// If the messages are not the original messages, translate them.
	if !params.messages.Original {
		var err error

		params.messages, err = t.translator.Translate(ctx, params.messages, params.messages.Language)
		if err != nil {
			return nil, status.Errorf(codes.Unknown, err.Error()) // TODO: For now we don't know the cause of the error.
		}
	}

	if err := t.repo.SaveMessages(ctx, params.serviceID, params.messages); errors.Is(err, common.ErrNotFound) {
		return nil, status.Errorf(codes.NotFound, "service not found")
	} else if err != nil {
		return nil, status.Errorf(codes.Internal, "")
	}

	return messagesToProto(params.messages), nil
}

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
	messages             *model.Messages
	serviceID            uuid.UUID
	populateTranslations bool
}

func parseUpdateMessagesRequestParams(req *translatev1.UpdateMessagesRequest) (*updateMessagesParams, error) {
	var (
		params = updateMessagesParams{populateTranslations: req.GetPopulateTranslations()}
		err    error
	)

	if params.serviceID, err = uuidFromProto(req.ServiceId); err != nil {
		return nil, fmt.Errorf("parse service_id: %w", err)
	}

	if params.messages, err = messagesFromProto(req.Messages); err != nil {
		return nil, fmt.Errorf("parse messages: %w", err)
	}

	return &params, nil
}

func (u *updateMessagesParams) validate() error {
	if u.serviceID == uuid.Nil {
		return errors.New("'service_id' is required")
	}

	if u.messages == nil {
		return errors.New("'messages' is nil")
	}

	if u.messages.Language == language.Und {
		return errors.New("'language' is required")
	}

	return nil
}

func (t *TranslateServiceServer) UpdateMessages(
	ctx context.Context,
	req *translatev1.UpdateMessagesRequest,
) (*translatev1.Messages, error) {
	params, err := parseUpdateMessagesRequestParams(req)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	if err = params.validate(); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	msgs, err := t.repo.LoadMessages(
		ctx,
		params.serviceID,
		common.LoadMessagesOpts{FilterLanguages: []language.Tag{}})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "")
	}

	if len(msgs) == 0 || !langExists(msgs, params.messages.Language) {
		return nil, status.Errorf(codes.NotFound, "messages not found for language: '%s'", params.messages.Language)
	}

	err = t.repo.SaveMessages(ctx, params.serviceID, params.messages)
	switch {
	default:
		// noop
	case errors.Is(err, common.ErrNotFound):
		return nil, status.Errorf(codes.NotFound, "service not found")
	case err != nil:
		return nil, status.Errorf(codes.Internal, "")
	}

	// If updating the original messages and populateTranslations flag is true, update all translated messages as well.
	if params.messages.Original && params.populateTranslations {
		if err := t.populateTranslatedMessages(ctx, params.serviceID, params.messages, msgs); err != nil {
			return nil, err
		}
	}

	return messagesToProto(params.messages), nil
}

// helpers

// populateTranslatedMessages adds any missing messages from the original messages to all translated messages.
func (t *TranslateServiceServer) populateTranslatedMessages(
	ctx context.Context,
	serviceID uuid.UUID,
	originalMessages *model.Messages,
	allMessages []model.Messages,
) error {
	// Iterate over the existing messages
	for _, messages := range allMessages {
		// Skip if Original is true
		if messages.Original {
			continue
		}

		// Create a map to store the IDs of the translated messages
		translatedMessageIDs := make(map[string]struct{}, len(messages.Messages))
		for _, m := range messages.Messages {
			translatedMessageIDs[m.ID] = struct{}{}
		}

		// Create a new messages to store the messages that need to be translated
		toBeTranslated := &model.Messages{
			Language: messages.Language,
			Original: messages.Original,
			Messages: make([]model.Message, 0, len(originalMessages.Messages)),
		}

		// Iterate over the original messages and add any missing messages to the toBeTranslated.Messages slice
		for _, m := range originalMessages.Messages {
			if _, ok := translatedMessageIDs[m.ID]; !ok {
				toBeTranslated.Messages = append(toBeTranslated.Messages, m)
			}
		}

		// Translate the messages
		translated, err := t.translator.Translate(ctx, toBeTranslated, messages.Language)
		if err != nil {
			return status.Errorf(codes.Unknown, err.Error()) // TODO: For now we don't know the cause of the error.
		}

		// Append the translated messages to the existing messages
		messages.Messages = append(messages.Messages, translated.Messages...)

		// Save the updated translated messages
		switch err := t.repo.SaveMessages(ctx, serviceID, &messages); {
		default:
			// noop
		case errors.Is(err, common.ErrNotFound):
			return status.Errorf(codes.NotFound, "service not found")
		case err != nil:
			return status.Errorf(codes.Internal, "")
		}
	}

	return nil
}

// langExists returns true if the provided language exists in the provided model.Messages slice.
func langExists(msgs []model.Messages, lang language.Tag) bool {
	return slices.ContainsFunc(msgs, func(m model.Messages) bool {
		return m.Language == lang
	})
}

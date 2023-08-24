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

	// Translate messages when messages are not original and original language is known.
	if !params.messages.Original {
		// Retrieve language from original messages.
		var originalLanguage *language.Tag
		// TODO: to improve performance should be replaced with CheckMessagesExist db function.
		msgs, err := t.repo.LoadMessages(ctx, params.serviceID, common.LoadMessagesOpts{})
		if err != nil {
			return nil, status.Errorf(codes.Internal, "")
		}

		for _, v := range msgs {
			if v.Original {
				originalLanguage = &v.Language

				// if incoming messages are empty populate with original messages.
				if params.messages.Messages == nil {
					params.messages.Messages = v.Messages
				}

				break
			}
		}

		if originalLanguage != nil {
			// Translate messages -
			// untranslated text in incoming messages will be translated from original to target language.
			targetLanguage := params.messages.Language
			params.messages.Language = *originalLanguage
			params.messages, err = t.translator.Translate(ctx, params.messages, targetLanguage)
			if err != nil {
				return nil, status.Errorf(codes.Unknown, err.Error()) // TODO: For now we don't know the cause of the error.
			}
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

	allMessages, err := t.repo.LoadMessages(ctx, params.serviceID, common.LoadMessagesOpts{})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "")
	}

	if len(allMessages) == 0 || !langExists(allMessages, params.messages.Language) {
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

	// When updating original messages, changes might affect translations - transform and update all translations.
	if params.messages.Original {
		// find original messages with altered text, translate & replace text in the associated messages for all translations.
		newMessages, err := t.alterTranslations(ctx, allMessages, params.messages)
		if err != nil {
			return nil, err
		}

		// if populateMessages is true - populate missing messages for all translations.
		if params.populateTranslations {
			if newMessages, err = t.populateTranslations(ctx, newMessages, params.messages); err != nil {
				return nil, err
			}
		}

		// update all translations
		for i := range newMessages {
			if newMessages[i].Original {
				continue
			}

			err = t.repo.SaveMessages(ctx, params.serviceID, &newMessages[i])
			switch {
			default:
				// noop
			case errors.Is(err, common.ErrNotFound):
				return nil, status.Errorf(codes.NotFound, "service not found")
			case err != nil:
				return nil, status.Errorf(codes.Internal, "")
			}
		}
	}

	return messagesToProto(params.messages), nil
}

// helpers

// alterTranslations translates & replaces text in messages for all translations
// based on changes in the original messages text, returns messagesSlice containing altered translations.
func (t *TranslateServiceServer) alterTranslations(
	ctx context.Context,
	allMessages model.MessagesSlice,
	newOriginalMessages *model.Messages,
) (model.MessagesSlice, error) {
	newMessages := allMessages.Clone()

	// return if only one language is present
	if len(newMessages) == 1 {
		return newMessages, nil
	}

	originalMessages, otherMessages := newMessages.SplitOriginal()

	// return if messages don't contain original language
	if originalMessages == nil {
		return newMessages, nil
	}

	// create lookup for original messages
	originalMessagesLookup := make(map[string]*model.Message, len(originalMessages.Messages))

	for i := range originalMessages.Messages {
		originalMessagesLookup[originalMessages.Messages[i].ID] = &originalMessages.Messages[i]
	}

	// find altered original messages
	alteredMessages := model.Messages{
		Language: newOriginalMessages.Language,
	}

	for _, newMessage := range newOriginalMessages.Messages {
		if message, ok := originalMessagesLookup[newMessage.ID]; ok && message.Message != newMessage.Message {
			alteredMessages.Messages = append(alteredMessages.Messages, newMessage)
		}
	}

	if len(alteredMessages.Messages) == 0 {
		return newMessages, nil
	}

	// translate and update altered messages for all translations
	for _, messages := range otherMessages {
		translated, err := t.translator.Translate(ctx, &alteredMessages, messages.Language)
		if err != nil {
			return nil, status.Errorf(codes.Unknown, err.Error()) // TODO: For now we don't know the cause of the error.
		}

		translatedMessagesLookup := make(map[string]*model.Message, len(translated.Messages))

		for i := range translated.Messages {
			translatedMessagesLookup[translated.Messages[i].ID] = &translated.Messages[i]
		}

		for i := range messages.Messages {
			if translatedMessage, ok := translatedMessagesLookup[messages.Messages[i].ID]; ok {
				messages.Messages[i].Message = translatedMessage.Message
				messages.Messages[i].Status = model.MessageStatusFuzzy
			}
		}
	}

	return newMessages, nil
}

// populateTranslations adds any missing messages from the original messages to all translated messages,
// returns messagesSlice containing populated translations.
func (t *TranslateServiceServer) populateTranslations(
	ctx context.Context,
	allMessages model.MessagesSlice,
	newOriginalMessages *model.Messages,
) ([]model.Messages, error) {
	newMessages := allMessages.Clone()

	// Iterate over the existing messages
	for i := range newMessages {
		// Skip if Original is true
		if newMessages[i].Original {
			continue
		}

		// Create a map to store the IDs of the translated messages
		translatedMessageIDs := make(map[string]struct{}, len(newMessages[i].Messages))
		for _, m := range newMessages[i].Messages {
			translatedMessageIDs[m.ID] = struct{}{}
		}

		// Create a new messages to store the messages that need to be translated
		toBeTranslated := &model.Messages{
			Language: newMessages[i].Language,
			Messages: make([]model.Message, 0, len(newOriginalMessages.Messages)),
		}

		// Iterate over the original messages and add any missing messages to the toBeTranslated.Messages slice
		for _, m := range newOriginalMessages.Messages {
			if _, ok := translatedMessageIDs[m.ID]; !ok {
				toBeTranslated.Messages = append(toBeTranslated.Messages, m)
			}
		}

		// Translate messages -
		// untranslated text in toBeTranslated messages will be translated from original to target language.
		targetLanguage := toBeTranslated.Language
		toBeTranslated.Language = newOriginalMessages.Language
		translated, err := t.translator.Translate(ctx, toBeTranslated, targetLanguage)
		if err != nil {
			return nil, status.Errorf(codes.Unknown, err.Error()) // TODO: For now we don't know the cause of the error.
		}

		// Append the translated messages to the existing messages
		newMessages[i].Messages = append(newMessages[i].Messages, translated.Messages...)
	}

	return newMessages, nil
}

// langExists returns true if the provided language exists in the provided model.Messages slice.
func langExists(msgs []model.Messages, lang language.Tag) bool {
	return slices.ContainsFunc(msgs, func(m model.Messages) bool {
		return m.Language == lang
	})
}

package server

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/google/uuid"
	"go.expect.digital/translate/pkg/model"
	translatev1 "go.expect.digital/translate/pkg/pb/translate/v1"
	"go.expect.digital/translate/pkg/repo"
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
		repo.LoadMessagesOpts{FilterLanguages: []language.Tag{params.messages.Language}})
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
		msgs, err := t.repo.LoadMessages(ctx, params.serviceID, repo.LoadMessagesOpts{})
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

	if err := t.repo.SaveMessages(ctx, params.serviceID, params.messages); errors.Is(err, repo.ErrNotFound) {
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

	messages, err := t.repo.LoadMessages(ctx, params.serviceID, repo.LoadMessagesOpts{})
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

	all, err := t.repo.LoadMessages(ctx, params.serviceID, repo.LoadMessagesOpts{})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "")
	}

	switch all.HasLanguage(params.messages.Language) {
	case true:
		all.Replace(*params.messages)
	case false:
		return nil, status.Errorf(codes.NotFound, "no messages for language: '%s'", params.messages.Language)
	}

	// When updating original messages, changes might affect translations - transform and update all translations.
	if params.messages.Original {
		originalMessages := all[all.OriginalIndex()]

		// Mark new or altered messages as untranslated.
		all.MarkUntranslated(originalMessages.FindChangedMessageIDs(params.messages))

		// If populateMessages is true - populate missing messages for all translations.
		if params.populateTranslations {
			all.PopulateTranslations()
		}

		// Fuzzy translate untranslated messages for all translations
		if all, err = t.fuzzyTranslate(ctx, all); err != nil {
			return nil, status.Errorf(codes.Unknown, err.Error())
		}
	}

	// Update messages for all translations
	for i := range all {
		err = t.repo.SaveMessages(ctx, params.serviceID, &all[i])
		if err != nil {
			return nil, status.Errorf(codes.Internal, "")
		}
	}

	return messagesToProto(params.messages), nil
}

// helpers

// fuzzyTranslate fuzzy translates any untranslated messages,
// returns messagesSlice containing refreshed translations.
func (t *TranslateServiceServer) fuzzyTranslate(
	ctx context.Context,
	all model.MessagesSlice,
) (model.MessagesSlice, error) {
	origIdx := all.OriginalIndex()
	if origIdx == -1 {
		return nil, errors.New("original messages not found")
	}

	for i := range all {
		// Skip original messages
		if i == origIdx {
			continue
		}

		// Create a map to store pointers to untranslated messages
		untranslatedMessagesLookup := make(map[string]*model.Message)

		// Iterate over the messages and add any untranslated messages to the untranslated messages lookup
		for j := range all[i].Messages {
			if all[i].Messages[j].Status == model.MessageStatusUntranslated {
				untranslatedMessagesLookup[all[i].Messages[j].ID] = &all[i].Messages[j]
			}
		}

		// Create a new messages to store the messages that need to be translated
		toBeTranslated := &model.Messages{
			Language: all[origIdx].Language,
			Messages: make([]model.Message, 0, len(untranslatedMessagesLookup)),
		}

		for _, msg := range untranslatedMessagesLookup {
			toBeTranslated.Messages = append(toBeTranslated.Messages, *msg)
		}

		// Translate messages -
		// untranslated messages in toBeTranslated will be translated from original to target language.
		targetLanguage := all[i].Language
		translated, err := t.translator.Translate(ctx, toBeTranslated, targetLanguage)
		if err != nil {
			return nil, fmt.Errorf("translator translate messages: %w", err)
		}

		// Overwrite untranslated messages with translated messages
		for _, translatedMessage := range translated.Messages {
			if untranslatedMessage, ok := untranslatedMessagesLookup[translatedMessage.ID]; ok {
				*untranslatedMessage = translatedMessage
			}
		}
	}

	return all, nil
}

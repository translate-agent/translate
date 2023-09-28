package server

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"go.expect.digital/translate/pkg/model"
	translatev1 "go.expect.digital/translate/pkg/pb/translate/v1"
	"go.expect.digital/translate/pkg/repo"
	"golang.org/x/text/language"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ----------------------CreateTranslation-------------------------------

type createTranslationParams struct {
	messages  *model.Translation
	serviceID uuid.UUID
}

func parseCreateTranslationRequestParams(req *translatev1.CreateTranslationRequest) (*createTranslationParams, error) {
	var (
		p   = &createTranslationParams{}
		err error
	)

	if p.serviceID, err = uuidFromProto(req.ServiceId); err != nil {
		return nil, fmt.Errorf("parse service_id: %w", err)
	}

	if p.messages, err = messagesFromProto(req.Translations); err != nil {
		return nil, fmt.Errorf("parse messages: %w", err)
	}

	return p, nil
}

func (c *createTranslationParams) validate() error {
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

func (t *TranslateServiceServer) CreateTranslation(
	ctx context.Context,
	req *translatev1.CreateTranslationRequest,
) (*translatev1.Translation, error) {
	params, err := parseCreateTranslationRequestParams(req)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	if err = params.validate(); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	msgs, err := t.repo.LoadTranslation(ctx, params.serviceID,
		repo.LoadTranslationOpts{FilterLanguages: []language.Tag{params.messages.Language}})
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
		msgs, err := t.repo.LoadTranslation(ctx, params.serviceID, repo.LoadTranslationOpts{})
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

	if err := t.repo.SaveTranslation(ctx, params.serviceID, params.messages); errors.Is(err, repo.ErrNotFound) {
		return nil, status.Errorf(codes.NotFound, "service not found")
	} else if err != nil {
		return nil, status.Errorf(codes.Internal, "")
	}

	return messagesToProto(params.messages), nil
}

// ----------------------ListTranslation-------------------------------

type listTranslationParams struct {
	serviceID uuid.UUID
}

func parseListTranslationRequestParams(req *translatev1.ListTranslationRequest) (*listTranslationParams, error) {
	serviceID, err := uuidFromProto(req.GetServiceId())
	if err != nil {
		return nil, fmt.Errorf("parse service_id: %w", err)
	}

	return &listTranslationParams{serviceID: serviceID}, nil
}

func (l *listTranslationParams) validate() error {
	if l.serviceID == uuid.Nil {
		return errors.New("'service_id' is required")
	}

	return nil
}

func (t *TranslateServiceServer) ListTranslation(
	ctx context.Context,
	req *translatev1.ListTranslationRequest,
) (*translatev1.ListTranslationsResponse, error) {
	params, err := parseListTranslationRequestParams(req)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	if err = params.validate(); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	messages, err := t.repo.LoadTranslation(ctx, params.serviceID, repo.LoadTranslationOpts{})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "")
	}

	return &translatev1.ListTranslationsResponse{Translations: messagesSliceToProto(messages)}, nil
}

// ----------------------UpdateTranslation-------------------------------

type updateTranslationParams struct {
	messages             *model.Translation
	serviceID            uuid.UUID
	populateTranslations bool
}

func parseUpdateTranslationRequestParams(req *translatev1.UpdateTranslationRequest) (*updateTranslationParams, error) {
	var (
		params = updateTranslationParams{populateTranslations: req.GetPopulateTranslations()}
		err    error
	)

	if params.serviceID, err = uuidFromProto(req.ServiceId); err != nil {
		return nil, fmt.Errorf("parse service_id: %w", err)
	}

	if params.messages, err = messagesFromProto(req.Translations); err != nil {
		return nil, fmt.Errorf("parse messages: %w", err)
	}

	return &params, nil
}

func (u *updateTranslationParams) validate() error {
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

func (t *TranslateServiceServer) UpdateTranslation(
	ctx context.Context,
	req *translatev1.UpdateTranslationRequest,
) (*translatev1.Translation, error) {
	params, err := parseUpdateTranslationRequestParams(req)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	if err = params.validate(); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	all, err := t.repo.LoadTranslation(ctx, params.serviceID, repo.LoadTranslationOpts{})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "")
	}

	if !all.HasLanguage(params.messages.Language) {
		return nil, status.Errorf(codes.NotFound, "no messages for language: '%s'", params.messages.Language)
	}

	// Case for when not original, or uploading original for the first time.
	updatedTranslations := model.TranslationSlice{*params.messages}

	if origIdx := all.OriginalIndex(); params.messages.Original && origIdx != -1 {
		oldOriginal := all[origIdx]

		// Mark new or altered original messages as untranslated for all translations.
		all.MarkUntranslated(oldOriginal.FindChangedMessageIDs(params.messages))
		// Replace original messages with new ones.
		all.Replace(*params.messages)
		// Add missing messages for all translations.
		if params.populateTranslations {
			all.PopulateTranslations()
		}

		if err := t.fuzzyTranslate(ctx, all); err != nil {
			return nil, status.Errorf(codes.Internal, "")
		}

		updatedTranslations = all

	}

	// Update messages for all translations
	for i := range updatedTranslations {
		err = t.repo.SaveTranslation(ctx, params.serviceID, &all[i])
		if err != nil {
			return nil, status.Errorf(codes.Internal, "")
		}
	}

	return messagesToProto(params.messages), nil
}

// helpers

// fuzzyTranslate fuzzy translates any untranslated messages,
// returns messagesSlice containing refreshed translations.
//
// TODO: This logic should be moved to fuzzy pkg.
func (t *TranslateServiceServer) fuzzyTranslate(
	ctx context.Context,
	all model.TranslationSlice,
) error {
	origIdx := all.OriginalIndex()
	if origIdx == -1 {
		return errors.New("original messages not found")
	}

	origMsgLookup := make(map[string]string, len(all[origIdx].Messages))
	for _, msg := range all[origIdx].Messages {
		origMsgLookup[msg.ID] = msg.Message
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
				all[i].Messages[j].Message = origMsgLookup[all[i].Messages[j].ID]
				untranslatedMessagesLookup[all[i].Messages[j].ID] = &all[i].Messages[j]
			}
		}

		// Create a new messages to store the messages that need to be translated
		toBeTranslated := &model.Translation{
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
			return fmt.Errorf("translator translate messages: %w", err)
		}

		// Overwrite untranslated messages with translated messages
		for _, translatedMessage := range translated.Messages {
			if untranslatedMessage, ok := untranslatedMessagesLookup[translatedMessage.ID]; ok {
				*untranslatedMessage = translatedMessage
			}
		}
	}

	return nil
}

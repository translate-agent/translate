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
	translation *model.Translation
	serviceID   uuid.UUID
}

func parseCreateTranslationRequestParams(req *translatev1.CreateTranslationRequest) (*createTranslationParams, error) {
	var (
		p   = &createTranslationParams{}
		err error
	)

	if p.serviceID, err = uuidFromProto(req.ServiceId); err != nil {
		return nil, fmt.Errorf("parse service_id: %w", err)
	}

	if p.translation, err = translationFromProto(req.Translation); err != nil {
		return nil, fmt.Errorf("parse translation: %w", err)
	}

	return p, nil
}

func (c *createTranslationParams) validate() error {
	if c.serviceID == uuid.Nil {
		return errors.New("'service_id' is required")
	}

	if c.translation == nil {
		return errors.New("'translation' is required")
	}

	if c.translation.Language == language.Und {
		return fmt.Errorf("invalid translation: %w", errors.New("'language' is required"))
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

	translations, err := t.repo.LoadTranslations(ctx, params.serviceID,
		repo.LoadTranslationsOpts{FilterLanguages: []language.Tag{params.translation.Language}})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "")
	}

	if len(translations) > 0 {
		return nil, status.Errorf(codes.AlreadyExists, "translation already exist for language: '%s'", params.translation.Language)
	}

	// Translate messages when translation is not original and original language is known.
	if !params.translation.Original {
		// Retrieve language from original translation.
		var originalLanguage *language.Tag
		// TODO: to improve performance should be replaced with CheckTranslationExist db function.
		loadTranslations, err := t.repo.LoadTranslations(ctx, params.serviceID, repo.LoadTranslationsOpts{})
		if err != nil {
			return nil, status.Errorf(codes.Internal, "")
		}

		for _, v := range loadTranslations {
			if v.Original {
				originalLanguage = &v.Language

				// if incoming translation is empty populate with original translation.
				if params.translation.Messages == nil {
					params.translation.Messages = v.Messages
				}

				break
			}
		}

		if originalLanguage != nil {
			// Translate messages -
			// untranslated text in incoming translation will be translated from original to target language.
			targetLanguage := params.translation.Language
			params.translation.Language = *originalLanguage
			params.translation, err = t.translator.Translate(ctx, params.translation, targetLanguage)
			if err != nil {
				return nil, status.Errorf(codes.Unknown, err.Error()) // TODO: For now we don't know the cause of the error.
			}
		}
	}

	if err := t.repo.SaveTranslation(ctx, params.serviceID, params.translation); errors.Is(err, repo.ErrNotFound) {
		return nil, status.Errorf(codes.NotFound, "service not found")
	} else if err != nil {
		return nil, status.Errorf(codes.Internal, "")
	}

	return translationToProto(params.translation), nil
}

// ----------------------ListTranslations-------------------------------

type listTranslationsParams struct {
	serviceID uuid.UUID
}

func parseListTranslationsRequestParams(req *translatev1.ListTranslationsRequest) (*listTranslationsParams, error) {
	serviceID, err := uuidFromProto(req.GetServiceId())
	if err != nil {
		return nil, fmt.Errorf("parse service_id: %w", err)
	}

	return &listTranslationsParams{serviceID: serviceID}, nil
}

func (l *listTranslationsParams) validate() error {
	if l.serviceID == uuid.Nil {
		return errors.New("'service_id' is required")
	}

	return nil
}

func (t *TranslateServiceServer) ListTranslations(
	ctx context.Context,
	req *translatev1.ListTranslationsRequest,
) (*translatev1.ListTranslationsResponse, error) {
	params, err := parseListTranslationsRequestParams(req)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	if err = params.validate(); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	translations, err := t.repo.LoadTranslations(ctx, params.serviceID, repo.LoadTranslationsOpts{})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "")
	}

	return &translatev1.ListTranslationsResponse{Translations: translationsToProto(translations)}, nil
}

// ----------------------UpdateTranslation-------------------------------

type updateTranslationParams struct {
	translation          *model.Translation
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

	if params.translation, err = translationFromProto(req.Translation); err != nil {
		return nil, fmt.Errorf("parse translation: %w", err)
	}

	return &params, nil
}

func (u *updateTranslationParams) validate() error {
	if u.serviceID == uuid.Nil {
		return errors.New("'service_id' is required")
	}

	if u.translation == nil {
		return errors.New("'translation' is nil")
	}

	if u.translation.Language == language.Und {
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

	all, err := t.repo.LoadTranslations(ctx, params.serviceID, repo.LoadTranslationsOpts{})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "")
	}

	if !all.HasLanguage(params.translation.Language) {
		return nil, status.Errorf(codes.NotFound, "no translation for language: '%s'", params.translation.Language)
	}

	// Case for when not original, or uploading original for the first time.
	updatedTranslations := model.Translations{*params.translation}

	if origIdx := all.OriginalIndex(); params.translation.Original && origIdx != -1 {
		oldOriginal := all[origIdx]

		// Compare repo and request original translation.
		// Change status for new or altered translation.messages to UNTRANSLATED for all languages
		all.MarkUntranslated(oldOriginal.FindChangedMessageIDs(params.translation))
		// Replace original translation with new one.
		all.Replace(*params.translation)
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
		err = t.repo.SaveTranslation(ctx, params.serviceID, &updatedTranslations[i])
		if err != nil {
			return nil, status.Errorf(codes.Internal, "")
		}
	}

	return translationToProto(params.translation), nil
}

// helpers

// fuzzyTranslate fuzzy translates any untranslated messages,
// returns translations containing refreshed messages.
//
// TODO: This logic should be moved to fuzzy pkg.
func (t *TranslateServiceServer) fuzzyTranslate(
	ctx context.Context,
	all model.Translations,
) error {
	origIdx := all.OriginalIndex()
	if origIdx == -1 {
		return errors.New("original translation not found")
	}

	origMsgLookup := make(map[string]string, len(all[origIdx].Messages))
	for _, msg := range all[origIdx].Messages {
		origMsgLookup[msg.ID] = msg.Message
	}

	for i := range all {
		// Skip original translation
		if i == origIdx {
			continue
		}

		// Create a map to store pointers to untranslated messages
		untranslatedMessagesLookup := make(map[string]*model.Message)

		// Iterate over the messages and add any untranslated message to the untranslated messages lookup
		for j := range all[i].Messages {
			if all[i].Messages[j].Status == model.MessageStatusUntranslated {
				all[i].Messages[j].Message = origMsgLookup[all[i].Messages[j].ID]
				untranslatedMessagesLookup[all[i].Messages[j].ID] = &all[i].Messages[j]
			}
		}

		// Create a new translation to store the messages that need to be translated
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

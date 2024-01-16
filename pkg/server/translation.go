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

	if p.serviceID, err = uuidFromProto(req.GetServiceId()); err != nil {
		return nil, fmt.Errorf("parse service_id: %w", err)
	}

	if p.translation, err = translationFromProto(req.GetTranslation()); err != nil {
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

	// TODO: fail validation if message contains '{\$(0|[1-9]\d*)\}'.

	if err = params.validate(); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	all, err := t.repo.LoadTranslations(ctx, params.serviceID, repo.LoadTranslationsOpts{})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "")
	}

	if all.HasLanguage(params.translation.Language) {
		return nil,
			status.Errorf(codes.AlreadyExists, "translation already exists for language: '%s'", params.translation.Language)
	}

	switch params.translation.Original {
	case true:
		if all.OriginalIndex() != -1 {
			return nil, status.Errorf(
				codes.InvalidArgument, "original translation already exists for service: '%s'", params.serviceID)
		}
	default: // Translate messages when translation is not original and original language is known.
		if origIdx := all.OriginalIndex(); origIdx != -1 {
			targetLanguage := params.translation.Language
			params.translation.Language = all[origIdx].Language

			// if incoming translation is empty populate with original translation.
			if params.translation.Messages == nil {
				params.translation.Messages = all[origIdx].Messages
			}

			// Translate messages -
			// untranslated text in incoming translation will be translated from original to target language.
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
	mask                 model.Mask
	populateTranslations bool
	serviceID            uuid.UUID
}

func parseUpdateTranslationRequestParams(req *translatev1.UpdateTranslationRequest) (*updateTranslationParams, error) {
	var (
		params = updateTranslationParams{populateTranslations: req.GetPopulateTranslations()}
		err    error
	)

	if params.serviceID, err = uuidFromProto(req.GetServiceId()); err != nil {
		return nil, fmt.Errorf("parse service_id: %w", err)
	}

	if params.translation, err = translationFromProto(req.GetTranslation()); err != nil {
		return nil, fmt.Errorf("parse translation: %w", err)
	}

	if params.mask, err = maskFromProto(req.GetTranslation(), req.GetUpdateMask()); err != nil {
		return nil, fmt.Errorf("parse mask: %w", err)
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

// https://github.com/protocolbuffers/protobuf/blob/9bbea4aa65bdaf5fc6c2583e045c07ff37ffb0e7/src/google/protobuf/field_mask.proto#L111
//
//nolint:lll
func (t *TranslateServiceServer) UpdateTranslation(
	ctx context.Context,
	req *translatev1.UpdateTranslationRequest,
) (*translatev1.Translation, error) {
	params, err := parseUpdateTranslationRequestParams(req)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	// TODO: fail validation if message contains '{\$(0|[1-9]\d*)\}'.

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

	origIdx := all.OriginalIndex()

	switch {
	default:
		// Original translation is not affected, changes will not affect other translations - update incoming translation.
		model.Update(params.translation,
			&all[all.LanguageIndex(params.translation.Language)],
			params.mask,
		)
	case params.translation.Original && origIdx != -1:
		if params.translation.Language != all[origIdx].Language {
			return nil, status.Errorf(
				codes.InvalidArgument, "original translation already exists for service: '%s'", params.serviceID)
		}

		// Original translation is affected, changes might affect other translations - transform and update all translations.
		oldOriginal := all[origIdx]

		// Compare repo and request original translation.
		// Change status for new or altered translation.messages to UNTRANSLATED for all languages
		all.MarkUntranslated(oldOriginal.FindChangedMessageIDs(params.translation))

		// Update original translation with new one.
		model.Update(params.translation,
			&all[all.LanguageIndex(params.translation.Language)],
			params.mask,
		)

		// Add missing messages for all translations.
		if params.populateTranslations {
			all.PopulateTranslations()
		}

		if err = t.fuzzyTranslate(ctx, all); err != nil {
			return nil, status.Errorf(codes.Internal, "")
		}
	}

	// Update affected translations
	if err = t.repo.Tx(ctx, func(ctx context.Context, r repo.Repo) error {
		for i := range all {
			if err = r.SaveTranslation(ctx, params.serviceID, &all[i]); err != nil {
				return fmt.Errorf("save translation: %w", err)
			}
		}

		return nil
	}); err != nil {
		return nil, status.Errorf(codes.Internal, "")
	}

	return translationToProto(&all[all.LanguageIndex(params.translation.Language)]), nil
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

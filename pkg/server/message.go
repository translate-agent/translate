package server

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"go.expect.digital/translate/pkg/model"
	translatev1 "go.expect.digital/translate/pkg/pb/translate/v1"
	"go.expect.digital/translate/pkg/repo"
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

	all := model.MessagesSlice{*params.messages}

	// When updating original messages, changes might affect translations - transform and update all translations.
	if params.messages.Original {
		all, err = t.repo.LoadMessages(ctx, params.serviceID, repo.LoadMessagesOpts{})
		if err != nil {
			return nil, status.Errorf(codes.Internal, "")
		}

		if !all.HasLanguage(params.messages.Language) {
			return nil, status.Errorf(codes.NotFound, "no messages for language: '%s'", params.messages.Language)
		}

		all = all.Replace(*params.messages)
		prev, _ := all.SplitOriginal()

		// Find original messages with altered text, then replace text in associated messages for all translations.
		all = t.alterTranslations(all, getUntranslatedIDs(prev, params.messages))

		// If populateMessages is true - populate missing messages for all translations.
		if params.populateTranslations {
			all = t.populateTranslations(all)
		}

		// Fuzzy translate untranslated messages for all translations
		if all, err = t.fuzzyTranslate(ctx, all); err != nil {
			return nil, fmt.Errorf("fuzzy translate messages: %w", err)
		}
	} else {
		prev, err := t.repo.LoadMessages(ctx, params.serviceID, repo.LoadMessagesOpts{FilterLanguages: []language.Tag{params.messages.Language}})
		switch {
		case err != nil:
			return nil, status.Errorf(codes.Internal, "")
		case len(prev) == 0:
			return nil, status.Errorf(codes.NotFound, "messages not found")
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

// alterTranslations alter the status of specific messages in multiple languages to "untranslated"
// based on a list of message IDs, while keeping the "original" language messages unchanged.
func (t *TranslateServiceServer) alterTranslations(
	all model.MessagesSlice,
	untranslatedIDs []string,
) model.MessagesSlice {
	if len(untranslatedIDs) == 0 || len(all) == 1 {
		return all
	}

	original, others := all.SplitOriginal()

	// Return if messages don't contain original language
	if original == nil {
		return all
	}

	slices.Sort(untranslatedIDs)

	// Update altered messages for all translations
	for _, msg := range others {
		for i := range msg.Messages {
			if _, found := slices.BinarySearch(untranslatedIDs, msg.Messages[i].ID); found {
				msg.Messages[i].Status = model.MessageStatusUntranslated
			}
		}
	}

	return all
}

/*
populateTranslations adds messages that exists in original language but not in other languages.
Example:

	Original:
	{ ..., Messages: [ { ID: "1", Message: "Hello" }, { ID: "2", Message: "World" } ] }

	Translated:
	{ ..., Messages: [ { ID: "1", Message: "Bonjour" } ] }

	Result:
	{ ..., Messages: [ { ID: "1", Message: "Bonjour" }, { ID: "2", Message: "World", Status: Untranslated } ] }
*/
func (t *TranslateServiceServer) populateTranslations(all model.MessagesSlice) model.MessagesSlice {
	original, others := all.SplitOriginal()

	for i := range original.Messages {
		for j := range others {
			found := slices.ContainsFunc(others[j].Messages, func(message model.Message) bool {
				return message.ID == original.Messages[i].ID
			})

			if !found {
				newMsg := original.Messages[i]
				newMsg.Status = model.MessageStatusUntranslated

				others[j].Messages = append(others[j].Messages, newMsg)
			}
		}
	}

	return append(others, *original)
}

// fuzzyTranslate fuzzy translates any untranslated messages,
// returns messagesSlice containing refreshed translations.
func (t *TranslateServiceServer) fuzzyTranslate(
	ctx context.Context,
	all model.MessagesSlice,
) ([]model.Messages, error) {
	for i := range all {
		if all[i].Original {
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
			Language: all[i].Language,
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
			return nil, status.Errorf(codes.Unknown, err.Error()) // TODO: For now we don't know the cause of the error.
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

/*
getUntranslatedIDs returns a list of message IDs that have been altered e.g.
 1. The message.message has been changed
 2. The message with new message.ID has been added
*/
func getUntranslatedIDs(old, new *model.Messages) []string {
	lookup := make(map[string]string, len(old.Messages))

	for _, msg := range old.Messages {
		lookup[msg.ID] = msg.Message
	}

	var ids []string

	for _, msg := range new.Messages {
		if oldMsg, ok := lookup[msg.ID]; !ok || oldMsg != msg.Message {
			ids = append(ids, msg.ID)
		}
	}

	return ids
}

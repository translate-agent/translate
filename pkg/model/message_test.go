package model

import (
	"slices"
	"testing"
)

//nolint:gocognit
func Test_MarkUntranslated(t *testing.T) {
	t.Parallel()

	original := func() Translation {
		return Translation{
			Original: true,
			Messages: []Message{
				{ID: "1", Status: MessageStatusTranslated},
				{ID: "2", Status: MessageStatusTranslated},
				{ID: "3", Status: MessageStatusTranslated},
			},
		}
	}

	nonOriginal := func() Translation {
		return Translation{
			Original: false,
			Messages: []Message{
				{ID: "1", Status: MessageStatusTranslated},
				{ID: "2", Status: MessageStatusTranslated},
				{ID: "3", Status: MessageStatusTranslated},
			},
		}
	}

	tests := []struct {
		name            string
		translations    Translations
		untranslatedIDs []string
	}{
		// Nothing is changed, untranslated IDs are not provided.
		{
			name:            "Without untranslated IDs",
			translations:    Translations{original(), nonOriginal()},
			untranslatedIDs: nil,
		},
		// Nothing is changed, translation with original flag should not be altered.
		{
			name:            "One original translation",
			translations:    Translations{original()},
			untranslatedIDs: []string{"1"},
		},
		// First message status is changed to untranslated for all translations, other messages are not changed.
		{
			name:            "Multiple translations",
			translations:    Translations{nonOriginal(), nonOriginal()},
			untranslatedIDs: []string{"1"},
		},
		// First message status is changed to untranslated for all translations except original one
		// other messages are not changed.
		{
			name:            "Mixed translations",
			translations:    Translations{original(), nonOriginal()},
			untranslatedIDs: []string{"1", "2"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			origIdx := test.translations.OriginalIndex()
			test.translations.MarkUntranslated(test.untranslatedIDs)

			// For original translations, no translation.messages should be altered, e.g.
			// all messages should be with status translated.
			if origIdx != -1 {
				for _, msg := range test.translations[origIdx].Messages {
					if msg.Status != MessageStatusTranslated {
						t.Errorf("want messages status '%s', got '%s'", ptr(MessageStatusTranslated), &msg.Status)
					}
				}
			}

			// For non original translations:
			// 1. if it's ID is in untranslated IDs then it's status should be changed to untranslated.
			// 2. if it's ID is not in untranslated IDs, it's status should be left as is, e.g. translated.
			for _, translation := range test.translations {
				if translation.Original {
					continue
				}

				for _, message := range translation.Messages {
					wantStatus := MessageStatusTranslated
					if slices.Contains(test.untranslatedIDs, message.ID) {
						wantStatus = MessageStatusUntranslated
					}

					if wantStatus != message.Status {
						t.Errorf("want message status '%s', got '%s'", &wantStatus, &message.Status)
					}
				}
			}
		})
	}
}

func Test_PopulateTranslations(t *testing.T) {
	t.Parallel()

	// for test1
	onlyOriginal := Translations{
		Translation{
			Original: true,
			Messages: []Message{
				{ID: "0", Message: "0", Status: MessageStatusTranslated},
				{ID: "1", Message: "1", Status: MessageStatusTranslated},
				{ID: "2", Message: "2", Status: MessageStatusTranslated},
			},
		},
	}

	// for test2
	mixed := Translations{
		Translation{
			Original: true,
			Messages: []Message{
				{ID: "0", Message: "0", Status: MessageStatusTranslated},
				{ID: "1", Message: "1", Status: MessageStatusTranslated},
				{ID: "2", Message: "2", Status: MessageStatusTranslated},
			},
		},
		// Same messages, nothing should be populated.
		Translation{
			Original: false,
			Messages: []Message{
				{ID: "0", Message: "0", Status: MessageStatusTranslated},
				{ID: "1", Message: "1", Status: MessageStatusTranslated},
				{ID: "2", Message: "2", Status: MessageStatusTranslated},
			},
		},
		// Missing ID:2, should be added
		Translation{
			Original: false,
			Messages: []Message{
				{ID: "0", Message: "0", Status: MessageStatusTranslated},
				{ID: "1", Message: "1", Status: MessageStatusTranslated},
			},
		},
		// Empty translations, all translation.messages from original should be added.
		Translation{
			Original: false,
			Messages: []Message{},
		},
	}

	wantLen := len(onlyOriginal[0].Messages)
	wantIDs := []string{"0", "1", "2"}

	tests := []struct {
		name         string
		translations Translations
	}{
		{
			// Only original translation -> noop
			name:         "Nothing to populate",
			translations: onlyOriginal,
		},
		{
			// Original translation has extra messages -> translated messages should be populated with the extra messages.
			name:         "Populate multiple",
			translations: mixed,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			test.translations.PopulateTranslations()

			for _, translation := range test.translations {
				if len(translation.Messages) != wantLen {
					t.Errorf("want messages length %d, got %d", wantLen, len(translation.Messages))
				}

				// Check that translation has all messages from original.
				// Status check not needed, as if translated messages
				// are successfully populated, they will also have status Untranslated
				for _, message := range translation.Messages {
					if !slices.Contains(wantIDs, message.ID) {
						t.Errorf("want %v to contain %s", wantIDs, message.ID)
						return
					}

					if !slices.Contains(wantIDs, message.Message) {
						t.Errorf("want %v to contain %s", wantIDs, message.Message)
					}
				}
			}
		})
	}
}

func Test_FindChangedMessageIDs(t *testing.T) {
	t.Parallel()

	old := Translation{
		Messages: []Message{
			{ID: "1", Message: "Hello"},
			{ID: "2", Message: "World"},
		},
	}
	current := Translation{
		Messages: []Message{
			{ID: "1", Message: "Hello"},
			{ID: "2", Message: "Go"},
			{ID: "3", Message: "Testing"},
		},
	}

	changedIDs := old.FindChangedMessageIDs(&current)

	// ID:1 -> Are the same (Should not be included)
	// ID:2 -> Messages has been changed (Should be included)
	// ID:3 -> Is new (Should be included)
	if !slices.Equal([]string{"2", "3"}, changedIDs) {
		t.Errorf("want %v, got %v", []string{"2", "3"}, changedIDs)
	}
}

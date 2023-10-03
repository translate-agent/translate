package model

import (
	"slices"
	"testing"

	"github.com/stretchr/testify/require"
)

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
		untranslatedIds []string
	}{
		// Nothing is changed, untranslated IDs are not provided.
		{
			name:            "Without untranslated IDs",
			translations:    Translations{original(), nonOriginal()},
			untranslatedIds: nil,
		},
		// Nothing is changed, messages with original flag should not be altered.
		{
			name:            "One original translation",
			translations:    Translations{original()},
			untranslatedIds: []string{"1"},
		},
		// First message status is changed to untranslated for all messages, other messages are not changed.
		{
			name:            "Multiple translations",
			translations:    Translations{nonOriginal(), nonOriginal()},
			untranslatedIds: []string{"1"},
		},
		// First message status is changed to untranslated for all translations except original one
		// other messages are not changed.
		{
			name:            "Mixed translations",
			translations:    Translations{original(), nonOriginal()},
			untranslatedIds: []string{"1", "2"},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			origIdx := tt.translations.OriginalIndex()
			tt.translations.MarkUntranslated(tt.untranslatedIds)

			// For original translations, no messages should be altered, e.g. all messages should be with status translated.
			if origIdx != -1 {
				for _, msg := range tt.translations[origIdx].Messages {
					require.Equal(t, MessageStatusTranslated.String(), msg.Status.String())
				}
			}

			// For non original messages:
			// 1. if it's ID is in untranslated IDs then it's status should be changed to untranslated.
			// 2. if it's ID is not in untranslated IDs, it's status should be left as is, e.g. translated.
			for _, translation := range tt.translations {
				if translation.Original {
					continue
				}

				for _, message := range translation.Messages {
					expectedStatus := MessageStatusTranslated
					if slices.Contains(tt.untranslatedIds, message.ID) {
						expectedStatus = MessageStatusUntranslated
					}

					require.Equal(t, expectedStatus.String(), message.Status.String())
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
		// Empty messages, all messages from original should be added.
		Translation{
			Original: false,
			Messages: []Message{},
		},
	}

	expectedLen := len(onlyOriginal[0].Messages)
	expectedIds := []string{"0", "1", "2"}

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
			// Original translation have extra messages -> translated messages should be populated with the extra messages.
			name:         "Populate multiple",
			translations: mixed,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tt.translations.PopulateTranslations()

			for _, translation := range tt.translations {
				require.Len(t, translation.Messages, expectedLen)

				// Check that all translation.messages has all messages from original.
				for _, message := range translation.Messages {
					require.Contains(t, expectedIds, message.ID)
					require.Contains(t, expectedIds, message.Message)

					// Status check not needed, as if translated messages
					// are successfully populated, they will also have status Untranslated
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
	new := Translation{
		Messages: []Message{
			{ID: "1", Message: "Hello"},
			{ID: "2", Message: "Go"},
			{ID: "3", Message: "Testing"},
		},
	}

	changedIDs := old.FindChangedMessageIDs(&new)

	// ID:1 -> Are the same (Should not be included)
	// ID:2 -> Messages has been changed (Should be included)
	// ID:3 -> Is new (Should be included)
	require.Equal(t, []string{"2", "3"}, changedIDs)
}

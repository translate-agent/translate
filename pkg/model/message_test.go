package model

import (
	"slices"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_MarkUntranslated(t *testing.T) {
	t.Parallel()

	originalMsgs := func() Translation {
		return Translation{
			Original: true,
			Messages: []Message{
				{ID: "1", Status: MessageStatusTranslated},
				{ID: "2", Status: MessageStatusTranslated},
				{ID: "3", Status: MessageStatusTranslated},
			},
		}
	}

	nonOriginalMsgs := func() Translation {
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
		messagesSlice   TranslationSlice
		untranslatedIds []string
	}{
		// Nothing is changed, untranslated IDs are not provided.
		{
			name:            "Without untranslated IDs",
			messagesSlice:   TranslationSlice{originalMsgs(), nonOriginalMsgs()},
			untranslatedIds: nil,
		},
		// Nothing is changed, messages with original flag should not be altered.
		{
			name:            "One original messages",
			messagesSlice:   TranslationSlice{originalMsgs()},
			untranslatedIds: []string{"1"},
		},
		// First message status is changed to untranslated for all messages, other messages are not changed.
		{
			name:            "Multiple translated messages",
			messagesSlice:   TranslationSlice{nonOriginalMsgs(), nonOriginalMsgs()},
			untranslatedIds: []string{"1"},
		},
		// First message status is changed to untranslated for all messages except original one
		// other messages are not changed.
		{
			name:            "Mixed messages",
			messagesSlice:   TranslationSlice{originalMsgs(), nonOriginalMsgs()},
			untranslatedIds: []string{"1", "2"},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			origIdx := tt.messagesSlice.OriginalIndex()
			tt.messagesSlice.MarkUntranslated(tt.untranslatedIds)

			// For original messages, no messages should be altered, e.g. all messages should be with status translated.
			if origIdx != -1 {
				for _, msg := range tt.messagesSlice[origIdx].Messages {
					require.Equal(t, MessageStatusTranslated.String(), msg.Status.String())
				}
			}

			// For non original messages:
			// 1. if it's ID is in untranslated IDs then it's status should be changed to untranslated.
			// 2. if it's ID is not in untranslated IDs, it's status should be left as is, e.g. translated.
			for _, messages := range tt.messagesSlice {
				if messages.Original {
					continue
				}

				for _, message := range messages.Messages {
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
	onlyOriginal := TranslationSlice{
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
	mixed := TranslationSlice{
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
		messageSlice TranslationSlice
	}{
		{
			// Only original messages -> noop
			name:         "Nothing to populate",
			messageSlice: onlyOriginal,
		},
		{
			// Original messages have extra messages -> translated messages should be populated with the extra messages.
			name:         "Populate multiple",
			messageSlice: mixed,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tt.messageSlice.PopulateTranslations()

			for _, messages := range tt.messageSlice {
				require.Len(t, messages.Messages, expectedLen)

				// Check that all messages.messages has all messages from original.
				for _, message := range messages.Messages {
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

package model

import (
	"testing"

	"github.com/stretchr/testify/require"

	"golang.org/x/exp/slices"
)

func Test_MarkUntranslated(t *testing.T) {
	t.Parallel()

	originalMsgs := func() Messages {
		return Messages{
			Original: true,
			Messages: []Message{
				{ID: "1", Status: MessageStatusTranslated},
				{ID: "2", Status: MessageStatusTranslated},
				{ID: "3", Status: MessageStatusTranslated},
			},
		}
	}

	nonOriginalMsgs := func() Messages {
		return Messages{
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
		messages        MessagesSlice
		untranslatedIds []string
	}{
		// Nothing is changed, untranslated IDs are not provided.
		{
			name:            "Without untranslated IDs",
			messages:        MessagesSlice{originalMsgs(), nonOriginalMsgs()},
			untranslatedIds: nil,
		},
		// Nothing is changed, messages with original flag should not be altered.
		{
			name:            "One original messages",
			messages:        MessagesSlice{originalMsgs()},
			untranslatedIds: []string{"1"},
		},
		// First message status is changed to untranslated for all messages, other messages are not changed.
		{
			name:            "Multiple translated messages",
			messages:        MessagesSlice{nonOriginalMsgs(), nonOriginalMsgs()},
			untranslatedIds: []string{"1"},
		},
		// First message status is changed to untranslated for all messages except original one
		// other messages are not changed.
		{
			name:            "Mixed messages",
			messages:        MessagesSlice{originalMsgs(), nonOriginalMsgs()},
			untranslatedIds: []string{"1", "2"},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tt.messages.MarkUntranslated(tt.untranslatedIds)

			original, others := tt.messages.SplitOriginal()

			// For original messages, no messages should be altered, e.g. all messages should be with status translated.
			if original != nil {
				for _, msg := range original.Messages {
					require.Equal(t, MessageStatusTranslated.String(), msg.Status.String())
				}
			}

			// For non original messages:
			// 1. if it's ID is in untranslated IDs then it's status should be changed to untranslated.
			// 2. if it's ID is not in untranslated IDs, it's status should be left as is, e.g. translated.
			for _, msgs := range others {
				for _, msg := range msgs.Messages {
					expectedStatus := MessageStatusTranslated
					if slices.Contains(tt.untranslatedIds, msg.ID) {
						expectedStatus = MessageStatusUntranslated
					}

					require.Equal(t, expectedStatus.String(), msg.Status.String())
				}
			}
		})
	}
}

func Test_PopulateTranslations(t *testing.T) {
	t.Parallel()

	// for test1
	onlyOriginal := MessagesSlice{
		Messages{
			Original: true,
			Messages: []Message{
				{ID: "0", Message: "0", Status: MessageStatusTranslated},
				{ID: "1", Message: "1", Status: MessageStatusTranslated},
				{ID: "2", Message: "2", Status: MessageStatusTranslated},
			},
		},
	}

	// for test2
	mixed := MessagesSlice{
		Messages{
			Original: true,
			Messages: []Message{
				{ID: "0", Message: "0", Status: MessageStatusTranslated},
				{ID: "1", Message: "1", Status: MessageStatusTranslated},
				{ID: "2", Message: "2", Status: MessageStatusTranslated},
			},
		},
		// Same messages, nothing should be populated.
		Messages{
			Original: false,
			Messages: []Message{
				{ID: "0", Message: "0", Status: MessageStatusTranslated},
				{ID: "1", Message: "1", Status: MessageStatusTranslated},
				{ID: "2", Message: "2", Status: MessageStatusTranslated},
			},
		},
		// Missing ID:2, should be added
		Messages{
			Original: false,
			Messages: []Message{
				{ID: "0", Message: "0", Status: MessageStatusTranslated},
				{ID: "1", Message: "1", Status: MessageStatusTranslated},
			},
		},
		// Empty messages, all messages from original should be added.
		Messages{
			Original: false,
			Messages: []Message{},
		},
	}

	expectedLen := len(onlyOriginal[0].Messages)
	expectedIds := []string{"0", "1", "2"}

	tests := []struct {
		name        string
		allMessages MessagesSlice
	}{
		{
			// Only original messages -> noop
			name:        "Nothing to populate",
			allMessages: onlyOriginal,
		},
		{
			// Original messages have extra messages -> translated messages should be populated with the extra messages.
			name:        "Populate multiple",
			allMessages: mixed,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tt.allMessages.PopulateTranslations()

			for _, messages := range tt.allMessages {
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

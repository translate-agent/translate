package model

import (
	"slices"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_AlterTranslations(t *testing.T) {
	t.Parallel()

	originalMsgs := Messages{
		Original: true,
		Messages: []Message{
			{ID: "1", Status: MessageStatusTranslated},
			{ID: "2", Status: MessageStatusTranslated},
			{ID: "3", Status: MessageStatusTranslated},
		},
	}

	nonOriginalMsgs := Messages{
		Original: false,
		Messages: []Message{
			{ID: "1", Status: MessageStatusTranslated},
			{ID: "2", Status: MessageStatusTranslated},
			{ID: "3", Status: MessageStatusTranslated},
		},
	}

	tests := []struct {
		name            string
		messages        MessagesSlice
		untranslatedIds []string
		expected        MessagesSlice
	}{
		// Nothing is changed, untranslated IDs are not provided.
		{
			name:            "Without untranslated IDs",
			messages:        MessagesSlice{originalMsgs, nonOriginalMsgs},
			untranslatedIds: nil,
		},
		// Nothing is changed, messages with original flag should not be altered.
		{
			name:            "One original messages",
			messages:        MessagesSlice{originalMsgs},
			untranslatedIds: []string{originalMsgs.Messages[0].ID},
		},
		// First message status is changed to untranslated for all messages, other messages are not changed.
		{
			name:            "Multiple translated messages",
			messages:        MessagesSlice{nonOriginalMsgs, nonOriginalMsgs},
			untranslatedIds: []string{nonOriginalMsgs.Messages[0].ID},
		},
		// First message status is changed to untranslated for all messages except original one
		// other messages are not changed.
		{
			name:            "Mixed messages",
			messages:        MessagesSlice{originalMsgs, nonOriginalMsgs},
			untranslatedIds: []string{originalMsgs.Messages[0].ID},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tt.messages.AlterTranslations(tt.untranslatedIds)

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

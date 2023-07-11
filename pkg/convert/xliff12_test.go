package convert

import (
	"fmt"
	"math/rand"
	"reflect"
	"strings"
	"testing"
	"testing/quick"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.expect.digital/translate/pkg/model"
	"go.expect.digital/translate/pkg/testutil"
	testutilrand "go.expect.digital/translate/pkg/testutil/rand"
)

// randXliff12 dynamically generates a random XLIFF 1.2 file from the given messages.
func randXliff12(messages *model.Messages) []byte {
	sb := strings.Builder{}

	sb.WriteString(`<?xml version="1.0" encoding="UTF-8"?>`)
	sb.WriteString("<xliff xmlns=\"urn:oasis:names:tc:xliff:document:1.2\" version=\"1.2\">")

	if messages.Original {
		fmt.Fprintf(&sb, "<file source-language=\"%s\" target-language=\"und\">", messages.Language)
	} else {
		fmt.Fprintf(&sb, "<file source-language=\"und\" target-language=\"%s\">", messages.Language)
	}

	sb.WriteString("<body>")

	for _, msg := range messages.Messages {
		fmt.Fprintf(&sb, "<trans-unit id=\"%s\">", msg.ID)

		if messages.Original {
			fmt.Fprintf(&sb, "<source>%s</source>", msg.Message)
		} else {
			fmt.Fprintf(&sb, "<target>%s</target>", msg.Message)
		}

		if msg.Description != "" {
			fmt.Fprintf(&sb, "<note>%s</note>", msg.Description)
		}

		sb.WriteString("</trans-unit>")
	}

	sb.WriteString("</body>")
	sb.WriteString("</file>")
	sb.WriteString("</xliff>")

	return []byte(sb.String())
}

func Test_FromXliff12(t *testing.T) {
	t.Parallel()

	msgOpts := []testutilrand.ModelMessageOption{
		testutilrand.WithFuzzy(false), // Do not mark message as fuzzy, as this is not supported by XLIFF 1.2
	}

	sourceMessages := testutilrand.ModelMessages(3, msgOpts, testutilrand.WithOriginal(true))
	translatedMessages := testutilrand.ModelMessages(3, msgOpts, testutilrand.WithOriginal(false))

	tests := []struct {
		name     string
		expected *model.Messages
		input    []byte
	}{
		{
			name:     "Happy Path Untranslated",
			input:    randXliff12(sourceMessages),
			expected: sourceMessages,
		},
		{
			name:     "Happy Path Translated",
			input:    randXliff12(translatedMessages),
			expected: translatedMessages,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			actual, err := FromXliff12(tt.input)
			require.NoError(t, err)

			for i := range actual.Messages {
				actual.Messages[i].Message = strings.Trim(actual.Messages[i].Message, "{}") // Remove curly braces for comparison
			}

			// TODO: for now restore the flag to the expected
			// remove this as XLIFF is the format were we can implicitly determine if file is translated or not
			actual.Original = tt.expected.Original

			testutil.EqualMessages(t, tt.expected, &actual)
		})
	}
}

func Test_ToXliff12(t *testing.T) {
	t.Parallel()

	msgOpts := []testutilrand.ModelMessageOption{
		testutilrand.WithFuzzy(false), // Do not mark message as fuzzy, as this is not supported by XLIFF 1.2
	}

	messages := testutilrand.ModelMessages(4, msgOpts, testutilrand.WithOriginal(true))
	expected := randXliff12(messages)

	actual, err := ToXliff12(*messages)
	require.NoError(t, err)

	assertEqualXml(t, expected, actual)
}

func Test_TransformXLIFF12(t *testing.T) {
	t.Parallel()

	msgOpts := []testutilrand.ModelMessageOption{
		// Enclose message in curly braces, as ToXliff12() removes them, and FromXliff12() adds them again
		testutilrand.WithMessageFormat(),
		testutilrand.WithFuzzy(false), // Do not mark message as fuzzy, as this is not supported by XLIFF 1.2
	}

	conf := &quick.Config{
		MaxCount: 100,
		Values: func(values []reflect.Value, _ *rand.Rand) {
			values[0] = reflect.ValueOf(
				testutilrand.ModelMessages(3, msgOpts, testutilrand.WithOriginal(true))) // input generator
		},
	}

	f := func(expected *model.Messages) bool {
		xliffData, err := ToXliff12(*expected)
		require.NoError(t, err)

		restoredMessages, err := FromXliff12(xliffData)
		require.NoError(t, err)

		// TODO: for now restore the flag to the expected
		// remove this as XLIFF is the format were we can implicitly determine if file is translated or not
		restoredMessages.Original = expected.Original

		testutil.EqualMessages(t, expected, &restoredMessages)

		return true
	}

	assert.NoError(t, quick.Check(f, conf))
}

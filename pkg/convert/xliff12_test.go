package convert

import (
	"bytes"
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
	b := new(bytes.Buffer)

	b.WriteString(`<?xml version="1.0" encoding="UTF-8"?>`)
	b.WriteString("<xliff xmlns=\"urn:oasis:names:tc:xliff:document:1.2\" version=\"1.2\">")

	if messages.Original {
		fmt.Fprintf(b, "<file source-language=\"%s\" target-language=\"und\">", messages.Language)
	} else {
		fmt.Fprintf(b, "<file source-language=\"und\" target-language=\"%s\">", messages.Language)
	}

	b.WriteString("<body>")

	writeMsg := func(s string) { fmt.Fprintf(b, "<target>%s</target>", s) }
	if messages.Original {
		writeMsg = func(s string) { fmt.Fprintf(b, "<source>%s</source>", s) }
	}

	for _, msg := range messages.Messages {
		fmt.Fprintf(b, "<trans-unit id=\"%s\">", msg.ID)

		writeMsg(msg.Message)

		if msg.Description != "" {
			fmt.Fprintf(b, "<note>%s</note>", msg.Description)
		}

		for _, pos := range msg.Positions {
			b.WriteString(`<context-group purpose="location">`)

			if strings.Contains(pos, ":") {
				p := strings.Split(pos, ":")
				fmt.Fprintf(b, `<context context-type="sourcefile">%s</context>`, p[0])
				fmt.Fprintf(b, `<context context-type="linenumber">%s</context>`, p[1])
			} else {
				fmt.Fprintf(b, `<context context-type="sourcefile">%s</context>`, pos)
			}

			b.WriteString(`</context-group>`)
		}

		b.WriteString("</trans-unit>")
	}

	b.WriteString("</body>")
	b.WriteString("</file>")
	b.WriteString("</xliff>")

	return b.Bytes()
}

func Test_FromXliff12(t *testing.T) {
	t.Parallel()

	msgOpts := []testutilrand.ModelMessageOption{
		// Do not mark message as fuzzy, as this is not supported by XLIFF 1.2
		testutilrand.WithStatus(model.MessageStatusUntranslated),
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

			actual, err := FromXliff12(tt.input, false)
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
		// Do not mark message as fuzzy, as this is not supported by XLIFF 1.2
		testutilrand.WithStatus(model.MessageStatusUntranslated),
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
		// Do not mark message as fuzzy, as this is not supported by XLIFF 1.2
		testutilrand.WithStatus(model.MessageStatusUntranslated),
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

		restoredMessages, err := FromXliff12(xliffData, false)
		require.NoError(t, err)

		// TODO: for now restore the flag to the expected
		// remove this as XLIFF is the format were we can implicitly determine if file is translated or not
		restoredMessages.Original = expected.Original

		testutil.EqualMessages(t, expected, &restoredMessages)

		return true
	}

	assert.NoError(t, quick.Check(f, conf))
}

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

// TODO: XLIFF1.2 and XLIFF2.0 uses same test data and same tests, so we can merge them into one test file

// randXliff12 dynamically generates a random XLIFF 1.2 file from the given translations.
func randXliff12(translations *model.Translation) []byte {
	b := new(bytes.Buffer)

	b.WriteString(`<?xml version="1.0" encoding="UTF-8"?>`)
	b.WriteString("<xliff xmlns=\"urn:oasis:names:tc:xliff:document:1.2\" version=\"1.2\">")

	if translations.Original {
		fmt.Fprintf(b, "<file source-language=\"%s\" target-language=\"und\">", translations.Language)
	} else {
		fmt.Fprintf(b, "<file source-language=\"und\" target-language=\"%s\">", translations.Language)
	}

	b.WriteString("<body>")

	writeMsg := func(s string) { fmt.Fprintf(b, "<target>%s</target>", s) }
	if translations.Original {
		writeMsg = func(s string) { fmt.Fprintf(b, "<source>%s</source>", s) }
	}

	for _, msg := range translations.Messages {
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

	originalTranslations := testutilrand.ModelTranslation(
		3,
		[]testutilrand.ModelMessageOption{testutilrand.WithStatus(model.MessageStatusTranslated)},
		testutilrand.WithOriginal(true),
	)

	nonOriginalTranslations := testutilrand.ModelTranslation(
		3,
		[]testutilrand.ModelMessageOption{testutilrand.WithStatus(model.MessageStatusUntranslated)},
		testutilrand.WithOriginal(false),
	)

	tests := []struct {
		name     string
		expected *model.Translation
		data     []byte
	}{
		{
			name:     "Original",
			data:     randXliff12(originalTranslations),
			expected: originalTranslations,
		},
		{
			name:     "Different language",
			data:     randXliff12(nonOriginalTranslations),
			expected: nonOriginalTranslations,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			actual, err := FromXliff12(tt.data, tt.expected.Original)
			require.NoError(t, err)

			for i := range actual.Messages {
				actual.Messages[i].Message = strings.Trim(actual.Messages[i].Message, "{}") // Remove curly braces for comparison
			}

			testutil.EqualTranslations(t, tt.expected, &actual)
		})
	}
}

func Test_ToXliff12(t *testing.T) {
	t.Parallel()

	msgOpts := []testutilrand.ModelMessageOption{
		// Do not mark message as fuzzy, as this is not supported by XLIFF 1.2
		testutilrand.WithStatus(model.MessageStatusUntranslated),
	}

	translation := testutilrand.ModelTranslation(4, msgOpts, testutilrand.WithOriginal(true))
	expected := randXliff12(translation)

	actual, err := ToXliff12(*translation)
	require.NoError(t, err)

	assertEqualXml(t, expected, actual)
}

func Test_TransformXLIFF12(t *testing.T) {
	t.Parallel()

	msgOpts := []testutilrand.ModelMessageOption{
		// Enclose message in curly braces, as ToXliff2() removes them, and FromXliff2() adds them again
		testutilrand.WithMessageFormat(),
		testutilrand.WithStatus(model.MessageStatusTranslated),
	}

	conf := &quick.Config{
		MaxCount: 100,
		Values: func(values []reflect.Value, _ *rand.Rand) {
			values[0] = reflect.ValueOf(
				testutilrand.ModelTranslation(3, msgOpts, testutilrand.WithOriginal(true))) // input generator
		},
	}

	f := func(expected *model.Translation) bool {
		serialized, err := ToXliff12(*expected)
		require.NoError(t, err)

		parsed, err := FromXliff12(serialized, expected.Original)
		require.NoError(t, err)

		testutil.EqualTranslations(t, expected, &parsed)

		return true
	}

	assert.NoError(t, quick.Check(f, conf))
}

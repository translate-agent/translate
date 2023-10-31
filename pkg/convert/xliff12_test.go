package convert

import (
	"bytes"
	"fmt"
	"math/rand"
	"reflect"
	"strings"
	"testing"
	"testing/quick"

	"github.com/stretchr/testify/require"
	"go.expect.digital/translate/pkg/model"
	"go.expect.digital/translate/pkg/testutil"
	testutilrand "go.expect.digital/translate/pkg/testutil/rand"
	"golang.org/x/text/language"
)

// TODO: XLIFF1.2 and XLIFF2.0 uses same test data and same tests, so we can merge them into one test file

// randXliff12 dynamically generates a random XLIFF 1.2 file from the given translation.
func randXliff12(translation *model.Translation) []byte {
	b := new(bytes.Buffer)

	b.WriteString(`<?xml version="1.0" encoding="UTF-8"?>`)
	b.WriteString("<xliff xmlns=\"urn:oasis:names:tc:xliff:document:1.2\" version=\"1.2\">")

	if translation.Original {
		fmt.Fprintf(b, "<file source-language=\"%s\" target-language=\"und\">", translation.Language)
	} else {
		fmt.Fprintf(b, "<file source-language=\"und\" target-language=\"%s\">", translation.Language)
	}

	b.WriteString("<body>")

	writeMsg := func(s string) { fmt.Fprintf(b, "<target>%s</target>", s[1:len(s)-1]) }
	if translation.Original {
		writeMsg = func(s string) { fmt.Fprintf(b, "<source>%s</source>", s[1:len(s)-1]) }
	}

	for _, msg := range translation.Messages {
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

	originalTranslation := testutilrand.ModelTranslation(
		3,
		[]testutilrand.ModelMessageOption{testutilrand.WithStatus(model.MessageStatusTranslated)},
		testutilrand.WithOriginal(true),
	)

	nonOriginalTranslation := testutilrand.ModelTranslation(
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
			data:     randXliff12(originalTranslation),
			expected: originalTranslation,
		},
		{
			name:     "Different language",
			data:     randXliff12(nonOriginalTranslation),
			expected: nonOriginalTranslation,
		},
		{
			name: "Message with special chars {}",
			data: randXliff12(
				&model.Translation{
					Language: language.English,
					Original: false,
					Messages: []model.Message{
						{
							ID:      "order canceled",
							Message: `{Order #{Id} has been canceled for {ClientName} | \}`,
						},
					},
				},
			),
			expected: &model.Translation{
				Original: false,
				Language: language.English,
				Messages: []model.Message{
					{
						ID:      "order canceled",
						Message: `{Order #\{Id\} has been canceled for \{ClientName\} \| \\}`,
						Status:  model.MessageStatusUntranslated,
					},
				},
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			actual, err := FromXliff12(tt.data, tt.expected.Original)
			require.NoError(t, err)

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

	tests := []struct {
		name     string
		data     *model.Translation
		expected []byte
	}{
		{
			name:     "valid input",
			data:     translation,
			expected: randXliff12(translation),
		},
		{
			name: "message with special chars",
			data: &model.Translation{
				Original: true,
				Language: language.English,
				Messages: []model.Message{
					{
						ID:      "common.welcome",
						Message: `{User #\{ID\} \| \\}`,
					},
				},
			},
			expected: []byte(`<?xml version="1.0" encoding="UTF-8"?>
<xliff xmlns="urn:oasis:names:tc:xliff:document:1.2" version="1.2">
  <file source-language="en" target-language="und">
    <body>
      <trans-unit id="common.welcome">
        <source>User #{ID} | \</source>
      </trans-unit>
    </body>
  </file>
</xliff>`),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			actual, err := ToXliff12(*tt.data)
			require.NoError(t, err)

			assertEqualXML(t, tt.expected, actual)
		})
	}
}

func Test_TransformXLIFF12(t *testing.T) {
	t.Parallel()

	msgOpts := []testutilrand.ModelMessageOption{
		// Enclose message in curly braces, as ToXliff2() removes them, and FromXliff2() adds them again
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

	require.NoError(t, quick.Check(f, conf))
}

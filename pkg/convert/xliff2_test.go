package convert

import (
	"encoding/xml"
	"fmt"
	"math/rand"
	"reflect"
	"regexp"
	"testing"
	"testing/quick"

	"golang.org/x/text/language"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.expect.digital/translate/pkg/model"
	"go.expect.digital/translate/pkg/testutil"
	testutilrand "go.expect.digital/translate/pkg/testutil/rand"
)

func randXliff2(translation *model.Translation) []byte {
	xliff := xliff2{
		Version: "2.0",
	}

	if translation.Original {
		xliff.SrcLang = translation.Language
		xliff.TrgLang = language.Und
	} else {
		xliff.SrcLang = language.Und
		xliff.TrgLang = translation.Language
	}

	for _, msg := range translation.Messages {
		xmlMsg := unit{
			ID:     msg.ID,
			Source: "",
			Target: "",
		}

		if translation.Original {
			xmlMsg.Source = msg.Message
		} else {
			xmlMsg.Target = msg.Message
		}

		if msg.Description != "" || len(msg.Positions) > 0 {
			notes := make([]note, 0)

			for _, pos := range msg.Positions {
				notes = append(notes, note{Category: "location", Content: pos})
			}

			if msg.Description != "" {
				notes = append(notes, note{Category: "description", Content: msg.Description})
			}

			xmlMsg.Notes = &notes
		}

		xliff.File.Units = append(xliff.File.Units, xmlMsg)
	}

	xmlData, err := xml.MarshalIndent(xliff, "", "  ")
	if err != nil {
		fmt.Printf("marshaling XLIFF2.0: %v\n", err)
		return nil
	}

	xmlWithDeclaration := []byte(xml.Header + string(xmlData))

	return xmlWithDeclaration
}

func assertEqualXML(t *testing.T, expected, actual []byte) bool { //nolint:unparam
	t.Helper()
	// Matches a substring that starts with > and ends with < with zero or more whitespace in between.
	re := regexp.MustCompile(`>(\s*)<`)
	expectedTrimmed := re.ReplaceAllString(string(expected), "><")
	actualTrimmed := re.ReplaceAllString(string(actual), "><")

	return assert.Equal(t, expectedTrimmed, actualTrimmed)
}

func Test_FromXliff2(t *testing.T) {
	t.Parallel()

	originalTranslation := testutilrand.ModelTranslation(
		3,
		[]testutilrand.ModelMessageOption{testutilrand.WithStatus(model.MessageStatusTranslated)},
		testutilrand.WithOriginal(true), testutilrand.WithSimpleMF2Messages(),
	)

	nonOriginalTranslation := testutilrand.ModelTranslation(
		3,
		[]testutilrand.ModelMessageOption{testutilrand.WithStatus(model.MessageStatusUntranslated)},
		testutilrand.WithOriginal(false), testutilrand.WithSimpleMF2Messages(),
	)

	tests := []struct {
		name     string
		expected *model.Translation
		data     []byte
	}{
		{
			name:     "Original",
			data:     randXliff2(originalTranslation),
			expected: originalTranslation,
		},
		{
			name:     "Different language",
			data:     randXliff2(nonOriginalTranslation),
			expected: nonOriginalTranslation,
		},
		{
			name: "Message with special chars",
			data: randXliff2(
				&model.Translation{
					Language: language.English,
					Original: false,
					Messages: []model.Message{
						{
							ID:      "order canceled",
							Message: `Order #{Id} has been canceled for {ClientName} | \`,
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
						Message: `Order #\{Id\} has been canceled for \{ClientName\} | \\`,
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

			actual, err := FromXliff2(tt.data, &tt.expected.Original)
			require.NoError(t, err)

			testutil.EqualTranslations(t, tt.expected, &actual)
		})
	}
}

func Test_ToXliff2(t *testing.T) {
	t.Parallel()

	msgOpts := []testutilrand.ModelMessageOption{
		// Do not mark message as fuzzy, as this is not supported by XLIFF 2.0
		testutilrand.WithStatus(model.MessageStatusUntranslated),
	}

	translation := testutilrand.ModelTranslation(4,
		msgOpts,
		testutilrand.WithOriginal(true),
		testutilrand.WithSimpleMF2Messages())

	tests := []struct {
		name     string
		data     *model.Translation
		expected []byte
	}{
		{
			name:     "valid input",
			data:     translation,
			expected: randXliff2(translation),
		},
		{
			name: "message with special chars",
			data: &model.Translation{
				Original: true,
				Language: language.English,
				Messages: []model.Message{
					{
						ID:      "common.welcome",
						Message: `User #\{ID\} | \\`,
					},
				},
			},
			expected: []byte(`<?xml version="1.0" encoding="UTF-8"?>
<xliff xmlns="urn:oasis:names:tc:xliff:document:2.0" version="2.0" srcLang="en" trgLang="und">
  <file>
    <unit id="common.welcome">
      <segment>
        <source>User #{ID} | \</source>
      </segment>
    </unit>
  </file>
</xliff>`),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			actual, err := ToXliff2(*tt.data)
			require.NoError(t, err)

			assertEqualXML(t, tt.expected, actual)
		})
	}
}

func Test_TransformXLIFF2(t *testing.T) {
	t.Parallel()

	msgOpts := []testutilrand.ModelMessageOption{
		// Enclose message in curly braces, as ToXliff2() removes them, and FromXliff2() adds them again
		testutilrand.WithStatus(model.MessageStatusTranslated),
	}

	conf := &quick.Config{
		MaxCount: 100,
		Values: func(values []reflect.Value, _ *rand.Rand) {
			values[0] = reflect.ValueOf(
				testutilrand.ModelTranslation(3,
					msgOpts,
					testutilrand.WithOriginal(true),
					testutilrand.WithSimpleMF2Messages())) // input generator
		},
	}

	f := func(expected *model.Translation) bool {
		serialized, err := ToXliff2(*expected)
		require.NoError(t, err)

		parsed, err := FromXliff2(serialized, &expected.Original)
		require.NoError(t, err)

		testutil.EqualTranslations(t, expected, &parsed)

		return true
	}

	require.NoError(t, quick.Check(f, conf))
}

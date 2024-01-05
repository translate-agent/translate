package convert

import (
	"bytes"
	"fmt"
	"math/rand"
	"reflect"
	"regexp"
	"strings"
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
	b := new(bytes.Buffer)

	b.WriteString(`<?xml version="1.0" encoding="UTF-8"?>`)

	if translation.Original {
		fmt.Fprintf(
			b,
			"<xliff xmlns=\"urn:oasis:names:tc:xliff:document:2.0\" version=\"2.0\" srcLang=\"%s\" trgLang=\"und\">",
			translation.Language)
	} else {
		fmt.Fprintf(
			b,
			"<xliff xmlns=\"urn:oasis:names:tc:xliff:document:2.0\" version=\"2.0\" srcLang=\"und\" trgLang=\"%s\">",
			translation.Language)
	}

	b.WriteString("<file>")

	writeMsg := func(s string) { fmt.Fprintf(b, "<segment><target>%s</target></segment>", s) }
	if translation.Original {
		writeMsg = func(s string) { fmt.Fprintf(b, "<segment><source>%s</source></segment>", s) }
	}

	for _, msg := range translation.Messages {
		fmt.Fprintf(b, "<unit id=\"%s\">", msg.ID)

		if msg.Description != "" || len(msg.Positions) > 0 {
			fmt.Fprintf(b, "<notes>")

			for _, pos := range msg.Positions {
				fmt.Fprintf(b, "<note category=\"location\">%s</note>", pos)
			}

			if msg.Description != "" {
				fmt.Fprintf(b, "<note category=\"description\">%s</note>", msg.Description)
			}

			fmt.Fprintf(b, "</notes>")
		}

		// TODO: temporary fix, remove after #205-MF2-Complex-Messages-Support is resolved.
		msg.Message = strings.TrimPrefix(msg.Message, "{")
		msg.Message = strings.TrimSuffix(msg.Message, "}")

		writeMsg(msg.Message)

		b.WriteString("</unit>")
	}

	b.WriteString("</file>")
	b.WriteString("</xliff>")

	return b.Bytes()
}

func assertEqualXML(t *testing.T, expected, actual []byte) bool { //nolint:unparam
	t.Helper()
	// Matches a substring that starts with > and ends with < with zero or more whitespace in between.
	re := regexp.MustCompile(`>(\s*)<`)
	expectedTrimmed := re.ReplaceAllString(string(expected), "><")
	actualTrimmed := re.ReplaceAllString(string(actual), "><")

	return assert.Equal(t, expectedTrimmed, actualTrimmed)
}

func Test_FromXliff2_Random(t *testing.T) {
	t.Parallel()

	t.Skip() // TODO

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

	// TODO: temporary fix, remove after #205-MF2-Complex-Messages-Support is resolved.
	for i := range originalTranslation.Messages {
		originalTranslation.Messages[i].Message = strings.TrimPrefix(originalTranslation.Messages[i].Message, "{")
		originalTranslation.Messages[i].Message = strings.TrimSuffix(originalTranslation.Messages[i].Message, "}")
	}

	for i := range nonOriginalTranslation.Messages {
		nonOriginalTranslation.Messages[i].Message = strings.TrimPrefix(nonOriginalTranslation.Messages[i].Message, "{")
		nonOriginalTranslation.Messages[i].Message = strings.TrimSuffix(nonOriginalTranslation.Messages[i].Message, "}")
	}

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
						Message: `Order \#\{Id\} has been canceled for \{ClientName\} \| \\`,
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

// Test_FromXliff2_Default tests Xliff 2.0 default 'acc. to specification' format.
func Test_FromXliff2_Default(t *testing.T) {
	t.Parallel()

	//nolint:lll
	tests := []struct {
		expected    *model.Translation
		expectedErr error
		name        string
		data        []byte
	}{
		{
			name: "original, source content with placeholders",
			data: []byte(`<?xml version="1.0" encoding="UTF-8" ?>
			<xliff version="2.0"
				xmlns="urn:oasis:names:tc:xliff:document:2.0" srcLang="en">
				<file>
					<unit id="1">
						<originalData>
							<data id="d1">%d</data>
							<data id="d2">&lt;br/></data>
						</originalData>
						<segment>
							<source>Entries: <ph id="1" dataRef="d1" canCopy="no" canDelete="no" canOverlap="yes"/>!` +
				`<ph id="2" dataRef="d2" canCopy="no" canDelete="no" canOverlap="yes"/>(Filtered)</source>
					</segment>
				</unit>
			</file>
			</xliff>`),
			expected: &model.Translation{
				Original: true,
				Language: language.English,
				Messages: []model.Message{
					{
						ID: "1",
						Message: "Entries: {:Placeholder format=printf type=int value=%d id=1 dataRef=d1 canCopy=no canDelete=no canOverlap=yes}\\!" +
							"{:Placeholder format=misc value=<br/> id=2 dataRef=d2 canCopy=no canDelete=no canOverlap=yes}(Filtered)",
						Status: model.MessageStatusTranslated,
					},
				},
			},
		},
		{
			name: "translation, target content with placeholders, placeholder specifiers not provided",
			data: []byte(`<?xml version="1.0" encoding="UTF-8" ?>
			<xliff version="2.0"
				xmlns="urn:oasis:names:tc:xliff:document:2.0" srcLang ="en" trgLang="en">
				<file>
					<unit id="1">
						<segment>
							<target>Entries: <ph id="1" canCopy="no" canDelete="no" canOverlap="yes"/>!` +
				`<ph id="2" canCopy="no" canDelete="no" canOverlap="yes"/>(Filtered)</target>
					</segment>
				</unit>
			</file>
			</xliff>`),
			expected: &model.Translation{
				Original: false,
				Language: language.English,
				Messages: []model.Message{
					{
						ID: "1",
						Message: "Entries: {:Placeholder format=misc id=1 canCopy=no canDelete=no canOverlap=yes}\\!" +
							"{:Placeholder format=misc id=2 canCopy=no canDelete=no canOverlap=yes}(Filtered)",
						Status: model.MessageStatusUntranslated,
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

	t.Skip() // TODO

	msgOpts := []testutilrand.ModelMessageOption{
		// Do not mark message as fuzzy, as this is not supported by XLIFF 2.0
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
						Message: `{User #\{ID\} \| \\}`,
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
	t.Skip() // TODO
	t.Parallel()

	msgOpts := []testutilrand.ModelMessageOption{
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
		serialized, err := ToXliff2(*expected)
		require.NoError(t, err)

		parsed, err := FromXliff2(serialized, &expected.Original)
		require.NoError(t, err)

		// TODO: temporary fix, remove after #205-MF2-Complex-Messages-Support is resolved.
		for i := range expected.Messages {
			expected.Messages[i].Message = strings.TrimPrefix(expected.Messages[i].Message, "{")
			expected.Messages[i].Message = strings.TrimSuffix(expected.Messages[i].Message, "}")
		}

		testutil.EqualTranslations(t, expected, &parsed)

		return true
	}

	require.NoError(t, quick.Check(f, conf))
}

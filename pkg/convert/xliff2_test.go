package convert

import (
	"encoding/xml"
	"math/rand"
	"reflect"
	"regexp"
	"testing"
	"testing/quick"

	"golang.org/x/text/language"

	"go.expect.digital/translate/pkg/model"
	"go.expect.digital/translate/pkg/testutil/expect"
	testutilrand "go.expect.digital/translate/pkg/testutil/rand"
)

func randXliff2(t *testing.T, translation *model.Translation) []byte {
	xliff := xliff2{
		Version: "2.0",
	}

	if translation.Original {
		xliff.SrcLang = translation.Language
	} else {
		xliff.TrgLang = translation.Language
	}

	for _, msg := range translation.Messages {
		xmlMsg := unit{
			ID: msg.ID,
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

	xmlData, err := xml.Marshal(xliff)
	if err != nil {
		t.Error(err)
		return nil
	}

	return append([]byte(xml.Header), xmlData...)
}

func assertEqualXML(t *testing.T, want, got []byte) bool { //nolint:unparam
	t.Helper()
	// Matches a substring that starts with > and ends with < with zero or more whitespace in between.
	re := regexp.MustCompile(`>(\s*)<`)
	wantTrimmed := re.ReplaceAllString(string(want), "><")
	gotTrimmed := re.ReplaceAllString(string(got), "><")

	if wantTrimmed != gotTrimmed {
		t.Errorf("want '%s', got '%s'", wantTrimmed, gotTrimmed)
		return false
	}

	return true
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
		name string
		want *model.Translation
		data []byte
	}{
		{
			name: "Original",
			data: randXliff2(t, originalTranslation),
			want: originalTranslation,
		},
		{
			name: "Different language",
			data: randXliff2(t, nonOriginalTranslation),
			want: nonOriginalTranslation,
		},
		{
			name: "Message with special chars",
			data: randXliff2(t,
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
			want: &model.Translation{
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

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			got, err := FromXliff2(test.data, &test.want.Original)
			if err != nil {
				t.Error(err)
				return
			}

			if !reflect.DeepEqual(*test.want, got) {
				t.Errorf("\nwant %v\ngot  %v", test.want, got)
			}
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
		name string
		data *model.Translation
		want []byte
	}{
		{
			name: "valid input",
			data: translation,
			want: randXliff2(t, translation),
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
			want: []byte(`<?xml version="1.0" encoding="UTF-8"?>
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

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			got, err := ToXliff2(*test.data)
			if err != nil {
				t.Error(err)
				return
			}

			assertEqualXML(t, test.want, got)
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

	f := func(want *model.Translation) bool {
		serialized, err := ToXliff2(*want)
		if err != nil {
			t.Error(err)
			return false
		}

		parsed, err := FromXliff2(serialized, &want.Original)
		if err != nil {
			t.Error(err)
			return false
		}

		if !reflect.DeepEqual(*want, parsed) {
			t.Errorf("\nwant %v\ngot  %v", want, parsed)
		}

		return true
	}

	expect.NoError(t, quick.Check(f, conf))
}

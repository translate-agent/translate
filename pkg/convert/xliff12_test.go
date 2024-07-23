package convert

import (
	"encoding/xml"
	"math/rand"
	"reflect"
	"strings"
	"testing"
	"testing/quick"

	"go.expect.digital/translate/pkg/model"
	"go.expect.digital/translate/pkg/testutil/expect"
	testutilrand "go.expect.digital/translate/pkg/testutil/rand"
	"golang.org/x/text/language"
)

// TODO: XLIFF1.2 and XLIFF2.0 uses same test data and same tests, so we can merge them into one test file

// randXliff12 dynamically generates a random XLIFF 1.2 file from the given translation.
func randXliff12(t *testing.T, translation *model.Translation) []byte {
	xliff := xliff12{
		Version: "1.2",
	}

	if translation.Original {
		xliff.File.SourceLanguage = translation.Language
	} else {
		xliff.File.TargetLanguage = translation.Language
	}

	for _, msg := range translation.Messages {
		xmlMsg := transUnit{
			ID:   msg.ID,
			Note: msg.Description,
		}

		if translation.Original {
			xmlMsg.Source = msg.Message
		} else {
			xmlMsg.Target = msg.Message
		}

		for _, pos := range msg.Positions {
			if strings.Contains(pos, ":") {
				p := strings.Split(pos, ":")
				xmlMsg.ContextGroups = append(xmlMsg.ContextGroups, contextGroup{
					Purpose: "location",
					Contexts: []context{
						{Type: "sourcefile", Content: p[0]},
						{Type: "linenumber", Content: p[1]},
					},
				})
			} else {
				xmlMsg.ContextGroups = append(xmlMsg.ContextGroups, contextGroup{
					Purpose: "location",
					Contexts: []context{
						{Type: "sourcefile", Content: pos},
					},
				})
			}
		}

		xliff.File.Body.TransUnits = append(xliff.File.Body.TransUnits, xmlMsg)
	}

	xmlData, err := xml.Marshal(xliff)
	if err != nil {
		t.Error(err)
		return nil
	}

	return append([]byte(xml.Header), xmlData...)
}

func Test_FromXliff12(t *testing.T) {
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
			data: randXliff12(t, originalTranslation),
			want: originalTranslation,
		},
		{
			name: "Different language",
			data: randXliff12(t, nonOriginalTranslation),
			want: nonOriginalTranslation,
		},
		{
			name: "Message with special chars {}",
			data: randXliff12(t,
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

			got, err := FromXliff12(test.data, &test.want.Original)
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

func Test_ToXliff12(t *testing.T) {
	t.Parallel()

	msgOpts := []testutilrand.ModelMessageOption{
		// Do not mark message as fuzzy, as this is not supported by XLIFF 1.2
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
			want: randXliff12(t, translation),
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

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			got, err := ToXliff12(*test.data)
			if err != nil {
				t.Error(err)
				return
			}

			assertEqualXML(t, test.want, got)
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
				testutilrand.ModelTranslation(3,
					msgOpts,
					testutilrand.WithOriginal(true),
					testutilrand.WithSimpleMF2Messages())) // input generator
		},
	}

	f := func(want *model.Translation) bool {
		serialized, err := ToXliff12(*want)
		if err != nil {
			t.Error(err)
			return false
		}

		parsed, err := FromXliff12(serialized, &want.Original)
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

package convert

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.expect.digital/translate/pkg/model"
	"golang.org/x/text/language"
)

func assertEqualXml(t *testing.T, expected, actual []byte) bool { //nolint:unparam
	t.Helper()
	// Matches a substring that starts with > and ends with < with zero or more whitespace in between.
	re := regexp.MustCompile(`>(\s*)<`)
	expectedTrimmed := re.ReplaceAllString(string(expected), "><")
	actualTrimmed := re.ReplaceAllString(string(actual), "><")

	return assert.Equal(t, expectedTrimmed, actualTrimmed)
}

func TestFromXliff2(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		wantErr error
		data    []byte
		want    model.Messages
	}{
		{
			name: "All OK",
			data: []byte(`<?xml version="1.0" encoding="UTF-8"?>
<xliff version="2.0" xmlns="urn:oasis:names:tc:xliff:document:2.0" srcLang="en" trgLang="fr">
  <file id="ngi18n" original="ng.template">
    <unit id="common.welcome">
      <notes>
        <note category="location">src/app/app.component.html:16</note>
      </notes>
      <segment>
        <source>Welcome!</source>
        <target>Bienvenue!</target>
      </segment>
    </unit>
    <unit id="common.app.title">
      <notes>
        <note category="location">src/app/app.component.html:4</note>
        <note category="description">App title</note>
      </notes>
      <segment>
        <source>Diary</source>
        <target>Agenda</target>
      </segment>
    </unit>
  </file>
</xliff>`),
			want: model.Messages{
				Language: language.English,
				Messages: []model.Message{
					{
						ID:      "common.welcome",
						Message: "Welcome!",
					},
					{
						ID:          "common.app.title",
						Message:     "Diary",
						Description: "App title",
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "Malformed language tag",
			data: []byte(`<?xml version="1.0" encoding="UTF-8"?>
<xliff version="2.0" xmlns="urn:oasis:names:tc:xliff:document:2.0" srcLang="xyz-ZY-Latn" trgLang="fr">
  <file id="ngi18n" original="ng.template">
    <unit id="common.welcome">
      <notes>
        <note category="location">src/app/app.component.html:16</note>
      </notes>
      <segment>
        <source>Welcome!</source>
        <target>Bienvenue!</target>
      </segment>
    </unit>
    <unit id="common.app.title">
      <notes>
        <note category="location">src/app/app.component.html:4</note>
        <note category="description">App title</note>
      </notes>
      <segment>
        <source>Diary</source>
        <target>Agenda</target>
      </segment>
    </unit>
  </file>
</xliff>`),
			wantErr: fmt.Errorf("language: subtag \"xyz\" is well-formed but unknown"),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result, err := FromXliff2(tt.data)
			if tt.wantErr != nil {
				assert.ErrorContains(t, err, tt.wantErr.Error())
				return
			}

			if !assert.NoError(t, err) {
				return
			}

			assert.Equal(t, tt.want.Language, result.Language)
			assert.ElementsMatch(t, tt.want.Messages, result.Messages)
		})
	}
}

func Test_ToXliff2(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		want     []byte
		wantErr  error
		messages model.Messages
	}{
		{
			name: "All OK",
			want: []byte(`<?xml version="1.0" encoding="UTF-8"?>
<xliff xmlns="urn:oasis:names:tc:xliff:document:2.0" version="2.0" srcLang="en">
  <file>
    <unit id="Welcome">
      <notes>
        <note category="description">To welcome a new visitor</note>
      </notes>
      <segment>
        <source>Welcome to our website!</source>
      </segment>
    </unit>
    <unit id="Error">
      <notes>
        <note category="description">To inform the user of an error</note>
      </notes>
      <segment>
        <source>Something went wrong. Please try again later.</source>
      </segment>
    </unit>
    <unit id="Feedback">
      <segment>
        <source>We appreciate your feedback. Thank you for using our service.</source>
      </segment>
    </unit>
  </file>
</xliff>`),
			wantErr: nil,
			messages: model.Messages{
				Language: language.English,
				Messages: []model.Message{
					{
						ID:          "Welcome",
						Message:     "Welcome to our website!",
						Description: "To welcome a new visitor",
					},
					{
						ID:          "Error",
						Message:     "Something went wrong. Please try again later.",
						Description: "To inform the user of an error",
					},
					{
						ID:      "Feedback",
						Message: "We appreciate your feedback. Thank you for using our service.",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result, err := ToXliff2(tt.messages)

			if tt.wantErr != nil {
				assert.ErrorContains(t, err, tt.wantErr.Error())
				return
			}

			if !assert.NoError(t, err) {
				return
			}

			assertEqualXml(t, tt.want, result)
		})
	}
}

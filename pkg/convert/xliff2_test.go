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
		name        string
		expectedErr error
		input       []byte
		expected    model.Messages
	}{
		{
			name: "All OK",
			input: []byte(`<?xml version="1.0" encoding="UTF-8"?>
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
			expected: model.Messages{
				Language: language.English,
				Messages: []model.Message{
					{
						ID:      "common.welcome",
						Message: "{Welcome!}",
					},
					{
						ID:          "common.app.title",
						Message:     "{Diary}",
						Description: "App title",
					},
				},
			},
			expectedErr: nil,
		},
		{
			name: "Malformed language",
			input: []byte(`<?xml version="1.0" encoding="UTF-8"?>
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
			expectedErr: fmt.Errorf("language: subtag \"xyz\" is well-formed but unknown"),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			actual, err := FromXliff2(tt.input)
			if tt.expectedErr != nil {
				assert.ErrorContains(t, err, tt.expectedErr.Error())
				return
			}

			if !assert.NoError(t, err) {
				return
			}

			assert.Equal(t, tt.expected.Language, actual.Language)
			assert.ElementsMatch(t, tt.expected.Messages, actual.Messages)
		})
	}
}

func Test_ToXliff2(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		expected    []byte
		expectedErr error
		input       model.Messages
	}{
		{
			name: "All OK",
			expected: []byte(`<?xml version="1.0" encoding="UTF-8"?>
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
			expectedErr: nil,
			input: model.Messages{
				Language: language.English,
				Messages: []model.Message{
					{
						ID:          "Welcome",
						Message:     "{Welcome to our website!}",
						Description: "To welcome a new visitor",
					},
					{
						ID:          "Error",
						Message:     "{Something went wrong. Please try again later.}",
						Description: "To inform the user of an error",
					},
					{
						ID:      "Feedback",
						Message: "{We appreciate your feedback. Thank you for using our service.}",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			actual, err := ToXliff2(tt.input)

			if tt.expectedErr != nil {
				assert.ErrorContains(t, err, tt.expectedErr.Error())
				return
			}

			if !assert.NoError(t, err) {
				return
			}

			assertEqualXml(t, tt.expected, actual)
		})
	}
}

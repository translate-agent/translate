package convert

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.expect.digital/translate/pkg/model"
	"golang.org/x/text/language"
)

func Test_FromXliff12(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		wantErr error
		data    []byte
		want    model.Messages
	}{
		{
			name: "All OK",
			data: []byte(`
<?xml version="1.0" encoding="UTF-8"?>
<xliff version="1.2" xmlns="urn:oasis:names:tc:xliff:document:1.2">
  <file source-language="en" target-language="fr" datatype="plaintext" original="ng2.template">
    <body>
      <trans-unit id="introductionHeader" datatype="html">
        <source>Hello!</source>
        <note priority="1" from="description">An introduction header for this sample</note>
      </trans-unit>
      <trans-unit id="welcomeMessage" datatype="html">
        <source>Welcome</source>
      </trans-unit>
    </body>
  </file>
</xliff>`),
			want: model.Messages{
				Language: language.English,
				Messages: []model.Message{
					{
						ID:          "introductionHeader",
						Message:     "Hello!",
						Description: "An introduction header for this sample",
					},
					{
						ID:      "welcomeMessage",
						Message: "Welcome",
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "Malformed language tag",
			data: []byte(`
<?xml version="1.0" encoding="UTF-8"?>
<xliff version="1.2" xmlns="urn:oasis:names:tc:xliff:document:1.2">
  <file source-language="xyz-ZY-Latn" target-language="fr" datatype="plaintext" original="ng2.template">
    <body>
      <trans-unit id="introductionHeader" datatype="html">
        <source>Hello!</source>
        <note priority="1" from="developer">An introduction header for this sample</note>
      </trans-unit>
      <trans-unit id="welcomeMessage" datatype="html">
        <source>Welcome</source>
      </trans-unit>
    </body>
  </file>
</xliff>`),
			wantErr: fmt.Errorf("language: subtag \"xyz\" is well-formed but unknown"),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			messages, err := FromXliff12(tt.data)

			if tt.wantErr != nil {
				assert.ErrorContains(t, err, tt.wantErr.Error())
				return
			}

			if !assert.NoError(t, err) {
				return
			}

			assert.Equal(t, tt.want.Language, messages.Language)
			assert.ElementsMatch(t, tt.want.Messages, messages.Messages)
		})
	}
}

func Test_ToXliff12(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		want     []byte
		wantErr  error
		messages model.Messages
	}{
		{
			name: "All OK",
			want: []byte(`
<xliff xmlns="urn:oasis:names:tc:xliff:document:1.2">
  <file source-language="en">
    <body>
      <trans-unit id="Welcome">
        <source>Welcome to our website!</source>
        <note>To welcome a new visitor</note>
      </trans-unit>
      <trans-unit id="Error">
        <source>Something went wrong. Please try again later.</source>
        <note>To inform the user of an error</note>
      </trans-unit>
      <trans-unit id="Feedback">
        <source>We appreciate your feedback. Thank you for using our service.</source>
      </trans-unit>
    </body>
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

			result, err := ToXliff12(tt.messages)

			if tt.wantErr != nil {
				assert.ErrorContains(t, err, tt.wantErr.Error())
				return
			}

			if !assert.NoError(t, err) {
				return
			}

			// Matches a substring that starts with > and ends with < with zero or more whitespace in between.
			re := regexp.MustCompile(`>(\s*)<`)
			resultTrimmed := re.ReplaceAllString(string(result), "><")
			wantTrimmed := re.ReplaceAllString(string(tt.want), "><")

			assert.Equal(t, resultTrimmed, wantTrimmed)
		})
	}
}

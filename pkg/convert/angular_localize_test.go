package convert

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.expect.digital/translate/pkg/model"
	"golang.org/x/text/language"
)

func Test_FromNG_JSON(t *testing.T) {
	t.Parallel()

	tests := []struct {
		wantErr error
		data    []byte
		name    string
		want    model.Messages
	}{
		{
			name: "All OK",
			data: []byte(`{
  "locale": "fr",
  "translations": {
    "Hello": "Bonjour",
    "Welcome": "Bienvenue"
  }
}
`),
			want: model.Messages{
				Language: language.French,
				Messages: []model.Message{
					{
						ID:      "Hello",
						Message: "Bonjour",
					},
					{
						ID:      "Welcome",
						Message: "Bienvenue",
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "Malformed language tag",
			data: []byte(`{
  "locale": "xyz-ZY-Latn",
  "translations": {
    "Hello": "Bonjour",
    "Welcome": "Bienvenue"
  }
}
`),
			wantErr: fmt.Errorf("subtag \"xyz\" is well-formed but unknown"),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			messages, err := FromNG(tt.data)

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

func Test_FromNG_XLF12(t *testing.T) {
	t.Parallel()

	tests := []struct {
		wantErr error
		data    []byte
		name    string
		want    model.Messages
	}{
		{
			name: "All OK",
			data: []byte(`<?xml version="1.0" encoding="UTF-8"?>
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
</xliff>
`),
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
			data: []byte(`<?xml version="1.0" encoding="UTF-8"?>
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
</xliff>
`),
			wantErr: fmt.Errorf("subtag \"xyz\" is well-formed but unknown"),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			messages, err := FromNG(tt.data)

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

package convert

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.expect.digital/translate/pkg/model"
	"golang.org/x/text/language"
)

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
			data: []byte(`<?xml version="1.0" encoding="UTF-8" ?>
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
			data: []byte(`<?xml version="1.0" encoding="UTF-8" ?>
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

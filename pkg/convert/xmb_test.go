package convert

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.expect.digital/translate/pkg/model"
	"golang.org/x/text/language"
)

func TestFromXMB(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		expectedErr error
		input       []byte
		expected    model.Messages
	}{
		{
			name: "Valid xmb",
			input: []byte(`<?xml version="1.0" encoding="utf-8"?>
<messagebundle srcLang="en">
  <message id="hello" desc="Greeting" fuzzy="true">
    <content>Hello World!</content>
  </message>
  <message id="goodbye" desc="farewell" fuzzy="true">
    <content>Goodbye!</content>
  </message>
</messagebundle>`),
			expected: model.Messages{
				Language: language.English,
				Messages: []model.Message{
					{
						ID:          "hello",
						Message:     "Hello World!",
						Description: "Greeting",
						Fuzzy:       true,
					},
					{
						ID:          "goodbye",
						Message:     "Goodbye!",
						Description: "farewell",
						Fuzzy:       true,
					},
				},
			},
			expectedErr: nil,
		},
		{
			name: "When fuzzy values are missing",
			input: []byte(`<?xml version="1.0" encoding="utf-8"?>
<messagebundle srcLang="en">
  <message id="hello" desc="Greeting" fuzzy="false">
    <content>Hello World!</content>
  </message>
  <message id="goodbye" desc="farewell" fuzzy="false">
    <content>Goodbye!</content>
  </message>
</messagebundle>`),
			expected: model.Messages{
				Language: language.English,
				Messages: []model.Message{
					{
						ID:          "hello",
						Message:     "Hello World!",
						Description: "Greeting",
					},
					{
						ID:          "goodbye",
						Message:     "Goodbye!",
						Description: "farewell",
					},
				},
			},
			expectedErr: nil,
		},
		{
			name: "When description values are missing",
			input: []byte(`<?xml version="1.0" encoding="utf-8"?>
<messagebundle srcLang="en">
  <message id="hello" fuzzy="true">
    <content>Hello World!</content>
  </message>
  <message id="goodbye" fuzzy="true">
    <content>Goodbye!</content>
  </message>
</messagebundle>`),
			expected: model.Messages{
				Language: language.English,
				Messages: []model.Message{
					{
						ID:      "hello",
						Message: "Hello World!",
						Fuzzy:   true,
					},
					{
						ID:      "goodbye",
						Message: "Goodbye!",
						Fuzzy:   true,
					},
				},
			},
			expectedErr: nil,
		},
		{
			name:        "Invalid JSON",
			input:       []byte(`{"message": "example"`),
			expectedErr: fmt.Errorf("unmarshal from xmb to model.Messages: unexpected end of JSON input"),
			expected:    model.Messages{},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			actual, err := FromXMB(tt.input)

			if tt.expectedErr != nil {
				assert.Errorf(t, err, tt.expectedErr.Error())
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.expected, actual)
		})
	}
}

func TestToXMB(t *testing.T) {
	t.Parallel()

	input := model.Messages{
		Language: language.English,
		Messages: []model.Message{
			{
				ID:          "hello",
				Message:     "Hello World!",
				Description: "Greeting",
				Fuzzy:       true,
			},
			{
				ID:          "goodbye",
				Message:     "Goodbye!",
				Description: "farewell",
				Fuzzy:       true,
			},
		},
	}

	expected := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<messagebundle srcLang="en">
<message id="hello" desc="Greeting" fuzzy="true">
<content>Hello World!</content>
</message>
<message id="goodbye" desc="farewell" fuzzy="true">
<content>Goodbye!</content>
</message>
</messagebundle>`)

	actual, err := ToXMB(input)

	if !assert.NoError(t, err) {
		return
	}

	r := regexp.MustCompile(`>(\s*)<`)
	actualTrimmed := r.ReplaceAllString(string(actual), "><")
	expectedTrimmed := r.ReplaceAllString(string(expected), "><")

	assert.Equal(t, expectedTrimmed, actualTrimmed)
}

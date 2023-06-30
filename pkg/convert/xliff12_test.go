package convert

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.expect.digital/translate/pkg/model"
	"golang.org/x/text/language"
)

func Test_FromXliff12(t *testing.T) {
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
<xliff version="1.2" xmlns="urn:oasis:names:tc:xliff:document:1.2">
  <file source-language="en" target-language="fr" datatype="plaintext" original="ng2.template">
    <body>
      <trans-unit id="introductionHeader" datatype="html">
        <source>Hello!</source>
		<context-group purpose="location">
			<context context-type="sourcefile">src/config.ts</context>
			<context context-type="linenumber">50</context>
	  	</context-group>
	  	<context-group purpose="location">
			<context context-type="sourcefile">src/config.html</context>
			<context context-type="linenumber">60</context>
		</context-group>
        <note priority="1" from="description">An introduction header for this sample</note>
      </trans-unit>
      <trans-unit id="welcomeMessage" datatype="html">
        <source>Welcome</source>
		<context-group purpose="location">
			<context context-type="sourcefile">src/config1.ts</context>
			<context context-type="linenumber">40</context>
	  	</context-group>
	  	<context-group purpose="location">
			<context context-type="sourcefile">src/config2.html</context>
			<context context-type="linenumber">50</context>
			<context context-type="sourcefile">src/config3.html</context>
			<context context-type="sourcefile">src/config4.html</context>
			<context context-type="sourcefile">src/config5.html</context>
			<context context-type="linenumber">60</context>
		</context-group>
      </trans-unit>
    </body>
  </file>
</xliff>`),
			expected: model.Messages{
				Language: language.English,
				Messages: []model.Message{
					{
						ID:          "introductionHeader",
						Message:     "{Hello!}",
						Description: "An introduction header for this sample",
						Positions:   []string{"src/config.ts:50", "src/config.html:60"},
					},
					{
						ID:      "welcomeMessage",
						Message: "{Welcome}",
						Positions: []string{
							"src/config1.ts:40", "src/config2.html:50",
							"src/config3.html", "src/config4.html", "src/config5.html:60",
						},
					},
				},
			},
			expectedErr: nil,
		},
		{
			name: "Malformed language",
			input: []byte(`<?xml version="1.0" encoding="UTF-8"?>
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
			expectedErr: fmt.Errorf("language: subtag \"xyz\" is well-formed but unknown"),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			actual, err := FromXliff12(tt.input)

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

func Test_ToXliff12(t *testing.T) {
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
<xliff xmlns="urn:oasis:names:tc:xliff:document:1.2" version="1.2">
  <file source-language="en">
    <body>
      <trans-unit id="Welcome">
        <source>Welcome to our website!</source>
        <note>To welcome a new visitor</note>
		<context-group purpose="location">
			<context context-type="sourcefile">src/config.ts</context>
			<context context-type="linenumber">50</context>
		</context-group>
		<context-group purpose="location">
			<context context-type="sourcefile">src/config.html</context>
			<context context-type="linenumber">60</context>
		</context-group>
      </trans-unit>
      <trans-unit id="Error">
        <source>Something went wrong. Please try again later.</source>
        <note>To inform the user of an error</note>
		<context-group purpose="location">
			<context context-type="sourcefile">src/config1.ts</context>
			<context context-type="linenumber">40</context>
		</context-group>
		<context-group purpose="location">
			<context context-type="sourcefile">src/config2.html</context>
			<context context-type="linenumber">50</context>
		</context-group>
		<context-group purpose="location">
			<context context-type="sourcefile">src/config3.html</context>
		</context-group>
		<context-group purpose="location">
			<context context-type="sourcefile">src/config4.html</context>
		</context-group>
		<context-group purpose="location">
			<context context-type="sourcefile">src/config5.html</context>
			<context context-type="linenumber">60</context>
		</context-group>
      </trans-unit>
      <trans-unit id="Feedback">
        <source>We appreciate your feedback. Thank you for using our service.</source>
      </trans-unit>
    </body>
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
						Positions:   []string{"src/config.ts:50", "src/config.html:60"},
					},
					{
						ID:          "Error",
						Message:     "{Something went wrong. Please try again later.}",
						Description: "To inform the user of an error",
						Positions: []string{
							"src/config1.ts:40", "src/config2.html:50",
							"src/config3.html", "src/config4.html", "src/config5.html:60",
						},
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

			actual, err := ToXliff12(tt.input)

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

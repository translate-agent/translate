package convert

import (
	"testing"

	"golang.org/x/text/language"

	"github.com/stretchr/testify/require"
	"go.expect.digital/translate/pkg/model"
	"go.expect.digital/translate/pkg/testutil"
)

// Test_FromXliff2 tests Xliff 2.0 default 'acc. to specification' format.
func Test_FromXliff2_Default(t *testing.T) {
	t.Parallel()

	tests := []struct {
		expected    *model.Translation
		expectedErr error
		name        string
		data        []byte
	}{
		{
			name: "original, source content with escaped chars",
			data: []byte(`<?xml version="1.0" encoding="UTF-8" ?>
			<xliff version="2.0"
				xmlns="urn:oasis:names:tc:xliff:document:2.0" srcLang="en">
				<file>
					<unit id="1">
						<segment>
							<source>escaped chars: &amp;&lt;&gt;&quot;&apos;</source>
						</segment>
					</unit>
				</file>
			</xliff>`),
			expected: &model.Translation{
				Original: true,
				Language: language.English,
				Messages: []model.Message{
					{
						ID:      "1",
						Message: `escaped chars: &<>"'`,
						Status:  model.MessageStatusTranslated,
					},
				},
			},
		},
		{
			name: "original, source content with marked span of text",
			data: []byte(`<?xml version="1.0" encoding="UTF-8" ?>
			<xliff version="2.0"
				xmlns="urn:oasis:names:tc:xliff:document:2.0" srcLang="en">
				<file>
					<unit id="2">
						<segment>
							<source>This is a <mrk id="1" type="term">placeholder</mrk> example.</source>
						</segment>
					</unit>
				</file>
			</xliff>`),
			expected: &model.Translation{
				Original: true,
				Language: language.English,
				Messages: []model.Message{
					{
						ID: "2",
						Message: `.local $mrk1 = { |<mrk id="1" type="term">placeholder</mrk>| }
{{This is a { $mrk1 } example.}}`,
						Status: model.MessageStatusTranslated,
					},
				},
			},
		},
		{
			name: "original, source content with placeholders",
			data: []byte(`<?xml version="1.0" encoding="UTF-8" ?>
			<xliff version="2.0"
				xmlns="urn:oasis:names:tc:xliff:document:2.0" srcLang="en">
				<file>
					<unit id="3">
						<notes>
						<note category="location">src/app/component.html:19</note>
						<note category="description">component</note>
						</notes>
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
						ID: "3",
						Message: `.local $ph1 = { |<ph id="1" dataRef="d1" canCopy="no" canDelete="no" canOverlap="yes"/>| }
.local $ph2 = { |<ph id="2" dataRef="d2" canCopy="no" canDelete="no" canOverlap="yes"/>| }
{{Entries: { $ph1 }!{ $ph2 }(Filtered)}}`,
						Status:      model.MessageStatusTranslated,
						Description: "component",
						Positions:   []string{"src/app/component.html:19"},
					},
				},
			},
		},
		{
			name: "translation, target content with placeholders",
			data: []byte(`<?xml version="1.0" encoding="UTF-8" ?>
					<xliff version="2.0"
						xmlns="urn:oasis:names:tc:xliff:document:2.0" srcLang ="en" trgLang="en">
						<file>
							<unit id="4">
							<notes>
							<note category="location">src/app/component.html:19</note>
							<note category="description">component</note>
							<note category="description">project-id</note>
							</notes>
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
						ID: "4",
						Message: `.local $ph1 = { |<ph id="1" canCopy="no" canDelete="no" canOverlap="yes"/>| }
.local $ph2 = { |<ph id="2" canCopy="no" canDelete="no" canOverlap="yes"/>| }
{{Entries: { $ph1 }!{ $ph2 }(Filtered)}}`,
						Status:      model.MessageStatusUntranslated,
						Description: "component\nproject-id",
						Positions:   []string{"src/app/component.html:19"},
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

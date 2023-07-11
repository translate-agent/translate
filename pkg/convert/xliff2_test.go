package convert

import (
	"fmt"
	"math/rand"
	"reflect"
	"regexp"
	"strings"
	"testing"
	"testing/quick"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.expect.digital/translate/pkg/model"
	"go.expect.digital/translate/pkg/testutil"
	testutilrand "go.expect.digital/translate/pkg/testutil/rand"
)

func randXliff2(messages *model.Messages) []byte {
	sb := strings.Builder{}

	sb.WriteString(`<?xml version="1.0" encoding="UTF-8"?>`)

	if messages.Original {
		fmt.Fprintf(
			&sb,
			"<xliff xmlns=\"urn:oasis:names:tc:xliff:document:2.0\" version=\"2.0\" srcLang=\"%s\" trgLang=\"und\">",
			messages.Language)
	} else {
		fmt.Fprintf(
			&sb,
			"<xliff xmlns=\"urn:oasis:names:tc:xliff:document:2.0\" version=\"2.0\" srcLang=\"und\" trgLang=\"%s\">",
			messages.Language)
	}

	sb.WriteString("<file>")

	for _, msg := range messages.Messages {
		fmt.Fprintf(&sb, "<unit id=\"%s\">", msg.ID)

		if msg.Description != "" {
			fmt.Fprintf(&sb, "<notes><note category=\"description\">%s</note></notes>", msg.Description)
		}

		if messages.Original {
			fmt.Fprintf(&sb, "<segment><source>%s</source></segment>", msg.Message)
		} else {
			fmt.Fprintf(&sb, "<segment><target>%s</target></segment>", msg.Message)
		}

		sb.WriteString("</unit>")
	}

	sb.WriteString("</file>")
	sb.WriteString("</xliff>")

	return []byte(sb.String())
}

func assertEqualXml(t *testing.T, expected, actual []byte) bool { //nolint:unparam
	t.Helper()
	// Matches a substring that starts with > and ends with < with zero or more whitespace in between.
	re := regexp.MustCompile(`>(\s*)<`)
	expectedTrimmed := re.ReplaceAllString(string(expected), "><")
	actualTrimmed := re.ReplaceAllString(string(actual), "><")

	return assert.Equal(t, expectedTrimmed, actualTrimmed)
}

func Test_FromXliff2(t *testing.T) {
	t.Parallel()

	msgOpts := []testutilrand.ModelMessageOption{
		testutilrand.WithFuzzy(false), // Do not mark message as fuzzy, as this is not supported by XLIFF 2.0
	}

	sourceMessages := testutilrand.ModelMessages(3, msgOpts, testutilrand.WithOriginal(true))
	translatedMessages := testutilrand.ModelMessages(3, msgOpts, testutilrand.WithOriginal(false))

	tests := []struct {
		name     string
		expected *model.Messages
		input    []byte
	}{
		{
			name:     "Happy Path Untranslated",
			input:    randXliff2(sourceMessages),
			expected: sourceMessages,
		},
		{
			name:     "Happy Path Translated",
			input:    randXliff2(translatedMessages),
			expected: translatedMessages,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			actual, err := FromXliff2(tt.input)
			require.NoError(t, err)

			for i := range actual.Messages {
				actual.Messages[i].Message = strings.Trim(actual.Messages[i].Message, "{}") // Remove curly braces for comparison
			}

			// TODO: for now restore the flag to the expected
			// remove this as XLIFF is the format were we can implicitly determine if file is original or not
			actual.Original = tt.expected.Original

			testutil.EqualMessages(t, tt.expected, &actual)
		})
	}
}

func Test_ToXliff2(t *testing.T) {
	t.Parallel()

	msgOpts := []testutilrand.ModelMessageOption{
		testutilrand.WithFuzzy(false), // Do not mark message as fuzzy, as this is not supported by XLIFF 2.0
	}

	messages := testutilrand.ModelMessages(4, msgOpts, testutilrand.WithOriginal(true))
	expected := randXliff2(messages)

	actual, err := ToXliff2(*messages)
	require.NoError(t, err)

	assertEqualXml(t, expected, actual)
}

func Test_TransformXLIFF2(t *testing.T) {
	t.Parallel()

	msgOpts := []testutilrand.ModelMessageOption{
		// Enclose message in curly braces, as ToXliff2() removes them, and FromXliff2() adds them again
		testutilrand.WithMessageFormat(),
		testutilrand.WithFuzzy(false), // Do not mark message as fuzzy, as this is not supported by XLIFF 1.2
	}

	conf := &quick.Config{
		MaxCount: 100,
		Values: func(values []reflect.Value, _ *rand.Rand) {
			values[0] = reflect.ValueOf(
				testutilrand.ModelMessages(3, msgOpts, testutilrand.WithOriginal(true))) // input generator
		},
	}

	f := func(expected *model.Messages) bool {
		xliffData, err := ToXliff2(*expected)
		require.NoError(t, err)

		restoredMessages, err := FromXliff2(xliffData)
		require.NoError(t, err)

		// TODO: for now restore the flag to the expected
		// remove this as XLIFF is the format were we can implicitly determine if file is original or not
		restoredMessages.Original = expected.Original

		testutil.EqualMessages(t, expected, &restoredMessages)

		return true
	}

	assert.NoError(t, quick.Check(f, conf))
}

//go:build integration

package po

import (
	"bytes"
	"io"
	"net/http"
	"reflect"
	"slices"
	"testing"
)

const link = "https://raw.githubusercontent.com/apache/superset/cc2f6f1ed962ae1886c4eb5c4ce1b094ddc7fe9c/superset/translations/nl/LC_MESSAGES/messages.po" //nolint:lll

func parseFile(t *testing.T, reader io.Reader) PO {
	t.Helper()

	var buff bytes.Buffer

	_, err := io.Copy(&buff, reader)
	if err != nil {
		t.Error(err)
		return PO{}
	}

	po, err := Parse(buff.Bytes())
	if err != nil {
		t.Error(err)
		return PO{}
	}

	return po
}

func Test_ParseSuperset(t *testing.T) {
	t.Parallel()

	resp, err := http.Get(link) //nolint:noctx
	if err != nil {
		t.Error(err)
		return
	}

	defer resp.Body.Close()

	got := parseFile(t, resp.Body)

	wantHeaders := Headers{
		{Name: "Project-Id-Version", Value: "Superset VERSION"},
		{Name: "Report-Msgid-Bugs-To", Value: "EMAIL@ADDRESS"},
		{Name: "POT-Creation-Date", Value: "2024-02-16 13:50-0500"},
		{Name: "PO-Revision-Date", Value: "2022-02-25 11:59+0100"},
		{Name: "Last-Translator", Value: "FULL NAME <EMAIL@ADDRESS>"},
		{Name: "Language", Value: "nl"},
		{Name: "Language-Team", Value: "nl <LL@li.org>"},
		{Name: "Plural-Forms", Value: "nplurals=2; plural=(n != 1)"},
		{Name: "MIME-Version", Value: "1.0"},
		{Name: "Content-Type", Value: "text/plain; charset=utf-8"},
		{Name: "Content-Transfer-Encoding", Value: "8bit"},
		{Name: "Generated-By", Value: "Babel 2.9.1"},
	}

	if !slices.Equal(wantHeaders, got.Headers) {
		t.Errorf("want headers %v, got %v", wantHeaders, got.Headers)
	}

	someWantMessages := []Message{
		{
			MsgID:  "STEP %(stepCurr)s OF %(stepLast)s",
			MsgStr: []string{},
			Flags:  []string{"python-format"},
			References: []string{
				"superset-frontend/src/features/databases/DatabaseModal/ModalHeader.tsx:93",
				"superset-frontend/src/features/databases/DatabaseModal/ModalHeader.tsx:117",
				"superset-frontend/src/features/databases/DatabaseModal/ModalHeader.tsx:135",
				"superset-frontend/src/features/databases/DatabaseModal/ModalHeader.tsx:164",
				"superset-frontend/src/features/databases/DatabaseModal/ModalHeader.tsx:178",
			},
		},
		{
			MsgID:       "Please reach out to the Chart Owner for assistance.",
			MsgIDPlural: "Please reach out to the Chart Owners for assistance.",
			MsgStr:      []string{"Neem contact op met de eigenaar van de grafiek voor hulp.", ""},
			Flags:       []string{"fuzzy"},
			References: []string{
				"superset-frontend/src/components/ErrorMessage/DatabaseErrorMessage.tsx:59",
				"superset-frontend/src/components/ErrorMessage/TimeoutErrorMessage.tsx:72",
			},
		},
		{
			MsgID:       "Deleted %(num)d annotation layer",
			MsgIDPlural: "Deleted %(num)d annotation layers",
			MsgStr: []string{
				"%(num)d Aantekeningenlaag verwijderd",
				"%(num)d aantekeninglagen verwijderd",
			},
			Flags:      []string{"python-format"},
			References: []string{"superset/annotation_layers/api.py:346"},
		},
	}

	for _, want := range someWantMessages {
		if !slices.ContainsFunc(got.Messages, func(message Message) bool {
			return reflect.DeepEqual(message, want)
		}) {
			t.Errorf("want %v to contain %s", got.Messages, want)
		}
	}
}

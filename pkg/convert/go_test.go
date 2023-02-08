package convert

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"go.expect.digital/translate/pkg/model"
	"golang.org/x/text/language"
)

var modelMsg = model.Messages{
	Language: language.English,
	Messages: []model.Message{
		{
			ID:          "1",
			Message:     "message1",
			Description: "description1",
			Fuzzy:       true,
		},
		{
			ID:          "2",
			Message:     "message2",
			Description: "description2",
			Fuzzy:       false,
		},
	},
}

func TestToGo(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		m       model.Messages
		want    []byte
		wantErr bool
	}{
		{
			name:    "When input is correct, then return expected result",
			m:       modelMsg,
			want:    []byte(`{"language":"en","messages":[{"id":"1","meaning":"description1","message":"","translation":"message1","fuzzy":true},{"id":"2","meaning":"description2","message":"","translation":"message2"}]}`), //nolint:lll
			wantErr: false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result, err := ToGo(tt.m)

			assert.NoError(t, err)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestFromGo(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		m       []byte
		want    model.Messages
		wantErr bool
	}{
		{
			name: "When input is correct, then return expected result",
			m:    []byte(`{"language":"en","messages":[{"id":"1","meaning":"description1","message":"message1","translation":"","fuzzy":true},{"id":"2","meaning":"description2","message":"message2","translation":""}]}`), //nolint:lll

			want:    modelMsg,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result, err := FromGo(tt.m)

			assert.NoError(t, err)
			assert.Equal(t, tt.want, result)
		})
	}
}

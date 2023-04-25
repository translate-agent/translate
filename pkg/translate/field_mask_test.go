package translate

import (
	"errors"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/assert"
	translatev1 "go.expect.digital/translate/pkg/pb/translate/v1"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

func Test_ValidateFieldMask(t *testing.T) {
	t.Parallel()

	type args struct {
		fieldMask    *fieldmaskpb.FieldMask
		protoMessage proto.Message
	}

	happyPathService := args{
		fieldMask:    &fieldmaskpb.FieldMask{Paths: []string{"name"}},
		protoMessage: &translatev1.Service{},
	}

	happyPathMessage := args{
		fieldMask:    &fieldmaskpb.FieldMask{Paths: []string{"description"}},
		protoMessage: &translatev1.Message{},
	}

	randomFieldMaskPathMessage := args{
		fieldMask:    &fieldmaskpb.FieldMask{Paths: []string{gofakeit.FirstName()}},
		protoMessage: &translatev1.Message{},
	}

	tests := []struct {
		args        args
		expectedErr error
		name        string
	}{
		{
			name:        "Happy Path Service",
			args:        happyPathService,
			expectedErr: nil,
		},
		{
			name:        "Happy Path Message",
			args:        happyPathMessage,
			expectedErr: nil,
		},
		{
			name:        "Random Field Mask Path",
			args:        randomFieldMaskPathMessage,
			expectedErr: errors.New("not a valid"),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := validateFieldMask(tt.args.fieldMask, tt.args.protoMessage)
			if tt.expectedErr != nil {
				assert.ErrorContains(t, err, tt.expectedErr.Error())
				return
			}

			assert.NoError(t, err)
		})
	}
}

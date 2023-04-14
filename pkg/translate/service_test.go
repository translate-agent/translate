package translate

import (
	"errors"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.expect.digital/translate/pkg/model"
	translatev1 "go.expect.digital/translate/pkg/pb/translate/v1"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

// ----------------------GetService Parse Params-------------------------------

func randGetServiceReq() *translatev1.GetServiceRequest {
	return &translatev1.GetServiceRequest{
		Id: gofakeit.UUID(),
	}
}

func Test_ParseGetServiceParams(t *testing.T) {
	t.Parallel()

	happyReq := randGetServiceReq()

	malformedIDReq := randGetServiceReq()
	malformedIDReq.Id += "_FAIL" //nolint:goconst

	tests := []struct {
		expected    *getServiceParams
		input       *translatev1.GetServiceRequest
		expectedErr error
		name        string
	}{
		{
			name:  "Happy Path",
			input: happyReq,
			expected: &getServiceParams{
				id: uuid.MustParse(happyReq.Id),
			},
			expectedErr: nil,
		},
		{
			name:        "Malformed UUID",
			input:       malformedIDReq,
			expectedErr: errors.New("invalid UUID length"),
		},
		{
			name:        "NIL request",
			input:       nil,
			expectedErr: errors.New("request is nil"),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := (*getServiceRequest)(tt.input)

			actual, err := req.parseParams()

			if tt.expectedErr != nil {
				assert.ErrorContains(t, err, tt.expectedErr.Error())
				return
			}

			require.NoError(t, err)

			assert.Equal(t, tt.expected, actual)
		})
	}
}

// -----------------------UpdateService Parse Params-------------------------------

func randUpdateServiceReq() *translatev1.UpdateServiceRequest {
	return &translatev1.UpdateServiceRequest{
		UpdateMask: &fieldmaskpb.FieldMask{Paths: gofakeit.NiceColors()},
		Service: &translatev1.Service{
			Id:   gofakeit.UUID(),
			Name: gofakeit.Name(),
		},
	}
}

func Test_ParseUpdateServiceParams(t *testing.T) {
	t.Parallel()

	happyReq := randUpdateServiceReq()

	malformedIDReq := randUpdateServiceReq()
	malformedIDReq.Service.Id += "_FAIL"

	tests := []struct {
		expected    *updateServiceParams
		expectedErr error
		input       *translatev1.UpdateServiceRequest
		name        string
	}{
		{
			name:  "Happy Path",
			input: happyReq,
			expected: &updateServiceParams{
				mask: happyReq.UpdateMask,
				service: &model.Service{
					ID:   uuid.MustParse(happyReq.Service.Id),
					Name: happyReq.Service.Name,
				},
			},
			expectedErr: nil,
		},
		{
			name:        "Malformed Service UUID",
			input:       malformedIDReq,
			expectedErr: errors.New("invalid UUID length"),
		},
		{
			name:        "NIL request",
			input:       nil,
			expectedErr: errors.New("request is nil"),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := (*updateServiceRequest)(tt.input)

			actual, err := req.parseParams()

			if tt.expectedErr != nil {
				assert.ErrorContains(t, err, tt.expectedErr.Error())
				return
			}

			require.NoError(t, err)

			assert.Equal(t, tt.expected, actual)
		})
	}
}

// ----------------------DeleteService Parse Params------------------------------

func randDeleteServiceReq() *translatev1.DeleteServiceRequest {
	return &translatev1.DeleteServiceRequest{
		Id: gofakeit.UUID(),
	}
}

func Test_ParseDeleteServiceParams(t *testing.T) {
	t.Parallel()

	happyReq := randDeleteServiceReq()

	malformedIDReq := randDeleteServiceReq()
	malformedIDReq.Id += "_FAIL"

	tests := []struct {
		expected    *deleteServiceParams
		input       *translatev1.DeleteServiceRequest
		expectedErr error
		name        string
	}{
		{
			name:        "Happy Path",
			input:       happyReq,
			expected:    &deleteServiceParams{id: uuid.MustParse(happyReq.Id)},
			expectedErr: nil,
		},
		{
			name:        "Malformed UUID",
			input:       malformedIDReq,
			expectedErr: errors.New("invalid UUID length"),
		},
		{
			name:        "NIL request",
			input:       nil,
			expectedErr: errors.New("request is nil"),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := (*deleteServiceRequest)(tt.input)

			actual, err := req.parseParams()

			if tt.expectedErr != nil {
				assert.ErrorContains(t, err, tt.expectedErr.Error())
				return
			}

			require.NoError(t, err)

			assert.Equal(t, tt.expected, actual)
		})
	}
}

// ----------------------CreateService Parse Params------------------------------

func randCreateServiceReq() *translatev1.CreateServiceRequest {
	return &translatev1.CreateServiceRequest{
		Service: &translatev1.Service{
			Id:   gofakeit.UUID(),
			Name: gofakeit.Name(),
		},
	}
}

func Test_ParseCreateServiceParams(t *testing.T) {
	t.Parallel()

	happyReq := randCreateServiceReq()

	malformedIDReq := randCreateServiceReq()
	malformedIDReq.Service.Id += "_FAIL"

	tests := []struct {
		input       *translatev1.CreateServiceRequest
		expectedErr error
		expected    *createServiceParams
		name        string
	}{
		{
			name:        "Happy Path",
			input:       happyReq,
			expectedErr: nil,
			expected: &createServiceParams{
				service: &model.Service{
					ID:   uuid.MustParse(happyReq.Service.Id),
					Name: happyReq.Service.Name,
				},
			},
		},
		{
			name:        "Malformed UUID",
			input:       malformedIDReq,
			expectedErr: errors.New("invalid UUID length"),
		},
		{
			name:        "NIL request",
			input:       nil,
			expectedErr: errors.New("request is nil"),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := (*createServiceRequest)(tt.input)

			actual, err := req.parseParams()

			if tt.expectedErr != nil {
				assert.ErrorContains(t, err, tt.expectedErr.Error())
				return
			}

			require.NoError(t, err)

			assert.Equal(t, tt.expected, actual)
		})
	}
}

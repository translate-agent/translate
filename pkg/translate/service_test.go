package translate

import (
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
	malformedIDReq.Id += "_FAIL"

	tests := []struct {
		input       *translatev1.GetServiceRequest
		expectedErr *parseParamError
		name        string
		expected    getServiceParams
	}{
		{
			name:  "Happy Path",
			input: happyReq,
			expected: getServiceParams{
				id: uuid.MustParse(happyReq.Id),
			},
			expectedErr: nil,
		},
		{
			name:        "Malformed UUID",
			input:       malformedIDReq,
			expectedErr: &parseParamError{field: "id"},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := (*getServiceRequest)(tt.input)

			actual, err := req.parseParams()

			if tt.expectedErr != nil {
				var e *parseParamError
				require.ErrorAs(t, err, &e)

				// Check if parameter which caused error is the same as expected
				assert.Equal(t, tt.expectedErr.field, e.field)
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
		expected    updateServiceParams
		expectedErr *parseParamError
		input       *translatev1.UpdateServiceRequest
		name        string
	}{
		{
			name:  "Happy Path",
			input: happyReq,
			expected: updateServiceParams{
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
			expectedErr: &parseParamError{field: "service.id"},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := (*updateServiceRequest)(tt.input)

			actual, err := req.parseParams()

			if tt.expectedErr != nil {
				var e *parseParamError
				require.ErrorAs(t, err, &e)

				// Check if parameter which caused error is the same as expected
				assert.Equal(t, tt.expectedErr.field, e.field)
				return
			}

			require.NoError(t, err)

			require.Equal(t, tt.expected.service, actual.service)
			assert.ElementsMatch(t, tt.expected.mask.Paths, actual.mask.Paths)
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
		input       *translatev1.DeleteServiceRequest
		expectedErr *parseParamError
		name        string
		expected    deleteServiceParams
	}{
		{
			name:        "Happy Path",
			input:       happyReq,
			expected:    deleteServiceParams{id: uuid.MustParse(happyReq.Id)},
			expectedErr: nil,
		},
		{
			name:        "Malformed UUID",
			input:       malformedIDReq,
			expectedErr: &parseParamError{field: "id"},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := (*deleteServiceRequest)(tt.input)

			actual, err := req.parseParams()

			if tt.expectedErr != nil {
				var e *parseParamError
				require.ErrorAs(t, err, &e)

				// Check if parameter which caused error is the same as expected
				assert.Equal(t, tt.expectedErr.field, e.field)
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
		expectedErr *parseParamError
		expected    createServiceParams
		name        string
	}{
		{
			name:        "Happy Path",
			input:       happyReq,
			expectedErr: nil,
			expected: createServiceParams{
				service: &model.Service{
					ID:   uuid.MustParse(happyReq.Service.Id),
					Name: happyReq.Service.Name,
				},
			},
		},
		{
			name:        "Malformed UUID",
			input:       malformedIDReq,
			expectedErr: &parseParamError{field: "service.id"},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := (*createServiceRequest)(tt.input)

			actual, err := req.parseParams()

			if tt.expectedErr != nil {
				var e *parseParamError
				require.ErrorAs(t, err, &e)

				// Check if parameter which caused error is the same as expected
				assert.Equal(t, tt.expectedErr.field, e.field)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expected, actual)
		})
	}
}

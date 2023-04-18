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

// ----------------------GetService-------------------------------

func Test_ParseGetServiceParams(t *testing.T) {
	t.Parallel()

	randReq := func() *translatev1.GetServiceRequest {
		return &translatev1.GetServiceRequest{
			Id: gofakeit.UUID(),
		}
	}

	happyReqWithID := randReq()

	happyReqWithoutID := randReq()
	happyReqWithoutID.Id = ""

	malformedIDReq := randReq()
	malformedIDReq.Id += "_FAIL"

	tests := []struct {
		expected    *getServiceParams
		request     *translatev1.GetServiceRequest
		expectedErr *parseParamError
		name        string
	}{
		{
			name:        "Happy Path",
			request:     happyReqWithID,
			expected:    &getServiceParams{id: uuid.MustParse(happyReqWithID.Id)},
			expectedErr: nil,
		},
		{
			name:     "Happy Path Empty ID",
			request:  happyReqWithoutID,
			expected: &getServiceParams{id: uuid.Nil},
		},
		{
			name:        "Malformed ID",
			request:     malformedIDReq,
			expectedErr: &parseParamError{field: "id"},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			actual, err := parseGetServiceRequestParams(tt.request)

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

func Test_ValidateGetServiceParams(t *testing.T) {
	t.Parallel()

	tests := []struct {
		params      *getServiceParams
		expectedErr *validateParamError
		name        string
	}{
		{
			name:        "Happy Path",
			expectedErr: nil,
			params:      &getServiceParams{id: uuid.New()},
		},
		{
			name:        "Empty ID",
			params:      &getServiceParams{id: uuid.Nil},
			expectedErr: &validateParamError{param: "id"},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := validateGetServiceRequestParams(tt.params)

			if tt.expectedErr != nil {
				var e *validateParamError
				require.ErrorAs(t, err, &e)

				// Check if parameter which caused error is the same as expected
				assert.Equal(t, tt.expectedErr.param, e.param)
				return
			}

			assert.NoError(t, err)
		})
	}
}

// -----------------------UpdateService-------------------------------

func randService() *model.Service {
	return &model.Service{
		ID:   uuid.New(),
		Name: gofakeit.Name(),
	}
}

func Test_ParseUpdateServiceParams(t *testing.T) {
	t.Parallel()

	randReq := func() *translatev1.UpdateServiceRequest {
		return &translatev1.UpdateServiceRequest{
			UpdateMask: &fieldmaskpb.FieldMask{Paths: gofakeit.NiceColors()},
			Service:    serviceToProto(randService()),
		}
	}

	happyReq := randReq()

	happyReqWithoutServiceID := randReq()
	happyReqWithoutServiceID.Service.Id = ""

	happyReqWithoutService := randReq()
	happyReqWithoutService.Service = nil

	happyReqWithoutUpdateMask := randReq()
	happyReqWithoutUpdateMask.UpdateMask = nil

	malformedIDReq := randReq()
	malformedIDReq.Service.Id += "_FAIL"

	tests := []struct {
		expected    *updateServiceParams
		expectedErr *parseParamError
		request     *translatev1.UpdateServiceRequest
		name        string
	}{
		{
			name:    "Happy Path",
			request: happyReq,
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
			name:    "Happy Path Without Service ID",
			request: happyReqWithoutServiceID,
			expected: &updateServiceParams{
				mask: happyReqWithoutServiceID.UpdateMask,
				service: &model.Service{
					ID:   uuid.Nil,
					Name: happyReqWithoutServiceID.Service.Name,
				},
			},
			expectedErr: nil,
		},

		{
			name:    "Happy Path Without Service",
			request: happyReqWithoutService,
			expected: &updateServiceParams{
				mask:    happyReqWithoutService.UpdateMask,
				service: nil,
			},
		},
		{
			name:    "Happy Path Without Update Mask",
			request: happyReqWithoutUpdateMask,
			expected: &updateServiceParams{
				mask: nil,
				service: &model.Service{
					ID:   uuid.MustParse(happyReqWithoutUpdateMask.Service.Id),
					Name: happyReqWithoutUpdateMask.Service.Name,
				},
			},
		},
		{
			name:        "Malformed Service ID",
			request:     malformedIDReq,
			expectedErr: &parseParamError{field: "service.id"},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			actual, err := parseUpdateServiceParams(tt.request)

			if tt.expectedErr != nil {
				var e *parseParamError
				require.ErrorAs(t, err, &e)

				// Check if parameter which caused error is the same as expected
				assert.Equal(t, tt.expectedErr.field, e.field)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.expected, actual)
		})
	}
}

func Test_ValidateUpdateServiceParams(t *testing.T) {
	t.Parallel()

	randParams := func() *updateServiceParams {
		return &updateServiceParams{
			mask:    &fieldmaskpb.FieldMask{Paths: []string{"name"}},
			service: randService(),
		}
	}

	happyParams := randParams()

	happyParamsMissingServiceID := randParams()
	happyParamsMissingServiceID.service.ID = uuid.Nil

	happyParamsMissingUpdateMask := randParams()
	happyParamsMissingUpdateMask.mask = nil

	missingServiceParams := randParams()
	missingServiceParams.service = nil

	invalidUpdateMaskPathParams := randParams()
	invalidUpdateMaskPathParams.mask.Paths = gofakeit.NiceColors()

	tests := []struct {
		params      *updateServiceParams
		expectedErr *validateParamError
		name        string
	}{
		{
			name:        "Happy Path",
			params:      happyParams,
			expectedErr: nil,
		},
		{
			name:        "Happy Path Missing Service ID",
			params:      happyParamsMissingServiceID,
			expectedErr: nil,
		},
		{
			name:        "Happy Path Missing Update Mask",
			params:      happyParamsMissingUpdateMask,
			expectedErr: nil,
		},
		{
			name:        "Missing Service",
			params:      missingServiceParams,
			expectedErr: &validateParamError{param: "service"},
		},

		{
			name:        "Invalid Update Mask Path",
			params:      invalidUpdateMaskPathParams,
			expectedErr: &validateParamError{param: "update_mask.paths"},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := validateUpdateServiceParams(tt.params)

			if tt.expectedErr != nil {
				var e *validateParamError
				require.ErrorAs(t, err, &e)

				// Check if parameter which caused error is the same as expected
				assert.Equal(t, tt.expectedErr.param, e.param)
				return
			}

			assert.NoError(t, err)
		})
	}
}

func Test_UpdateServiceFromParams(t *testing.T) {
	t.Parallel()

	originalService1 := randService()
	originalService2 := randService()
	originalService3 := randService()

	// For now, we only support updating the name, as that is the only field that is updatable.

	randParams := func(originalId uuid.UUID) *updateServiceParams {
		return &updateServiceParams{
			mask:    &fieldmaskpb.FieldMask{Paths: []string{"name"}},
			service: &model.Service{ID: originalId, Name: gofakeit.Name()},
		}
	}

	updateNameField := randParams(originalService1.ID)

	updateAllFields := randParams(originalService2.ID)
	updateAllFields.mask = nil

	nothingToUpdate := randParams(originalService3.ID)
	nothingToUpdate.mask = &fieldmaskpb.FieldMask{Paths: []string{"random_path"}}

	tests := []struct {
		params          *updateServiceParams
		originalService *model.Service
		expectedService *model.Service
		name            string
	}{
		{
			name:            "Update Name Field",
			params:          updateNameField,
			originalService: originalService1,
			expectedService: &model.Service{ID: originalService1.ID, Name: updateNameField.service.Name},
		},
		{
			name:            "Update All Fields",
			params:          updateAllFields,
			originalService: originalService2,
			expectedService: &model.Service{ID: originalService2.ID, Name: updateAllFields.service.Name},
		},
		{
			name:            "Nothing To Update",
			params:          nothingToUpdate,
			originalService: originalService3,
			expectedService: originalService3,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			actualService := updateServiceFromParams(tt.originalService, tt.params)

			assert.Equal(t, tt.expectedService, actualService)
		})
	}
}

// ----------------------DeleteService Parse Params------------------------------

func Test_ParseDeleteServiceParams(t *testing.T) {
	t.Parallel()

	randReq := func() *translatev1.DeleteServiceRequest {
		return &translatev1.DeleteServiceRequest{
			Id: gofakeit.UUID(),
		}
	}

	happyReqWithID := randReq()

	happyReqWithoutID := randReq()
	happyReqWithoutID.Id = ""

	malformedIDReq := randReq()
	malformedIDReq.Id += "_FAIL"

	tests := []struct {
		expected    *deleteServiceParams
		request     *translatev1.DeleteServiceRequest
		expectedErr *parseParamError
		name        string
	}{
		{
			name:        "Happy Path With ID",
			request:     happyReqWithID,
			expected:    &deleteServiceParams{id: uuid.MustParse(happyReqWithID.Id)},
			expectedErr: nil,
		},
		{
			name:        "Happy Path Without ID",
			request:     happyReqWithoutID,
			expected:    &deleteServiceParams{id: uuid.Nil},
			expectedErr: nil,
		},
		{
			name:        "Malformed ID",
			request:     malformedIDReq,
			expectedErr: &parseParamError{field: "id"},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			actual, err := parseDeleteServiceRequest(tt.request)

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

func Test_ValidateDeleteServiceParams(t *testing.T) {
	t.Parallel()

	randParams := func() *deleteServiceParams {
		return &deleteServiceParams{id: uuid.New()}
	}

	happyParams := randParams()

	emptyIdParams := randParams()
	emptyIdParams.id = uuid.Nil

	tests := []struct {
		params      *deleteServiceParams
		expectedErr *validateParamError
		name        string
	}{
		{
			name:        "Happy Path",
			params:      happyParams,
			expectedErr: nil,
		},
		{
			name:        "Empty ID",
			params:      emptyIdParams,
			expectedErr: &validateParamError{param: "id"},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := validateDeleteServiceParams(tt.params)

			if tt.expectedErr != nil {
				var e *validateParamError
				require.ErrorAs(t, err, &e)

				// Check if parameter which caused error is the same as expected
				assert.Equal(t, tt.expectedErr.param, e.param)
				return
			}

			assert.NoError(t, err)
		})
	}
}

// ----------------------CreateService------------------------------

func Test_ParseCreateServiceParams(t *testing.T) {
	t.Parallel()

	randReq := func() *translatev1.CreateServiceRequest {
		return &translatev1.CreateServiceRequest{
			Service: serviceToProto(randService()),
		}
	}

	happyReqWithServiceID := randReq()

	happyReqWithoutServiceID := randReq()
	happyReqWithoutServiceID.Service.Id = ""

	happyReqWithoutService := randReq()
	happyReqWithoutService.Service = nil

	malformedServiceIDReq := randReq()
	malformedServiceIDReq.Service.Id += "_FAIL"

	tests := []struct {
		request     *translatev1.CreateServiceRequest
		expectedErr *parseParamError
		expected    *createServiceParams
		name        string
	}{
		{
			name:        "Happy Path With Service ID",
			request:     happyReqWithServiceID,
			expectedErr: nil,
			expected: &createServiceParams{
				service: &model.Service{
					ID:   uuid.MustParse(happyReqWithServiceID.Service.Id),
					Name: happyReqWithServiceID.Service.Name,
				},
			},
		},
		{
			name:        "Happy Path Without Service ID",
			request:     happyReqWithoutServiceID,
			expectedErr: nil,
			expected: &createServiceParams{
				service: &model.Service{
					ID:   uuid.Nil,
					Name: happyReqWithoutServiceID.Service.Name,
				},
			},
		},
		{
			name:        "Happy Path Without Service",
			request:     happyReqWithoutService,
			expectedErr: nil,
			expected: &createServiceParams{
				service: nil,
			},
		},
		{
			name:        "Malformed Service ID",
			request:     malformedServiceIDReq,
			expectedErr: &parseParamError{field: "service.id"},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			actual, err := parseCreateServiceParams(tt.request)

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

func Test_ValidateCreateServiceParams(t *testing.T) {
	t.Parallel()

	randParams := func() *createServiceParams {
		return &createServiceParams{service: randService()}
	}

	happyParams := randParams()

	emptyServiceParams := randParams()
	emptyServiceParams.service = nil

	tests := []struct {
		params      *createServiceParams
		expectedErr error
		name        string
	}{
		{
			name:        "Happy Path",
			params:      happyParams,
			expectedErr: nil,
		},
		{
			name:        "Empty Service",
			params:      emptyServiceParams,
			expectedErr: errors.New("'service' is required"),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := validateCreateServiceParams(tt.params)

			if tt.expectedErr != nil {
				assert.ErrorContains(t, err, tt.expectedErr.Error())
				return
			}

			require.NoError(t, err)
		})
	}
}

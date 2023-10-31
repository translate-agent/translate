package server

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
		expectedErr error
		name        string
	}{
		{
			name:    "Happy Path",
			request: happyReqWithID,
			expected: &getServiceParams{
				id: uuid.MustParse(happyReqWithID.GetId()),
			},
			expectedErr: nil,
		},
		{
			name:    "Happy Path Empty ID",
			request: happyReqWithoutID,
			expected: &getServiceParams{
				id: uuid.Nil,
			},
		},
		{
			name:        "Malformed UUID",
			request:     malformedIDReq,
			expectedErr: errors.New("invalid UUID length"),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			actual, err := parseGetServiceRequestParams(tt.request)

			if tt.expectedErr != nil {
				require.ErrorContains(t, err, tt.expectedErr.Error())
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
		expectedErr error
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
			expectedErr: errors.New("'id' is required"),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.params.validate()

			if tt.expectedErr != nil {
				require.ErrorContains(t, err, tt.expectedErr.Error())
				return
			}

			require.NoError(t, err)
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
			UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"name"}},
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

	invalidUpdateMaskPathParams := randReq()
	invalidUpdateMaskPathParams.UpdateMask.Paths = gofakeit.NiceColors()

	tests := []struct {
		expected    *updateServiceParams
		expectedErr error
		request     *translatev1.UpdateServiceRequest
		name        string
	}{
		{
			name:    "Happy Path",
			request: happyReq,
			expected: &updateServiceParams{
				mask: happyReq.GetUpdateMask().GetPaths(),
				service: &model.Service{
					ID:   uuid.MustParse(happyReq.GetService().GetId()),
					Name: happyReq.GetService().GetName(),
				},
			},
			expectedErr: nil,
		},
		{
			name:    "Happy Path Without Service ID",
			request: happyReqWithoutServiceID,
			expected: &updateServiceParams{
				mask: happyReqWithoutServiceID.GetUpdateMask().GetPaths(),
				service: &model.Service{
					ID:   uuid.Nil,
					Name: happyReqWithoutServiceID.GetService().GetName(),
				},
			},
			expectedErr: nil,
		},

		{
			name:    "Happy Path Without Service",
			request: happyReqWithoutService,
			expected: &updateServiceParams{
				mask:    happyReqWithoutService.GetUpdateMask().GetPaths(),
				service: nil,
			},
		},
		{
			name:    "Happy Path Without Update Mask",
			request: happyReqWithoutUpdateMask,
			expected: &updateServiceParams{
				mask: nil,
				service: &model.Service{
					ID:   uuid.MustParse(happyReqWithoutUpdateMask.GetService().GetId()),
					Name: happyReqWithoutUpdateMask.GetService().GetName(),
				},
			},
		},
		{
			name:        "Malformed Service ID",
			request:     malformedIDReq,
			expectedErr: errors.New("invalid UUID length"),
		},
		{
			name:        "Invalid Update Mask Path",
			request:     invalidUpdateMaskPathParams,
			expectedErr: errors.New("invalid path"),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			actual, err := parseUpdateServiceParams(tt.request)

			if tt.expectedErr != nil {
				require.ErrorContains(t, err, tt.expectedErr.Error())
				return
			}

			require.NoError(t, err)

			assert.Equal(t, tt.expected, actual)
		})
	}
}

func Test_ValidateUpdateServiceParams(t *testing.T) {
	t.Parallel()

	randParams := func() *updateServiceParams {
		return &updateServiceParams{
			mask:    []string{"name"},
			service: randService(),
		}
	}

	happyParams := randParams()

	happyParamsMissingUpdateMask := randParams()
	happyParamsMissingUpdateMask.mask = nil

	missingServiceParams := randParams()
	missingServiceParams.service = nil

	// when updating a service, the service.ID is required and validation fails without it.
	missingServiceIDParams := randParams()
	missingServiceIDParams.service.ID = uuid.Nil

	tests := []struct {
		params      *updateServiceParams
		expectedErr error
		name        string
	}{
		{
			name:        "Happy Path",
			params:      happyParams,
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
			expectedErr: errors.New("'service' is required"),
		},
		{
			name:        "Missing Service ID",
			params:      missingServiceIDParams,
			expectedErr: errors.New("'service.id' is required"),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.params.validate()

			if tt.expectedErr != nil {
				require.ErrorContains(t, err, tt.expectedErr.Error())
				return
			}

			require.NoError(t, err)
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
			mask:    []string{"name"},
			service: &model.Service{ID: originalId, Name: gofakeit.Name()},
		}
	}

	updateNameField := randParams(originalService1.ID)

	updateAllFields := randParams(originalService2.ID)
	updateAllFields.mask = nil

	nothingToUpdate := randParams(originalService3.ID)
	nothingToUpdate.mask = []string{"random_path"}

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
		expectedErr error
		name        string
	}{
		{
			name:        "Happy Path With ID",
			request:     happyReqWithID,
			expected:    &deleteServiceParams{id: uuid.MustParse(happyReqWithID.GetId())},
			expectedErr: nil,
		},
		{
			name:        "Happy Path Without ID",
			request:     happyReqWithoutID,
			expected:    &deleteServiceParams{id: uuid.Nil},
			expectedErr: nil,
		},
		{
			name:        "Malformed UUID",
			request:     malformedIDReq,
			expectedErr: errors.New("invalid UUID length"),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			actual, err := parseDeleteServiceRequest(tt.request)

			if tt.expectedErr != nil {
				require.ErrorContains(t, err, tt.expectedErr.Error())
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

	emptyIDParams := randParams()
	emptyIDParams.id = uuid.Nil

	tests := []struct {
		params      *deleteServiceParams
		expectedErr error
		name        string
	}{
		{
			name:        "Happy Path",
			params:      happyParams,
			expectedErr: nil,
		},
		{
			name:        "Empty ID",
			params:      emptyIDParams,
			expectedErr: errors.New("'id' is required"),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.params.validate()

			if tt.expectedErr != nil {
				require.ErrorContains(t, err, tt.expectedErr.Error())
				return
			}

			require.NoError(t, err)
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
		expectedErr error
		expected    *createServiceParams
		name        string
	}{
		{
			name:        "Happy Path With Service ID",
			request:     happyReqWithServiceID,
			expectedErr: nil,
			expected: &createServiceParams{
				service: &model.Service{
					ID:   uuid.MustParse(happyReqWithServiceID.GetService().GetId()),
					Name: happyReqWithServiceID.GetService().GetName(),
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
					Name: happyReqWithoutServiceID.GetService().GetName(),
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
			expectedErr: errors.New("invalid UUID length"),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			actual, err := parseCreateServiceParams(tt.request)

			if tt.expectedErr != nil {
				require.ErrorContains(t, err, tt.expectedErr.Error())
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

	// when creating a service, the service.ID is optional and validation passes.
	happyParamsEmptyServiceID := randParams()
	happyParamsEmptyServiceID.service.ID = uuid.Nil

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
			name:        "Happy Path Empty Service ID",
			params:      happyParamsEmptyServiceID,
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

			err := tt.params.validate()

			if tt.expectedErr != nil {
				require.ErrorContains(t, err, tt.expectedErr.Error())
				return
			}

			require.NoError(t, err)
		})
	}
}

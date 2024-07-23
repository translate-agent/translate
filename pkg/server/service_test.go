package server

import (
	"errors"
	"reflect"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/google/uuid"
	"go.expect.digital/translate/pkg/model"
	translatev1 "go.expect.digital/translate/pkg/pb/translate/v1"
	"go.expect.digital/translate/pkg/testutil/expect"
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
		want    *getServiceParams
		request *translatev1.GetServiceRequest
		wantErr error
		name    string
	}{
		{
			name:    "Happy Path",
			request: happyReqWithID,
			want: &getServiceParams{
				id: uuid.MustParse(happyReqWithID.GetId()),
			},
			wantErr: nil,
		},
		{
			name:    "Happy Path Empty ID",
			request: happyReqWithoutID,
			want: &getServiceParams{
				id: uuid.Nil,
			},
		},
		{
			name:    "Malformed UUID",
			request: malformedIDReq,
			wantErr: errors.New("invalid UUID length"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			got, err := parseGetServiceRequestParams(test.request)

			if test.wantErr != nil {
				expect.ErrorContains(t, err, test.wantErr.Error())
				return
			}

			if err != nil {
				t.Error(err)
				return
			}

			if *test.want != *got {
				t.Errorf("want %v, got %v", test.want, got)
			}
		})
	}
}

func Test_ValidateGetServiceParams(t *testing.T) {
	t.Parallel()

	tests := []struct {
		params  *getServiceParams
		wantErr error
		name    string
	}{
		{
			name:    "Happy Path",
			wantErr: nil,
			params:  &getServiceParams{id: uuid.New()},
		},
		{
			name:    "Empty ID",
			params:  &getServiceParams{id: uuid.Nil},
			wantErr: errors.New("'id' is required"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			err := test.params.validate()

			if test.wantErr != nil {
				expect.ErrorContains(t, err, test.wantErr.Error())
				return
			}

			if err != nil {
				t.Error(err)
				return
			}
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
		want    *updateServiceParams
		wantErr error
		request *translatev1.UpdateServiceRequest
		name    string
	}{
		{
			name:    "Happy Path",
			request: happyReq,
			want: &updateServiceParams{
				mask: []string{"name"},
				service: &model.Service{
					ID:   uuid.MustParse(happyReq.GetService().GetId()),
					Name: happyReq.GetService().GetName(),
				},
			},
			wantErr: nil,
		},
		{
			name:    "Happy Path Without Service ID",
			request: happyReqWithoutServiceID,
			want: &updateServiceParams{
				mask: []string{"name"},
				service: &model.Service{
					ID:   uuid.Nil,
					Name: happyReqWithoutServiceID.GetService().GetName(),
				},
			},
			wantErr: nil,
		},

		{
			name:    "Happy Path Without Service",
			request: happyReqWithoutService,
			want: &updateServiceParams{
				mask:    []string{"name"},
				service: nil,
			},
		},
		{
			name:    "Happy Path Without Update Mask",
			request: happyReqWithoutUpdateMask,
			want: &updateServiceParams{
				mask: nil,
				service: &model.Service{
					ID:   uuid.MustParse(happyReqWithoutUpdateMask.GetService().GetId()),
					Name: happyReqWithoutUpdateMask.GetService().GetName(),
				},
			},
		},
		{
			name:    "Malformed Service ID",
			request: malformedIDReq,
			wantErr: errors.New("invalid UUID length"),
		},
		{
			name:    "Invalid Update Mask Path",
			request: invalidUpdateMaskPathParams,
			wantErr: errors.New("invalid path"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			got, err := parseUpdateServiceParams(test.request)

			if test.wantErr != nil {
				expect.ErrorContains(t, err, test.wantErr.Error())
				return
			}

			if err != nil {
				t.Error(err)
				return
			}

			if !reflect.DeepEqual(test.want, got) {
				t.Errorf("want %v, got %v", test.want, got)
			}
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
		params  *updateServiceParams
		wantErr error
		name    string
	}{
		{
			name:    "Happy Path",
			params:  happyParams,
			wantErr: nil,
		},
		{
			name:    "Happy Path Missing Update Mask",
			params:  happyParamsMissingUpdateMask,
			wantErr: nil,
		},
		{
			name:    "Missing Service",
			params:  missingServiceParams,
			wantErr: errors.New("'service' is required"),
		},
		{
			name:    "Missing Service ID",
			params:  missingServiceIDParams,
			wantErr: errors.New("'service.id' is required"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			err := test.params.validate()

			if test.wantErr != nil {
				expect.ErrorContains(t, err, test.wantErr.Error())
				return
			}

			if err != nil {
				t.Error(err)
			}
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
		want    *deleteServiceParams
		request *translatev1.DeleteServiceRequest
		wanterr error
		name    string
	}{
		{
			name:    "Happy Path With ID",
			request: happyReqWithID,
			want:    &deleteServiceParams{id: uuid.MustParse(happyReqWithID.GetId())},
			wanterr: nil,
		},
		{
			name:    "Happy Path Without ID",
			request: happyReqWithoutID,
			want:    &deleteServiceParams{id: uuid.Nil},
			wanterr: nil,
		},
		{
			name:    "Malformed UUID",
			request: malformedIDReq,
			wanterr: errors.New("invalid UUID length"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			got, err := parseDeleteServiceRequest(test.request)

			if test.wanterr != nil {
				expect.ErrorContains(t, err, test.wanterr.Error())
				return
			}

			if err != nil {
				t.Error(err)
				return
			}

			if !reflect.DeepEqual(test.want, got) {
				t.Errorf("want %v, got %v", test.want, got)
			}
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
		params  *deleteServiceParams
		wantErr error
		name    string
	}{
		{
			name:    "Happy Path",
			params:  happyParams,
			wantErr: nil,
		},
		{
			name:    "Empty ID",
			params:  emptyIDParams,
			wantErr: errors.New("'id' is required"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			err := test.params.validate()

			if test.wantErr != nil {
				expect.ErrorContains(t, err, test.wantErr.Error())
				return
			}

			if err != nil {
				t.Error(err)
			}
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
		request *translatev1.CreateServiceRequest
		wantErr error
		want    *createServiceParams
		name    string
	}{
		{
			name:    "Happy Path With Service ID",
			request: happyReqWithServiceID,
			wantErr: nil,
			want: &createServiceParams{
				service: &model.Service{
					ID:   uuid.MustParse(happyReqWithServiceID.GetService().GetId()),
					Name: happyReqWithServiceID.GetService().GetName(),
				},
			},
		},
		{
			name:    "Happy Path Without Service ID",
			request: happyReqWithoutServiceID,
			wantErr: nil,
			want: &createServiceParams{
				service: &model.Service{
					ID:   uuid.Nil,
					Name: happyReqWithoutServiceID.GetService().GetName(),
				},
			},
		},
		{
			name:    "Happy Path Without Service",
			request: happyReqWithoutService,
			wantErr: nil,
			want: &createServiceParams{
				service: nil,
			},
		},
		{
			name:    "Malformed Service ID",
			request: malformedServiceIDReq,
			wantErr: errors.New("invalid UUID length"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			got, err := parseCreateServiceParams(test.request)

			if test.wantErr != nil {
				expect.ErrorContains(t, err, test.wantErr.Error())
				return
			}

			if err != nil {
				t.Error(err)
				return
			}

			if !reflect.DeepEqual(test.want, got) {
				t.Errorf("want %v, got %v", test.want, got)
			}
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
		params  *createServiceParams
		wantErr error
		name    string
	}{
		{
			name:    "Happy Path",
			params:  happyParams,
			wantErr: nil,
		},
		{
			name:    "Happy Path Empty Service ID",
			params:  happyParamsEmptyServiceID,
			wantErr: nil,
		},
		{
			name:    "Empty Service",
			params:  emptyServiceParams,
			wantErr: errors.New("'service' is required"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			err := test.params.validate()

			if test.wantErr != nil {
				expect.ErrorContains(t, err, test.wantErr.Error())
				return
			}

			if err != nil {
				t.Error(err)
			}
		})
	}
}

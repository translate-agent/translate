package translate

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"golang.org/x/exp/slices"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/fieldmaskpb"

	"go.expect.digital/translate/pkg/model"
	translatev1 "go.expect.digital/translate/pkg/pb/translate/v1"
)

// ------------------------GetService-------------------------------

var errEmptyField = errors.New("must not be empty")

type getServiceParams struct {
	id uuid.UUID
}

func parseGetServiceRequestParams(req *translatev1.GetServiceRequest) (*getServiceParams, error) {
	id, err := uuidFromProto(req.GetId())
	if err != nil {
		return nil, &fieldViolationError{field: "id", err: err}
	}

	return &getServiceParams{id: id}, nil
}

func validateGetServiceRequestParams(params *getServiceParams) error {
	if params.id == uuid.Nil {
		return &fieldViolationError{field: "id", err: errEmptyField}
	}

	return nil
}

func (t *TranslateServiceServer) GetService(
	ctx context.Context,
	req *translatev1.GetServiceRequest,
) (*translatev1.Service, error) {
	params, err := parseGetServiceRequestParams(req)
	if err != nil {
		return nil, requestErrorToStatusErr(err)
	}

	if err = validateGetServiceRequestParams(params); err != nil {
		return nil, requestErrorToStatusErr(err)
	}

	service, err := t.repo.LoadService(ctx, params.id)
	if err != nil {
		return nil, repoErrorToStatusErr(err)
	}

	return serviceToProto(service), nil
}

// ----------------------ListServices-------------------------------

func (t *TranslateServiceServer) ListServices(
	ctx context.Context,
	req *translatev1.ListServicesRequest,
) (*translatev1.ListServicesResponse, error) {
	services, err := t.repo.LoadServices(ctx)
	if err != nil {
		return nil, repoErrorToStatusErr(err)
	}

	return &translatev1.ListServicesResponse{Services: servicesToProto(services)}, nil
}

// ---------------------CreateService-------------------------------

type createServiceParams struct {
	service *model.Service
}

func parseCreateServiceParams(req *translatev1.CreateServiceRequest) (*createServiceParams, error) {
	service, err := serviceFromProto(req.GetService())

	switch {
	case err == nil:
		return &createServiceParams{service: service}, nil
	case strings.Contains(err.Error(), "service id"):
		return nil, &fieldViolationError{field: "service.id", err: err}
	default:
		return nil, &fieldViolationError{field: "service", err: err}
	}
}

func validateCreateServiceParams(params *createServiceParams) error {
	if params.service == nil {
		return &fieldViolationError{field: "service", err: errEmptyField}
	}

	return nil
}

func (t *TranslateServiceServer) CreateService(
	ctx context.Context,
	req *translatev1.CreateServiceRequest,
) (*translatev1.Service, error) {
	params, err := parseCreateServiceParams(req)
	if err != nil {
		return nil, requestErrorToStatusErr(err)
	}

	if err := validateCreateServiceParams(params); err != nil {
		return nil, requestErrorToStatusErr(err)
	}

	if err := t.repo.SaveService(ctx, params.service); err != nil {
		return nil, repoErrorToStatusErr(err)
	}

	return serviceToProto(params.service), nil
}

// ---------------------UpdateService-------------------------------

// updateMaskAcceptablePaths is a list of acceptable paths for the Service update mask.
var updateMaskAcceptablePaths = []string{"name"}

type updateServiceParams struct {
	mask    *fieldmaskpb.FieldMask
	service *model.Service
}

func parseUpdateServiceParams(req *translatev1.UpdateServiceRequest) (*updateServiceParams, error) {
	service, err := serviceFromProto(req.GetService())

	switch {
	case err == nil:
		return &updateServiceParams{service: service, mask: req.GetUpdateMask()}, nil
	case strings.Contains(err.Error(), "service id"):
		return nil, &fieldViolationError{field: "service.id", err: err}
	default:
		return nil, &fieldViolationError{field: "service", err: err}
	}
}

func validateUpdateServiceParams(params *updateServiceParams) error {
	if params.service == nil {
		return &fieldViolationError{field: "service", err: errEmptyField}
	}

	if params.mask != nil {
		for _, path := range params.mask.Paths {
			if !slices.Contains(updateMaskAcceptablePaths, path) {
				return &fieldViolationError{field: "update_mask.paths", err: fmt.Errorf("'%s' is not an valid field", path)}
			}
		}
	}

	return nil
}

// updateServiceFromParams updates the service resource with the new values from the request.
func updateServiceFromParams(service *model.Service, reqParams *updateServiceParams) *model.Service {
	// Replace service resource with the new one from params (PUT)
	if reqParams.mask == nil {
		return &model.Service{ID: service.ID, Name: reqParams.service.Name}
	}

	updatedService := *service

	// Replace service resource's fields with the new ones from request (PATCH)
	for _, path := range reqParams.mask.Paths {
		switch path {
		default:
			// noop
		case "name":
			updatedService.Name = reqParams.service.Name
		}
	}

	return &updatedService
}

func (t *TranslateServiceServer) UpdateService(
	ctx context.Context,
	req *translatev1.UpdateServiceRequest,
) (*translatev1.Service, error) {
	params, err := parseUpdateServiceParams(req)
	if err != nil {
		return nil, requestErrorToStatusErr(err)
	}

	if err = validateUpdateServiceParams(params); err != nil {
		return nil, requestErrorToStatusErr(err)
	}

	oldService, err := t.repo.LoadService(ctx, params.service.ID)
	if err != nil {
		return nil, repoErrorToStatusErr(err)
	}

	updatedService := updateServiceFromParams(oldService, params)

	if err := t.repo.SaveService(ctx, updatedService); err != nil {
		return nil, repoErrorToStatusErr(err)
	}

	return serviceToProto(updatedService), nil
}

// ----------------------DeleteService------------------------------

type deleteServiceParams struct {
	id uuid.UUID
}

func parseDeleteServiceRequest(req *translatev1.DeleteServiceRequest) (*deleteServiceParams, error) {
	id, err := uuidFromProto(req.Id)
	if err != nil {
		return nil, &fieldViolationError{field: "id", err: err}
	}

	return &deleteServiceParams{id: id}, nil
}

func validateDeleteServiceParams(params *deleteServiceParams) error {
	if params.id == uuid.Nil {
		return &fieldViolationError{field: "id", err: errEmptyField}
	}

	return nil
}

func (t *TranslateServiceServer) DeleteService(
	ctx context.Context,
	req *translatev1.DeleteServiceRequest,
) (*emptypb.Empty, error) {
	params, err := parseDeleteServiceRequest(req)
	if err != nil {
		return nil, requestErrorToStatusErr(err)
	}

	if err := validateDeleteServiceParams(params); err != nil {
		return nil, requestErrorToStatusErr(err)
	}

	if err := t.repo.DeleteService(ctx, params.id); err != nil {
		return nil, repoErrorToStatusErr(err)
	}

	return &emptypb.Empty{}, nil
}

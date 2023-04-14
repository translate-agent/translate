package translate

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/fieldmaskpb"

	"go.expect.digital/translate/pkg/model"
	translatev1 "go.expect.digital/translate/pkg/pb/translate/v1"
)

type (
	getServiceRequest    translatev1.GetServiceRequest
	createServiceRequest translatev1.CreateServiceRequest
	updateServiceRequest translatev1.UpdateServiceRequest
	deleteServiceRequest translatev1.DeleteServiceRequest
)

// ------------------------GetService-------------------------------

type getServiceParams struct {
	id uuid.UUID
}

func (g *getServiceRequest) parseParams() (*getServiceParams, error) {
	if g == nil {
		return nil, errNilRequest
	}

	id, err := uuidFromProto(g.Id)
	if err != nil {
		return nil, &parseParamError{field: "id", err: err}
	}

	return &getServiceParams{id: id}, nil
}

func (t *TranslateServiceServer) GetService(
	ctx context.Context,
	req *translatev1.GetServiceRequest,
) (*translatev1.Service, error) {
	// getReq will be a pointer to the same underlying value as req, but as getReq type.
	getReq := (*getServiceRequest)(req)

	params, err := getReq.parseParams()
	if err != nil {
		return nil, requestErrorToStatus(err)
	}

	service, err := t.repo.LoadService(ctx, params.id)
	if err != nil {
		return nil, repoErrorToStatus(err)
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
		return nil, repoErrorToStatus(err)
	}

	return &translatev1.ListServicesResponse{Services: servicesToProto(services)}, nil
}

// ---------------------CreateService-------------------------------

type createServiceParams struct {
	service *model.Service
}

func (c *createServiceRequest) parseParams() (*createServiceParams, error) {
	if c == nil {
		return nil, errNilRequest
	}

	if c.Service == nil {
		return nil, errNilService
	}

	service, err := serviceFromProto(c.Service)

	switch {
	case err == nil:
		return &createServiceParams{service: service}, nil
	case strings.Contains(err.Error(), "service id"):
		return nil, &parseParamError{field: "service.id", err: err}
	default:
		return nil, &parseParamError{field: "service", err: err}
	}
}

func (t *TranslateServiceServer) CreateService(
	ctx context.Context,
	req *translatev1.CreateServiceRequest,
) (*translatev1.Service, error) {
	createReq := (*createServiceRequest)(req)

	params, err := createReq.parseParams()
	if err != nil {
		return nil, requestErrorToStatus(err)
	}

	if err := t.repo.SaveService(ctx, params.service); err != nil {
		return nil, repoErrorToStatus(err)
	}

	return serviceToProto(params.service), nil
}

// ---------------------UpdateService-------------------------------

type updateServiceParams struct {
	mask    *fieldmaskpb.FieldMask
	service *model.Service
}

func (u *updateServiceRequest) parseParams() (*updateServiceParams, error) {
	if u == nil {
		return nil, errNilRequest
	}

	if u.Service == nil {
		return nil, errNilService
	}

	service, err := serviceFromProto(u.Service)

	switch {
	case err == nil:
		return &updateServiceParams{service: service, mask: u.UpdateMask}, nil
	case strings.Contains(err.Error(), "service id"):
		return nil, &parseParamError{field: "service.id", err: err}
	default:
		return nil, &parseParamError{field: "service", err: err}
	}
}

func (u *updateServiceParams) updateServiceFromMask(service *model.Service) (*model.Service, error) {
	// Replace service resource with the new one from params (PUT)
	if u.mask == nil {
		return &model.Service{ID: service.ID, Name: u.service.Name}, nil
	}

	// Replace service resource's fields with the new ones from request (PATCH)
	for i, path := range u.mask.Paths {
		switch path {
		case "name":
			service.Name = u.service.Name
		default:
			return nil, &updateMaskError{field: fmt.Sprintf("update_mask.paths[%d]", i+1), value: path, entity: "Service"}
		}
	}

	return service, nil
}

func (t *TranslateServiceServer) UpdateService(
	ctx context.Context,
	req *translatev1.UpdateServiceRequest,
) (*translatev1.Service, error) {
	updateReq := (*updateServiceRequest)(req)

	params, err := updateReq.parseParams()
	if err != nil {
		return nil, requestErrorToStatus(err)
	}

	oldService, err := t.repo.LoadService(ctx, params.service.ID)
	if err != nil {
		return nil, repoErrorToStatus(err)
	}

	updatedService, err := params.updateServiceFromMask(oldService)
	if err != nil {
		return nil, requestErrorToStatus(err)
	}

	if err := t.repo.SaveService(ctx, updatedService); err != nil {
		return nil, repoErrorToStatus(err)
	}

	return serviceToProto(updatedService), nil
}

// ----------------------DeleteService------------------------------

type deleteServiceParams struct {
	id uuid.UUID
}

func (d *deleteServiceRequest) parseParams() (*deleteServiceParams, error) {
	if d == nil {
		return nil, errNilRequest
	}

	id, err := uuidFromProto(d.Id)
	if err != nil {
		return nil, &parseParamError{field: "id", err: err}
	}

	return &deleteServiceParams{id: id}, nil
}

func (t *TranslateServiceServer) DeleteService(
	ctx context.Context,
	req *translatev1.DeleteServiceRequest,
) (*emptypb.Empty, error) {
	deleteReq := (*deleteServiceRequest)(req)

	params, err := deleteReq.parseParams()
	if err != nil {
		return nil, requestErrorToStatus(err)
	}

	if err := t.repo.DeleteService(ctx, params.id); err != nil {
		return nil, repoErrorToStatus(err)
	}

	return &emptypb.Empty{}, nil
}

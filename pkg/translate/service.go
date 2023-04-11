package translate

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/fieldmaskpb"

	"go.expect.digital/translate/pkg/model"
	translatev1 "go.expect.digital/translate/pkg/pb/translate/v1"
	"go.expect.digital/translate/pkg/repo"
	"go.expect.digital/translate/pkg/transform"
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

func (g *getServiceRequest) parseParams() (getServiceParams, error) {
	if g == nil {
		return getServiceParams{}, errors.New("request is nil")
	}

	var (
		params getServiceParams
		err    error
	)

	params.id, err = uuid.Parse(g.Id)
	if err != nil {
		return getServiceParams{}, fmt.Errorf("parse uuid: %w", err)
	}

	return params, nil
}

func (t *TranslateServiceServer) GetService(
	ctx context.Context,
	req *translatev1.GetServiceRequest,
) (*translatev1.Service, error) {
	// getReq will be a pointer to the same underlying value as req, but as getReq type.
	getReq := (*getServiceRequest)(req)

	params, err := getReq.parseParams()
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	switch service, err := t.repo.LoadService(ctx, params.id); {
	default:
		return transform.ProtoServiceFromService(service), nil
	case errors.Is(err, repo.ErrNotFound):
		return nil, status.Errorf(codes.NotFound, err.Error())
	case err != nil:
		return nil, status.Errorf(codes.Internal, err.Error())
	}
}

// ----------------------ListServices-------------------------------

func (t *TranslateServiceServer) ListServices(
	ctx context.Context,
	req *translatev1.ListServicesRequest,
) (*translatev1.ListServicesResponse, error) {
	services, err := t.repo.LoadServices(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	resp := &translatev1.ListServicesResponse{
		Services: make([]*translatev1.Service, 0, len(services)),
	}

	for _, v := range services {
		service := v
		resp.Services = append(resp.Services, transform.ProtoServiceFromService(&service))
	}

	return resp, nil
}

// ---------------------CreateService-------------------------------

type createServiceParams struct {
	name string
	id   uuid.UUID
}

func (c *createServiceRequest) parseParams() (createServiceParams, error) {
	if c == nil {
		return createServiceParams{}, errors.New("request is nil")
	}

	if c.Service == nil {
		return createServiceParams{}, errors.New("service is nil")
	}

	params := createServiceParams{name: c.Service.Name}

	if c.Service.Id == "" {
		return params, nil
	}

	var err error

	params.id, err = uuid.Parse(c.Service.Id)
	if err != nil {
		return createServiceParams{}, fmt.Errorf("parse uuid: %w", err)
	}

	return params, nil
}

func (t *TranslateServiceServer) CreateService(
	ctx context.Context,
	req *translatev1.CreateServiceRequest,
) (*translatev1.Service, error) {
	createReq := (*createServiceRequest)(req)

	params, err := createReq.parseParams()
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	service := &model.Service{ID: params.id, Name: params.name}
	if err := t.repo.SaveService(ctx, service); err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	return transform.ProtoServiceFromService(service), nil
}

// ---------------------UpdateService-------------------------------

type updateServiceParams struct {
	mask *fieldmaskpb.FieldMask
	name string
	id   uuid.UUID
}

func (u *updateServiceRequest) parseParams() (updateServiceParams, error) {
	if u == nil {
		return updateServiceParams{}, errors.New("request is nil")
	}

	if u.Service == nil {
		return updateServiceParams{}, errors.New("service is nil")
	}

	params := updateServiceParams{name: u.Service.Name, mask: u.UpdateMask}

	var err error

	params.id, err = uuid.Parse(u.Service.Id)
	if err != nil {
		return updateServiceParams{}, fmt.Errorf("parse uuid: %w", err)
	}

	return params, nil
}

func (u *updateServiceParams) updateServiceFromMask(service *model.Service) (*model.Service, error) {
	// Replace service resource with the new one from params (PUT)
	if u.mask == nil {
		return &model.Service{ID: service.ID, Name: u.name}, nil
	}

	// Replace service resource's fields with the new ones from request (PATCH)
	for _, path := range u.mask.Paths {
		switch path {
		case "name":
			service.Name = u.name
		default:
			return nil, fmt.Errorf("'%s' is not a valid service field", path)
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
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	oldService, err := t.repo.LoadService(ctx, params.id)

	switch {
	case errors.Is(err, repo.ErrNotFound):
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	case err != nil:
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	updatedService, err := params.updateServiceFromMask(oldService)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	if err := t.repo.SaveService(ctx, updatedService); err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	return transform.ProtoServiceFromService(updatedService), nil
}

// ----------------------DeleteService------------------------------

type deleteServiceParams struct {
	id uuid.UUID
}

func (d *deleteServiceRequest) parseParams() (deleteServiceParams, error) {
	if d == nil {
		return deleteServiceParams{}, errors.New("request is nil")
	}

	var (
		params deleteServiceParams
		err    error
	)

	params.id, err = uuid.Parse(d.Id)
	if err != nil {
		return deleteServiceParams{}, fmt.Errorf("parse uuid: %w", err)
	}

	return params, nil
}

func (t *TranslateServiceServer) DeleteService(
	ctx context.Context,
	req *translatev1.DeleteServiceRequest,
) (*emptypb.Empty, error) {
	deleteReq := (*deleteServiceRequest)(req)

	params, err := deleteReq.parseParams()
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	switch err := t.repo.DeleteService(ctx, params.id); {
	default:
		return &emptypb.Empty{}, nil
	case errors.Is(err, repo.ErrNotFound):
		return nil, status.Errorf(codes.NotFound, err.Error())
	case err != nil:
		return nil, status.Errorf(codes.Internal, err.Error())
	}
}

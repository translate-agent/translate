package translate

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	"go.expect.digital/translate/pkg/model"
	translatev1 "go.expect.digital/translate/pkg/pb/translate/v1"
	"go.expect.digital/translate/pkg/repo"
)

type (
	getServiceRequest    translatev1.GetServiceRequest
	updateServiceRequest translatev1.UpdateServiceRequest
	deleteServiceRequest translatev1.DeleteServiceRequest
)

// protoServiceFromService converts model.Service -> translatev1.Service.
func protoServiceFromService(service *model.Service) *translatev1.Service {
	return &translatev1.Service{Id: service.ID.String(), Name: service.Name}
}

// ------------------------GetService-------------------------------

type getServiceParams struct {
	id uuid.UUID
}

func (g *getServiceRequest) parseParams() (getServiceParams, error) {
	if g == nil {
		return getServiceParams{}, errors.New("request is nil")
	}

	id, err := uuid.Parse(g.Id)
	if err != nil {
		return getServiceParams{}, fmt.Errorf("parse uuid: %w", err)
	}

	return getServiceParams{id: id}, nil
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
		return protoServiceFromService(service), nil
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
		resp.Services = append(resp.Services, protoServiceFromService(&service))
	}

	return resp, nil
}

// ---------------------UpdateService-------------------------------

type updateServiceParams struct {
	name string
	id   uuid.UUID
}

func (u *updateServiceRequest) parseParams() (updateServiceParams, error) {
	if u == nil {
		return updateServiceParams{}, errors.New("request is nil")
	}

	id, err := uuid.Parse(u.Service.Id)
	if err != nil {
		return updateServiceParams{}, fmt.Errorf("parse uuid: %w", err)
	}

	return updateServiceParams{id: id, name: u.Service.Name}, nil
}

func (u *updateServiceParams) validate() error {
	if u.name == "" {
		return errors.New("'name' is required")
	}

	return nil
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

	if err := params.validate(); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	service := model.Service{ID: params.id, Name: params.name}
	if err := t.repo.SaveService(ctx, &service); err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	return protoServiceFromService(&service), nil
}

// ----------------------DeleteService------------------------------

type deleteServiceParams struct {
	id uuid.UUID
}

func (d *deleteServiceRequest) parseParams() (deleteServiceParams, error) {
	if d == nil {
		return deleteServiceParams{}, errors.New("request is nil")
	}

	id, err := uuid.Parse(d.Id)
	if err != nil {
		return deleteServiceParams{}, fmt.Errorf("parse uuid: %w", err)
	}

	return deleteServiceParams{id: id}, nil
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

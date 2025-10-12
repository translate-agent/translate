package server

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"go.expect.digital/translate/pkg/model"
	translatev1 "go.expect.digital/translate/pkg/pb/translate/v1"
	"go.expect.digital/translate/pkg/repo"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// ------------------------GetService-------------------------------

type getServiceParams struct {
	id uuid.UUID
}

func parseGetServiceRequestParams(req *translatev1.GetServiceRequest) (*getServiceParams, error) {
	id, err := uuidFromProto(req.GetId())
	if err != nil {
		return nil, fmt.Errorf("parse id: %w", err)
	}

	return &getServiceParams{id: id}, nil
}

func (g *getServiceParams) validate() error {
	if g.id == uuid.Nil {
		return errors.New("'id' is required")
	}

	return nil
}

func (t *TranslateServiceServer) GetService(
	ctx context.Context,
	req *translatev1.GetServiceRequest,
) (*translatev1.Service, error) {
	params, err := parseGetServiceRequestParams(req)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	err = params.validate()
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	switch service, err := t.repo.LoadService(ctx, params.id); {
	default:
		return serviceToProto(service), nil
	case errors.Is(err, repo.ErrNotFound):
		return nil, status.Error(codes.NotFound, "service not found")
	case err != nil:
		return nil, status.Error(codes.Internal, "")
	}
}

// ----------------------ListServices-------------------------------

func (t *TranslateServiceServer) ListServices(
	ctx context.Context,
	_ *translatev1.ListServicesRequest,
) (*translatev1.ListServicesResponse, error) {
	services, err := t.repo.LoadServices(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, "")
	}

	return &translatev1.ListServicesResponse{Services: servicesToProto(services)}, nil
}

// ---------------------CreateService-------------------------------

type createServiceParams struct {
	service *model.Service
}

func parseCreateServiceParams(req *translatev1.CreateServiceRequest) (*createServiceParams, error) {
	service, err := serviceFromProto(req.GetService())
	if err != nil {
		return nil, fmt.Errorf("parse service: %w", err)
	}

	return &createServiceParams{service: service}, nil
}

func (c *createServiceParams) validate() error {
	if c.service == nil {
		return errors.New("'service' is required")
	}

	return nil
}

func (t *TranslateServiceServer) CreateService(
	ctx context.Context,
	req *translatev1.CreateServiceRequest,
) (*translatev1.Service, error) {
	params, err := parseCreateServiceParams(req)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	err = params.validate()
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	err = t.repo.SaveService(ctx, params.service)
	if err != nil {
		return nil, status.Error(codes.Internal, "")
	}

	return serviceToProto(params.service), nil
}

// ---------------------UpdateService-------------------------------

type updateServiceParams struct {
	service *model.Service
	mask    model.Mask
}

func parseUpdateServiceParams(req *translatev1.UpdateServiceRequest) (*updateServiceParams, error) {
	var (
		params updateServiceParams
		err    error

		reqService    = req.GetService()
		reqUpdateMask = req.GetUpdateMask()
	)

	// Parse service
	params.service, err = serviceFromProto(reqService)
	if err != nil {
		return nil, fmt.Errorf("parse service: %w", err)
	}

	params.mask, err = maskFromProto(reqService, reqUpdateMask)
	if err != nil {
		return nil, fmt.Errorf("parse field mask: %w", err)
	}

	return &params, nil
}

func (u *updateServiceParams) validate() error {
	if u.service == nil {
		return errors.New("'service' is required")
	}

	if u.service.ID == uuid.Nil {
		return errors.New("'service.id' is required")
	}

	return nil
}

func (t *TranslateServiceServer) UpdateService(
	ctx context.Context,
	req *translatev1.UpdateServiceRequest,
) (*translatev1.Service, error) {
	params, err := parseUpdateServiceParams(req)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	err = params.validate()
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	loadedService, err := t.repo.LoadService(ctx, params.service.ID)

	switch {
	case errors.Is(err, repo.ErrNotFound):
		return nil, status.Error(codes.NotFound, "service not found")
	case err != nil:
		return nil, status.Error(codes.Internal, "")
	}

	err = model.UpdateService(params.service, loadedService, params.mask)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	err = t.repo.SaveService(ctx, loadedService)
	if err != nil {
		return nil, status.Error(codes.Internal, "")
	}

	return serviceToProto(loadedService), nil
}

// ----------------------DeleteService------------------------------

type deleteServiceParams struct {
	id uuid.UUID
}

func parseDeleteServiceRequest(req *translatev1.DeleteServiceRequest) (*deleteServiceParams, error) {
	id, err := uuidFromProto(req.GetId())
	if err != nil {
		return nil, fmt.Errorf("parse id: %w", err)
	}

	return &deleteServiceParams{id: id}, nil
}

func (d *deleteServiceParams) validate() error {
	if d.id == uuid.Nil {
		return errors.New("'id' is required")
	}

	return nil
}

func (t *TranslateServiceServer) DeleteService(
	ctx context.Context,
	req *translatev1.DeleteServiceRequest,
) (*emptypb.Empty, error) {
	params, err := parseDeleteServiceRequest(req)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	err = params.validate()
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	switch err := t.repo.DeleteService(ctx, params.id); {
	default:
		return &emptypb.Empty{}, nil
	case errors.Is(err, repo.ErrNotFound):
		return nil, status.Error(codes.NotFound, "service not found")
	case err != nil:
		return nil, status.Error(codes.Internal, "")
	}
}

package translate

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	translatev1 "go.expect.digital/translate/pkg/pb/translate/v1"
)

type (
	getServiceRequest    translatev1.GetServiceRequest
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

	id, err := uuid.Parse(g.Id)
	if err != nil {
		return getServiceParams{}, fmt.Errorf("parse uuid: %w", err)
	}

	return getServiceParams{id: id}, nil
}

func (g *getServiceParams) validate() error {
	// validate if uuid is in DB
	return nil
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

	if err := params.validate(); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	return &translatev1.Service{}, nil
}

// ----------------------ListServices-------------------------------

func (t *TranslateServiceServer) ListServices(
	ctx context.Context,
	req *translatev1.ListServicesRequest,
) (*translatev1.ListServicesResponse, error) {
	return &translatev1.ListServicesResponse{}, nil
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

	return &translatev1.Service{}, nil
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

func (d *deleteServiceParams) validate() error {
	// validate if uuid is in DB
	return nil
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

	if err := params.validate(); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	return &emptypb.Empty{}, nil
}

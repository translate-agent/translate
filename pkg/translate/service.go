package translate

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	tpb "go.expect.digital/translate/pkg/pb/translate/v1"
)

type (
	getServiceRequest    tpb.GetServiceRequest
	updateServiceRequest tpb.UpdateServiceRequest
	deleteServiceRequest tpb.DeleteServiceRequest
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
	req *tpb.GetServiceRequest,
) (*tpb.Service, error) {
	// getReq will be a pointer to the same underlying value as req, but as getReq type.
	getReq := (*getServiceRequest)(req)

	params, err := getReq.parseParams()
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	if err := params.validate(); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	return &tpb.Service{}, nil
}

// ----------------------ListServices-------------------------------

func (t *TranslateServiceServer) ListServices(
	ctx context.Context,
	req *tpb.ListServicesRequest,
) (*tpb.ListServicesResponse, error) {
	return &tpb.ListServicesResponse{}, nil
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
	req *tpb.UpdateServiceRequest,
) (*tpb.Service, error) {
	updateReq := (*updateServiceRequest)(req)

	params, err := updateReq.parseParams()
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	if err := params.validate(); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	return &tpb.Service{}, nil
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
	req *tpb.DeleteServiceRequest,
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

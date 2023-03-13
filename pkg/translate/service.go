package translate

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	pb "go.expect.digital/translate/pkg/server/translate/v1"
)

type (
	loadServiceRequest   pb.LoadServiceRequest
	saveServiceRequest   pb.SaveServiceRequest
	deleteServiceRequest pb.DeleteServiceRequest
)

// ----------------------LoadService-------------------------------

type loadServiceParams struct {
	uuid uuid.UUID
}

func (l *loadServiceRequest) parseParams() (loadServiceParams, error) {
	if l == nil {
		return loadServiceParams{}, errors.New("request is nil")
	}

	uuid, err := uuid.Parse(l.Uuid)
	if err != nil {
		return loadServiceParams{}, fmt.Errorf("parse uuid: %w", err)
	}

	return loadServiceParams{uuid: uuid}, nil
}

func (l *loadServiceParams) validate() error {
	// validate if uuid is in DB
	return nil
}

func (t *TranslateServiceServer) LoadService(
	ctx context.Context,
	req *pb.LoadServiceRequest,
) (*pb.Service, error) {
	// loadReq will be a pointer to the same underlying value as req, but as loadReq type.
	loadReq := (*loadServiceRequest)(req)

	params, err := loadReq.parseParams()
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	if err := params.validate(); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	return &pb.Service{}, nil
}

// ----------------------LoadServices-------------------------------

func (t *TranslateServiceServer) LoadServices(
	ctx context.Context,
	req *pb.LoadServicesRequest,
) (*pb.LoadServicesResponse, error) {
	return &pb.LoadServicesResponse{}, nil
}

// -----------------------SaveService-------------------------------

type saveServiceParams struct {
	name string
	uuid uuid.UUID
}

func (s *saveServiceRequest) parseParams() (saveServiceParams, error) {
	if s == nil {
		return saveServiceParams{}, errors.New("request is nil")
	}

	uuid, err := uuid.Parse(s.Service.Uuid)
	if err != nil {
		return saveServiceParams{}, fmt.Errorf("parse uuid: %w", err)
	}

	return saveServiceParams{uuid: uuid, name: s.Service.Name}, nil
}

func (s *saveServiceParams) validate() error {
	if s.name == "" {
		return errors.New("'name' is required")
	}

	return nil
}

func (t *TranslateServiceServer) SaveService(
	ctx context.Context,
	req *pb.SaveServiceRequest,
) (*pb.Service, error) {
	saveReq := (*saveServiceRequest)(req)

	params, err := saveReq.parseParams()
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	if err := params.validate(); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	return &pb.Service{}, nil
}

// ----------------------DeleteService------------------------------

type deleteServiceParams struct {
	uuid uuid.UUID
}

func (d *deleteServiceRequest) parseParams() (deleteServiceParams, error) {
	if d == nil {
		return deleteServiceParams{}, errors.New("request is nil")
	}

	uuid, err := uuid.Parse(d.Uuid)
	if err != nil {
		return deleteServiceParams{}, fmt.Errorf("parse uuid: %w", err)
	}

	return deleteServiceParams{uuid: uuid}, nil
}

func (l *deleteServiceParams) validate() error {
	// validate if uuid is in DB
	return nil
}

func (t *TranslateServiceServer) DeleteService(
	ctx context.Context,
	req *pb.DeleteServiceRequest,
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

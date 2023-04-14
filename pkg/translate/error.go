package translate

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"strings"

	"go.expect.digital/translate/pkg/repo"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func getOriginalErr(err error) error {
	for {
		unwrappedErr := errors.Unwrap(err)
		if unwrappedErr == nil {
			break
		}

		err = unwrappedErr
	}

	return err
}

// ----------------------RequestRelatedErrors-----------------------------

var (
	errNilRequest = errors.New("request is nil")
	errNilService = errors.New("service is nil")
)

type parseParamError struct {
	err   error
	field string
}

func (p *parseParamError) Error() string {
	return fmt.Sprintf("parse %s: %s", p.field, p.err)
}

type updateMaskError struct {
	entity string
	field  string
	value  string
}

func (u *updateMaskError) Error() string {
	return fmt.Sprintf("'%s' is not valid field for %s", u.field, u.entity)
}

type validateParamError struct {
	param  string
	reason string
}

func (v *validateParamError) Error() string {
	return fmt.Sprintf("%s %s", v.param, v.reason)
}

func requestErrorToStatus(err error) error {
	var reqToErrStatus func() error

	var (
		parseParamErr    *parseParamError
		updateMaskErr    *updateMaskError
		validateParamErr *validateParamError
	)

	switch {
	case errors.Is(err, errNilRequest):
		reqToErrStatus = nilRequestErrStatus
	case errors.Is(err, errNilService):
		reqToErrStatus = nilServiceErrStatus
	case errors.As(err, &parseParamErr):
		reqToErrStatus = parseParamErr.status
	case errors.As(err, &updateMaskErr):
		reqToErrStatus = updateMaskErr.status
	case errors.As(err, &validateParamErr):
		reqToErrStatus = validateParamErr.status
	default:
		reqToErrStatus = func() error { return status.Errorf(codes.InvalidArgument, err.Error()) }
	}

	return reqToErrStatus()
}

func nilRequestErrStatus() error {
	return status.Errorf(codes.InvalidArgument, errNilRequest.Error())
}

func nilServiceErrStatus() error {
	return status.Errorf(codes.InvalidArgument, errNilService.Error())
}

func (p *parseParamError) status() error {
	st, err := status.Newf(
		codes.InvalidArgument,
		"Invalid %s",
		strings.Split(p.field, ".")[0],
	).WithDetails(
		&errdetails.BadRequest_FieldViolation{
			Field:       p.field,
			Description: getOriginalErr(p.err).Error(),
		},
	)
	// If we cannot construct error with details, return plain error.
	if err != nil {
		return status.Errorf(codes.InvalidArgument, p.Error())
	}

	return st.Err() //nolint:wrapcheck
}

func (u *updateMaskError) status() error {
	st, err := status.New(
		codes.InvalidArgument,
		"Invalid update mask",
	).WithDetails(
		&errdetails.BadRequest_FieldViolation{
			Field:       u.field,
			Description: getOriginalErr(u).Error(),
		},
	)
	// If we cannot construct error with details, return plain error.
	if err != nil {
		return status.Errorf(codes.InvalidArgument, u.Error())
	}

	return st.Err() //nolint:wrapcheck
}

func (v *validateParamError) status() error {
	st, err := status.Newf(
		codes.InvalidArgument,
		"Invalid %s",
		v.param,
	).WithDetails(
		&errdetails.BadRequest_FieldViolation{
			Field:       v.param,
			Description: getOriginalErr(v).Error(),
		},
	)
	// If we cannot construct error with details, return plain error.
	if err != nil {
		return status.Errorf(codes.InvalidArgument, v.Error())
	}

	return st.Err() //nolint:wrapcheck
}

// ----------------------RepoErrors------------------------------

func repoNotFoundErrStatus(repoErr *repo.NotFoundError) error {
	st, err := status.Newf(
		codes.NotFound,
		"%s does not exist",
		repoErr.Entity,
	).WithDetails(
		&errdetails.ErrorInfo{
			Reason: getOriginalErr(repoErr).Error(),
			Domain: "repo",
		},
	)
	if err != nil {
		return status.Errorf(codes.NotFound, repoErr.Error())
	}

	return st.Err() //nolint:wrapcheck
}

func repoDefaultErrStatus(repoErr *repo.DefaultError) error {
	st, err := status.Newf(
		codes.Internal,
		"Error while accessing %s",
		repoErr.Entity,
	).WithDetails(
		&errdetails.ErrorInfo{
			Reason: getOriginalErr(repoErr).Error(),
			Domain: "repo",
		},
	)
	if err != nil {
		return status.Errorf(codes.Internal, repoErr.Error())
	}

	return st.Err() //nolint:wrapcheck
}

func repoErrorToStatus(err error) error {
	var (
		repoDefaultErr  *repo.DefaultError
		repoNotFoundErr *repo.NotFoundError
	)

	switch {
	case errors.As(err, &repoNotFoundErr):
		return repoNotFoundErrStatus(repoNotFoundErr)
	case errors.As(err, &repoDefaultErr):
		return repoDefaultErrStatus(repoDefaultErr)
	default:
		return status.Errorf(codes.Internal, err.Error())
	}
}

// ----------------------ConvertErrors------------------------------

func convertErrorToStatus(err error) error {
	var (
		jsonSyntaxErr    *json.SyntaxError
		jsonUnmarshalErr *json.UnmarshalTypeError
		xmlSyntaxErr     *xml.SyntaxError
	)

	switch {
	case errors.As(err, &jsonSyntaxErr):
		return jsonSyntaxErrorToStatus(*jsonSyntaxErr)
	case errors.As(err, &jsonUnmarshalErr):
		return jsonUnmarshalTypeErrorToStatus(*jsonUnmarshalErr)
	case errors.As(err, &xmlSyntaxErr):
		return xmlSyntaxErrorToStatus(*xmlSyntaxErr)
	default:
		return status.Errorf(codes.InvalidArgument, err.Error())
	}
}

func jsonSyntaxErrorToStatus(syntaxErr json.SyntaxError) error {
	st, err := status.New(
		codes.InvalidArgument,
		"Invalid JSON",
	).WithDetails(
		&errdetails.BadRequest_FieldViolation{
			Field:       "data",
			Description: syntaxErr.Error(),
		},
	)
	if err != nil {
		return status.Errorf(codes.InvalidArgument, err.Error())
	}

	return st.Err() //nolint:wrapcheck
}

func jsonUnmarshalTypeErrorToStatus(unmarshalErr json.UnmarshalTypeError) error {
	st, err := status.Newf(
		codes.InvalidArgument,
		"Unmarshal '%s'",
		unmarshalErr.Field,
	).WithDetails(
		&errdetails.BadRequest_FieldViolation{
			Field:       "data",
			Description: unmarshalErr.Error(),
		},
	)
	if err != nil {
		return status.Errorf(codes.InvalidArgument, err.Error())
	}

	return st.Err() //nolint:wrapcheck
}

func xmlSyntaxErrorToStatus(syntaxErr xml.SyntaxError) error {
	st, err := status.New(
		codes.InvalidArgument,
		"Invalid XML",
	).WithDetails(
		&errdetails.BadRequest_FieldViolation{
			Field:       "data",
			Description: syntaxErr.Error(),
		},
	)
	if err != nil {
		return status.Errorf(codes.InvalidArgument, err.Error())
	}

	return st.Err() //nolint:wrapcheck
}

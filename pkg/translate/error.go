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

// getOriginalErr returns the original error from the error chain.
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

// parseParamError is an error that occurs when parsing request parameters.
type parseParamError struct {
	err   error
	field string
}

func (p *parseParamError) Error() string {
	return fmt.Sprintf("parse %s: %s", p.field, p.err)
}

// updateMaskError is an error that occurs when the updateMask contains a field which the entity does not have.
type updateMaskError struct {
	entity string
	field  string
}

func (u *updateMaskError) Error() string {
	return fmt.Sprintf("'%s' is not valid field for %s", u.field, u.entity)
}

// validateParamError is an error that occurs when validating request parameters.
type validateParamError struct {
	param  string
	reason string
}

func (v *validateParamError) Error() string {
	return fmt.Sprintf("%s %s", v.param, v.reason)
}

func nilRequestErrStatus() error {
	return status.Errorf(codes.InvalidArgument, errNilRequest.Error())
}

func nilServiceErrStatus() error {
	return status.Errorf(codes.InvalidArgument, errNilService.Error())
}

// parseParamsError.status returns gRPC error status with details about failed parameter.
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

// updateMaskError.status returns gRPC error status with details about failed updateMask.
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

// validateParamError.status returns gRPC error status with details about failed validation.
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

// requestErrorToStatus converts request-related error to gRPC error status.
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

// ----------------------RepoErrors------------------------------

// repoNotFoundErrStatus converts repo.NotFoundError to gRPC error status.
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

// repoDefaultErrStatus converts repo.DefaultError to gRPC error status.
func repoDefaultErrStatus(repoErr *repo.DefaultError) error {
	st, err := status.Newf(
		codes.Internal,
		"Cannot access %s",
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

// repoErrorToStatus converts repo-related error to gRPC error status.
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

// convertError is an error that occurs when converting ToMessages or FromMessages fails.
type convertError struct {
	schema string
	err    error
	field  string
}

func (c *convertError) Error() string {
	return fmt.Sprintf("convert: %s", c.err.Error())
}

// convertFromErrorToStatus converts convertError to gRPC status.
// This function assumes that convertError is always caused by user providing malformed data.
func convertFromErrorToStatus(convertErr *convertError) error {
	var (
		jsonSyntaxErr *json.SyntaxError
		xmlSyntaxErr  *xml.SyntaxError
	)

	var errMsg string

	switch {
	case errors.As(convertErr.err, &jsonSyntaxErr):
		errMsg = "Invalid JSON"
	case errors.As(convertErr.err, &xmlSyntaxErr):
		errMsg = "Invalid XML"
	default:
		errMsg = fmt.Sprintf("Cannot convert from %s schema", convertErr.schema)
	}

	st, err := status.New(
		codes.InvalidArgument,
		errMsg,
	).WithDetails(
		&errdetails.BadRequest_FieldViolation{
			Field:       convertErr.field,
			Description: getOriginalErr(convertErr.err).Error(),
		},
	)
	if err != nil {
		return status.Errorf(codes.InvalidArgument, err.Error())
	}

	return st.Err() //nolint:wrapcheck
}

// convertToErrorToStatus converts convertError to gRPC status.
// This function assumes that convertError is always caused by server inability
// to convert messages to requested schema.
func convertToErrorToStatus(convertError *convertError) error {
	st, err := status.Newf(
		codes.Internal,
		"Cannot convert to %s schema",
		convertError.schema,
	).WithDetails(
		&errdetails.ErrorInfo{
			Reason: getOriginalErr(convertError.err).Error(),
			Domain: "convert",
			Metadata: map[string]string{
				"full_error": convertError.Error(),
			},
		},
	)
	if err != nil {
		return status.Errorf(codes.Internal, err.Error())
	}

	return st.Err() //nolint:wrapcheck
}

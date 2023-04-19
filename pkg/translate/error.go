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
	"google.golang.org/protobuf/runtime/protoiface"
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

// newStatusWithDetails creates a new gRPC error status with details.
func newStatusWithDetails(code codes.Code, msg string, details ...protoiface.MessageV1) (*status.Status, error) {
	stWithDetails, err := status.New(code, msg).WithDetails(details...)
	if err != nil {
		return nil, fmt.Errorf("add details to status: %w", err)
	}

	return stWithDetails, nil
}

// ----------------------RequestRelatedErrors-----------------------------

// fieldViolationError occurs when a request field does not pass validation or could not be parsed.
type fieldViolationError struct {
	err   error
	field string
}

func (f *fieldViolationError) Error() string {
	return fmt.Sprintf("field '%s': %s", f.field, f.err)
}

// requestErrorToStatusErr converts a request error to a gRPC error status.
//
// NOTE:
// For now, it only supports field validation errors. In the future, it can support
// other types of request errors. E.g. UNAUTHENTICATED, PERMISSION_DENIED, etc.
func requestErrorToStatusErr(reqErr error) error {
	var fieldErr *fieldViolationError

	var (
		code    codes.Code
		msg     string
		details protoiface.MessageV1
	)

	switch {
	case errors.As(reqErr, &fieldErr):
		code = codes.InvalidArgument
		msg = fmt.Sprintf("Invalid %s", strings.Split(fieldErr.field, ".")[0])
		details = &errdetails.BadRequest_FieldViolation{
			Field:       fieldErr.field,
			Description: getOriginalErr(fieldErr.err).Error(),
		}
	default:
		code = codes.Unknown
		msg = getOriginalErr(reqErr).Error()
	}

	st, err := newStatusWithDetails(code, msg, details)
	// If newStatusWithDetails returns an error, it means that the details cannot be marshalled.
	// In this case, we return the original error.
	if err != nil {
		return status.Errorf(code, getOriginalErr(reqErr).Error())
	}

	return st.Err() //nolint:wrapcheck
}

// ----------------------RepoErrors------------------------------

// repoErrorToStatusErr converts repo-related error to gRPC error status.
func repoErrorToStatusErr(repoErr error) error {
	var (
		repoDefaultErr  *repo.DefaultError
		repoNotFoundErr *repo.NotFoundError
	)

	var (
		code    codes.Code
		msg     string
		details protoiface.MessageV1
	)

	switch {
	case errors.As(repoErr, &repoNotFoundErr):
		code = codes.NotFound
		msg = fmt.Sprintf("%s does not exist", repoNotFoundErr.Entity)
		details = &errdetails.ErrorInfo{
			Reason: getOriginalErr(repoNotFoundErr).Error(),
			Domain: "repo",
		}
	case errors.As(repoErr, &repoDefaultErr):
		code = codes.Internal
		msg = fmt.Sprintf("Cannot access %s", repoDefaultErr.Entity)
		details = &errdetails.ErrorInfo{
			Domain: "repo",
		}
	default:
		code = codes.Unknown
		msg = "Unknown error"
	}

	st, err := newStatusWithDetails(code, msg, details)
	// If newStatusWithDetails returns an error, it means that the details cannot be marshalled.
	// In this case, we return the original error.
	if err != nil {
		return status.Errorf(code, msg)
	}

	return st.Err() //nolint:wrapcheck
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

// convertFromErrorToStatusErr converts convertError to gRPC status.
// This function assumes that convertError is always caused by user providing malformed data.
func convertFromErrorToStatusErr(convertErr *convertError) error {
	var (
		jsonSyntaxErr *json.SyntaxError
		xmlSyntaxErr  *xml.SyntaxError
	)

	var (
		errMsg string
		code   = codes.InvalidArgument
	)

	switch {
	case errors.As(convertErr.err, &jsonSyntaxErr):
		errMsg = "Invalid JSON"
	case errors.As(convertErr.err, &xmlSyntaxErr):
		errMsg = "Invalid XML"
	default:
		errMsg = fmt.Sprintf("Cannot convert from %s schema", convertErr.schema)
	}

	st, err := newStatusWithDetails(
		code,
		errMsg,
		&errdetails.BadRequest_FieldViolation{
			Field:       convertErr.field,
			Description: getOriginalErr(convertErr.err).Error(),
		},
	)
	if err != nil {
		return status.Errorf(code, getOriginalErr(convertErr.err).Error())
	}

	return st.Err() //nolint:wrapcheck
}

// convertToErrorToStatusErr converts convertError to gRPC status.
// This function assumes that convertError is always caused by server inability
// to convert messages to requested schema.
func convertToErrorToStatusErr(convertError *convertError) error {
	st, err := newStatusWithDetails(
		codes.Internal,
		fmt.Sprintf("Cannot convert to %s schema", convertError.schema),
		&errdetails.ErrorInfo{
			Reason: getOriginalErr(convertError.err).Error(),
			Domain: "convert",
		})
	if err != nil {
		return status.Errorf(codes.Internal, getOriginalErr(convertError.err).Error())
	}

	return st.Err() //nolint:wrapcheck
}

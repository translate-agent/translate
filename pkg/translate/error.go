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
	"google.golang.org/protobuf/proto"
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
func newStatusWithDetails(code codes.Code, msg string, details ...proto.Message) (*status.Status, error) {
	st := status.New(code, msg)

	if len(details) == 0 {
		return st, nil
	}

	// Convert details to protoiface.MessageV1.
	v1Details := make([]protoiface.MessageV1, len(details))

	for i, detail := range details {
		if detail == nil {
			continue
		}

		v1Details[i] = detail.(protoiface.MessageV1)
	}

	stWithDetails, err := st.WithDetails(v1Details...)
	if err != nil {
		return nil, fmt.Errorf("append details: %w", err)
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

// requestErrorToStatus converts a request error to a gRPC error status.
//
// NOTE:
// For now, it only supports field validation errors. In the future, it can support
// other types of request errors. E.g. UNAUTHENTICATED, PERMISSION_DENIED, etc.
func requestErrorToStatus(reqErr error) error {
	var fieldErr *fieldViolationError

	var (
		code    codes.Code
		msg     string
		details proto.Message
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

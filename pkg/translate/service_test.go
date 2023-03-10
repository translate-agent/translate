package translate

import (
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	pb "go.expect.digital/translate/pkg/server/translate/v1"
)

// ----------------------GetService Parse Params-------------------------------

func Test_ParseGetServiceParams(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input       *pb.GetServiceRequest
		expectedErr error
		name        string
		expected    getServiceParams
	}{
		{
			name: "Happy Path",
			input: &pb.GetServiceRequest{
				Id: "599e59b8-3f2b-430f-baf7-8f837f7343a1",
			},
			expected: getServiceParams{
				id: uuid.MustParse("599e59b8-3f2b-430f-baf7-8f837f7343a1"),
			},
			expectedErr: nil,
		},
		{
			name: "Malformed UUID",
			input: &pb.GetServiceRequest{
				Id: "599e59b8-3f2b-430f-baf7-failTest",
			},
			expectedErr: errors.New("invalid UUID format"),
		},
		{
			name:        "NIL request",
			input:       nil,
			expectedErr: errors.New("request is nil"),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := (*getServiceRequest)(tt.input)

			actual, err := req.parseParams()

			if tt.expectedErr != nil {
				assert.ErrorContains(t, err, tt.expectedErr.Error())
				return
			}

			if !assert.NoError(t, err) {
				return
			}

			assert.Equal(t, tt.expected, actual)
		})
	}
}

// -----------------------UpdateService Parse Params-------------------------------

func Test_ParseUpdateServiceParams(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input       *pb.UpdateServiceRequest
		expectedErr error
		name        string
		expected    updateServiceParams
	}{
		{
			name: "Happy Path",
			input: &pb.UpdateServiceRequest{
				Service: &pb.Service{
					Id:   "599e59b8-3f2b-430f-baf7-8f837f7343a1",
					Name: "first service",
				},
			},
			expected: updateServiceParams{
				id:   uuid.MustParse("599e59b8-3f2b-430f-baf7-8f837f7343a1"),
				name: "first service",
			},
			expectedErr: nil,
		},
		{
			name: "Malformed UUID",
			input: &pb.UpdateServiceRequest{
				Service: &pb.Service{
					Id:   "599e59b8-3f2b-430f-baf7-failTest",
					Name: "failed service",
				},
			},
			expectedErr: errors.New("invalid UUID format"),
		},
		{
			name:        "NIL request",
			input:       nil,
			expectedErr: errors.New("request is nil"),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := (*updateServiceRequest)(tt.input)

			actual, err := req.parseParams()

			if tt.expectedErr != nil {
				assert.ErrorContains(t, err, tt.expectedErr.Error())
				return
			}

			if !assert.NoError(t, err) {
				return
			}

			assert.Equal(t, tt.expected, actual)
		})
	}
}

// ----------------------DeleteService Parse Params------------------------------

func Test_ParseDeleteServiceParams(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input       *pb.DeleteServiceRequest
		expectedErr error
		name        string
		expected    deleteServiceParams
	}{
		{
			name: "Happy Path",
			input: &pb.DeleteServiceRequest{
				Id: "599e59b8-3f2b-430f-baf7-8f837f7343a1",
			},
			expected: deleteServiceParams{
				id: uuid.MustParse("599e59b8-3f2b-430f-baf7-8f837f7343a1"),
			},
			expectedErr: nil,
		},
		{
			name: "Malformed UUID",
			input: &pb.DeleteServiceRequest{
				Id: "599e59b8-3f2b-430f-baf7-failTest",
			},
			expectedErr: errors.New("invalid UUID format"),
		},
		{
			name:        "NIL request",
			input:       nil,
			expectedErr: errors.New("request is nil"),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := (*deleteServiceRequest)(tt.input)

			actual, err := req.parseParams()

			if tt.expectedErr != nil {
				assert.ErrorContains(t, err, tt.expectedErr.Error())
				return
			}

			if !assert.NoError(t, err) {
				return
			}

			assert.Equal(t, tt.expected, actual)
		})
	}
}

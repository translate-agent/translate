package translate

import (
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	pb "go.expect.digital/translate/pkg/server/translate/v1"
)

// ----------------------LoadService Parse Params-------------------------------

func Test_ParseLoadServiceParams(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input       *pb.LoadServiceRequest
		expectedErr error
		name        string
		expected    loadServiceParams
	}{
		{
			name: "Happy Path",
			input: &pb.LoadServiceRequest{
				Uuid: "599e59b8-3f2b-430f-baf7-8f837f7343a1",
			},
			expected: loadServiceParams{
				uuid: uuid.MustParse("599e59b8-3f2b-430f-baf7-8f837f7343a1"),
			},
			expectedErr: nil,
		},
		{
			name: "Malformed UUID",
			input: &pb.LoadServiceRequest{
				Uuid: "599e59b8-3f2b-430f-baf7-failTest",
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

			req := (*loadServiceRequest)(tt.input)

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

// -----------------------SaveService Parse Params-------------------------------

func Test_ParseSaveServiceParams(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input       *pb.SaveServiceRequest
		expectedErr error
		name        string
		expected    saveServiceParams
	}{
		{
			name: "Happy Path",
			input: &pb.SaveServiceRequest{
				Service: &pb.Service{
					Uuid: "599e59b8-3f2b-430f-baf7-8f837f7343a1",
					Name: "first service",
				},
			},
			expected: saveServiceParams{
				uuid: uuid.MustParse("599e59b8-3f2b-430f-baf7-8f837f7343a1"),
				name: "first service",
			},
			expectedErr: nil,
		},
		{
			name: "Malformed UUID",
			input: &pb.SaveServiceRequest{
				Service: &pb.Service{
					Uuid: "599e59b8-3f2b-430f-baf7-failTest",
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

			req := (*saveServiceRequest)(tt.input)

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
				Uuid: "599e59b8-3f2b-430f-baf7-8f837f7343a1",
			},
			expected: deleteServiceParams{
				uuid: uuid.MustParse("599e59b8-3f2b-430f-baf7-8f837f7343a1"),
			},
			expectedErr: nil,
		},
		{
			name: "Malformed UUID",
			input: &pb.DeleteServiceRequest{
				Uuid: "599e59b8-3f2b-430f-baf7-failTest",
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

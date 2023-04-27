package translate

import (
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.expect.digital/translate/pkg/model"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

func Test_UpdateModelFromFieldMask(t *testing.T) {
	t.Parallel()

	//nolint: govet
	type s struct {
		A int    `protoName:"A"`
		B string `protoName:"B"`
		C struct {
			D float32 `protoName:"D"`
			E string
			F struct {
				G []string `protoName:"G"`
			} `protoName:"F"`
			H []struct {
				I string `protoName:"I"`
			} `protoName:"H"`
		} `protoName:"C"`
	}

	// Generate random source and destination structs
	src, dst := &s{}, &s{}
	require.NoError(t, gofakeit.Struct(src))
	require.NoError(t, gofakeit.Struct(dst))

	tests := []struct {
		mask       *fieldmaskpb.FieldMask
		assertFunc func(t *testing.T, dst, src, result *s)
		name       string
	}{
		{
			// Update one top-level field
			name: "Update A int",
			mask: &fieldmaskpb.FieldMask{Paths: []string{"A"}},
			assertFunc: func(t *testing.T, dst, src, result *s) {
				// Check if field is updated
				require.Equal(t, src.A, result.A)

				// Reset field to dst value, and perform full check, to ensure that nothing else was changed
				result.A = dst.A
				assert.Equal(t, dst, result)
			},
		},
		{
			// Update two top-level fields
			name: "Update A and B int and string",
			mask: &fieldmaskpb.FieldMask{Paths: []string{"A", "B"}},
			assertFunc: func(t *testing.T, dst, src, result *s) {
				require.Equal(t, src.A, result.A)

				result.A, result.B = dst.A, dst.B
				assert.Equal(t, dst, result)
			},
		},
		{
			// Update whole top-level struct
			name: "Update C struct",
			mask: &fieldmaskpb.FieldMask{Paths: []string{"C"}},
			assertFunc: func(t *testing.T, dst, src, result *s) {
				require.Equal(t, src.C, result.C)

				result.C = dst.C
				assert.Equal(t, dst, result)
			},
		},
		{
			// Update top-level field of a nested struct
			name: "Update C.D struct.float",
			mask: &fieldmaskpb.FieldMask{Paths: []string{"C.D"}},
			assertFunc: func(t *testing.T, dst, src, result *s) {
				require.Equal(t, src.C.D, result.C.D)

				result.C.D = dst.C.D
				assert.Equal(t, dst, result)
			},
		},
		{
			// Update a whole nested struct
			name: "Update C.F struct.struct",
			mask: &fieldmaskpb.FieldMask{Paths: []string{"C.F"}},
			assertFunc: func(t *testing.T, dst, src, result *s) {
				require.Equal(t, src.C.F, result.C.F)

				result.C.F = dst.C.F
				assert.Equal(t, dst, result)
			},
		},
		{
			// Update slice of strings in a double nested struct (merge two slices)
			name: "Update C.F.G struct.struct.[]string",
			mask: &fieldmaskpb.FieldMask{Paths: []string{"C.F.G"}},
			assertFunc: func(t *testing.T, dst, src, result *s) {
				// Merge slices
				la := len(dst.C.F.G)
				merged := make([]string, la+len(src.C.F.G))
				copy(merged, dst.C.F.G)
				copy(merged[la:], src.C.F.G)

				require.ElementsMatch(t, merged, result.C.F.G)

				result.C.F.G = dst.C.F.G
				assert.Equal(t, dst, result)
			},
		},
		{
			// Update slice of custom structs in a nested struct. (Merge two slices)
			name: "Update C.H struct.struct.[]struct",
			mask: &fieldmaskpb.FieldMask{Paths: []string{"C.H"}},
			assertFunc: func(t *testing.T, dst, src, result *s) {
				// Merge slices
				la := len(dst.C.H)
				merged := make([]struct {
					I string `protoName:"I"`
				}, la+len(src.C.H))
				copy(merged, dst.C.H)
				copy(merged[la:], src.C.H)

				require.ElementsMatch(t, merged, result.C.H)

				result.C.H = dst.C.H
				assert.Equal(t, dst, result)
			},
		},
		{
			// Try to update nested struct field with no protoName. (Nothing updates)
			name: "Try to Update C.E struct.string no protoName",
			mask: &fieldmaskpb.FieldMask{Paths: []string{"C.E"}},
			assertFunc: func(t *testing.T, dst, src, result *s) {
				require.Equal(t, dst.C.E, result.C.E)
				assert.Equal(t, dst, result)
			},
		},
		{
			// Update all fields
			name: "Update All",
			mask: nil,
			assertFunc: func(t *testing.T, dst, src, result *s) {
				assert.Equal(t, src, result)
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := updateModelFromFieldMask(tt.mask, dst, src)
			tt.assertFunc(t, dst, src, result)
		})
	}
}

func Test_UpdateServiceFromFieldMask(t *testing.T) {
	t.Parallel()

	tests := []struct {
		assertFunc func(t *testing.T, targetService, sourceService, updatedService *model.Service)
		fieldMask  *fieldmaskpb.FieldMask
		name       string
	}{
		{
			name:      "Update Name",
			fieldMask: &fieldmaskpb.FieldMask{Paths: []string{"name"}},
			assertFunc: func(t *testing.T, targetService, sourceService, updatedService *model.Service) {
				// Same ID updated name
				require.Equal(t, targetService.ID, updatedService.ID)
				assert.Equal(t, sourceService.Name, updatedService.Name)
			},
		},
		{
			name:      "Update All",
			fieldMask: nil,
			assertFunc: func(t *testing.T, targetService, sourceService, updatedService *model.Service) {
				// Same ID updated name, as ID cannot be updated, and service has only two fields.
				require.Equal(t, targetService.ID, updatedService.ID)
				assert.Equal(t, sourceService.Name, updatedService.Name)
			},
		},
		{
			name:      "Nothing to Update",
			fieldMask: &fieldmaskpb.FieldMask{},
			assertFunc: func(t *testing.T, targetService, sourceService, updatedService *model.Service) {
				// Same ID and name, as nothing was updated
				require.Equal(t, targetService.ID, updatedService.ID)
				assert.Equal(t, targetService.Name, updatedService.Name)
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			targetService := &model.Service{ID: uuid.New(), Name: gofakeit.FirstName()}
			sourceService := &model.Service{ID: uuid.New(), Name: gofakeit.FirstName()}

			updatedService := updateServiceFromFieldMask(tt.fieldMask, targetService, sourceService)

			tt.assertFunc(t, targetService, sourceService, updatedService)
		})
	}
}

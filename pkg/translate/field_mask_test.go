package translate

import (
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.expect.digital/translate/pkg/model"
	"golang.org/x/exp/maps"
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
		J struct {
			K map[string]string `protoName:"K"`
		} `protoName:"J"`
		L *string `protoName:"L"`
	}

	// Generate random source and destination structs
	var src, dst s

	require.NoError(t, gofakeit.Struct(&src))
	require.NoError(t, gofakeit.Struct(&dst))

	// In real life updateModelFromFieldMask is not supposed to be called directly.
	// It should be wrapped in a function that accepts source and destination structs,
	// Then function will be pure and will not have side effects.
	// e.g.
	updateSFromFieldMask := func(m *fieldmaskpb.FieldMask, v1, v2 s) *s {
		return updateModelFromFieldMask(m, &v1, &v2)
	}

	tests := []struct {
		mask       *fieldmaskpb.FieldMask
		assertFunc func(t *testing.T, dst, src, result s)
		name       string
	}{
		{
			// Update one top-level field
			name: "Update A int",
			mask: &fieldmaskpb.FieldMask{Paths: []string{"A"}},
			assertFunc: func(t *testing.T, dst, src, result s) {
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
			assertFunc: func(t *testing.T, dst, src, result s) {
				require.Equal(t, src.A, result.A)

				result.A, result.B = dst.A, dst.B
				assert.Equal(t, dst, result)
			},
		},
		{
			// Update whole top-level struct
			name: "Update C struct",
			mask: &fieldmaskpb.FieldMask{Paths: []string{"C"}},
			assertFunc: func(t *testing.T, dst, src, result s) {
				require.Equal(t, src.C, result.C)

				result.C = dst.C
				assert.Equal(t, dst, result)
			},
		},
		{
			// Update top-level field of a nested struct
			name: "Update C.D struct.float",
			mask: &fieldmaskpb.FieldMask{Paths: []string{"C.D"}},
			assertFunc: func(t *testing.T, dst, src, result s) {
				require.Equal(t, src.C.D, result.C.D)

				result.C.D = dst.C.D
				assert.Equal(t, dst, result)
			},
		},
		{
			// Update a whole nested struct
			name: "Update C.F struct.struct",
			mask: &fieldmaskpb.FieldMask{Paths: []string{"C.F"}},
			assertFunc: func(t *testing.T, dst, src, result s) {
				require.Equal(t, src.C.F, result.C.F)

				result.C.F = dst.C.F
				assert.Equal(t, dst, result)
			},
		},
		{
			// Update slice of strings in a double nested struct (merge two slices)
			name: "Update C.F.G struct.struct.[]string",
			mask: &fieldmaskpb.FieldMask{Paths: []string{"C.F.G"}},
			assertFunc: func(t *testing.T, dst, src, result s) {
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
			assertFunc: func(t *testing.T, dst, src, result s) {
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
			// Update map of strings in a nested field of struct. (Merge two maps)
			name: "Update J.K struct.map[string]string",
			mask: &fieldmaskpb.FieldMask{Paths: []string{"J.K"}},
			assertFunc: func(t *testing.T, dst, src, result s) {
				// Merge maps
				merged := make(map[string]string, len(dst.J.K)+len(src.J.K))
				maps.Copy(merged, dst.J.K)
				maps.Copy(merged, src.J.K)

				require.Equal(t, merged, result.J.K)

				result.J.K = dst.J.K
				assert.Equal(t, dst, result)
			},
		},
		{
			// Try to update nested struct field with no protoName. (Nothing updates)
			name: "Try to Update C.E struct.string no protoName",
			mask: &fieldmaskpb.FieldMask{Paths: []string{"C.E"}},
			assertFunc: func(t *testing.T, dst, src, result s) {
				assert.Equal(t, dst, result)
			},
		},
		{
			// Update top level pointer to string
			name: "Update L *string",
			mask: &fieldmaskpb.FieldMask{Paths: []string{"L"}},
			assertFunc: func(t *testing.T, dst, src, result s) {
				require.Equal(t, src.L, result.L)

				result.L = dst.L
				assert.Equal(t, dst, result)
			},
		},
		{
			// Update all fields
			name: "Update All",
			mask: nil,
			assertFunc: func(t *testing.T, dst, src, result s) {
				assert.Equal(t, src, result)
			},
		},
		{
			// No Paths in FieldMask. Updates nothing.
			name: "Update Nothing",
			mask: &fieldmaskpb.FieldMask{},
			assertFunc: func(t *testing.T, dst, src, result s) {
				assert.Equal(t, dst, result)
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := updateSFromFieldMask(tt.mask, dst, src)
			tt.assertFunc(t, dst, src, *result)
		})
	}
}

func Test_UpdateServiceFromFieldMask(t *testing.T) {
	t.Parallel()

	tests := []struct {
		assertFunc func(t *testing.T, dstService, srcService, result model.Service)
		fieldMask  *fieldmaskpb.FieldMask
		name       string
	}{
		{
			name:      "Update Name",
			fieldMask: &fieldmaskpb.FieldMask{Paths: []string{"name"}},
			assertFunc: func(t *testing.T, dstService, srcService, result model.Service) {
				// Same ID updated name
				require.Equal(t, dstService.ID, result.ID)
				assert.Equal(t, srcService.Name, result.Name)
			},
		},
		{
			name:      "Update All",
			fieldMask: nil,
			assertFunc: func(t *testing.T, dstService, srcService, result model.Service) {
				// Same ID updated name, as ID cannot be updated, and service has only two fields.
				require.Equal(t, dstService.ID, result.ID)
				assert.Equal(t, srcService.Name, result.Name)
			},
		},
		{
			name:      "Nothing to Update",
			fieldMask: &fieldmaskpb.FieldMask{},
			assertFunc: func(t *testing.T, dstService, srcService, result model.Service) {
				// Same ID and name, as nothing was updated
				require.Equal(t, dstService.ID, result.ID)
				assert.Equal(t, dstService.Name, result.Name)
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			dstService := model.Service{ID: uuid.New(), Name: gofakeit.FirstName()}
			srcService := model.Service{ID: uuid.New(), Name: gofakeit.FirstName()}

			result := updateServiceFromFieldMask(tt.fieldMask, dstService, srcService)

			tt.assertFunc(t, dstService, srcService, *result)
		})
	}
}

package translate

import (
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.expect.digital/translate/pkg/model"
	"golang.org/x/exp/maps"
)

func mergeSlices[T ~[]E, E any](a, b T) T {
	merged := make(T, len(a)+len(b))
	copy(merged, a)
	copy(merged[len(a):], b)

	return merged
}

func mergeMaps[T ~map[K]V, K comparable, V any](a, b T) T {
	merged := make(T, len(a)+len(b))
	maps.Copy(merged, a)
	maps.Copy(merged, b)

	return merged
}

func Test_UpdateModelFromFieldMask(t *testing.T) {
	t.Parallel()

	//nolint: govet
	type nestedStruct struct {
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
	var src, dst nestedStruct

	require.NoError(t, gofakeit.Struct(&src))
	require.NoError(t, gofakeit.Struct(&dst))

	// updateFromMask is not supposed to be called directly.
	// It should be wrapped in a function that accepts source and destination structs,
	// Then function will be pure and will not have side effects.
	// e.g.
	updateNestedStructFromMask := func(src, dst nestedStruct, m model.Mask) *nestedStruct {
		return updateFromMask(&src, &dst, m)
	}

	tests := []struct {
		assertFunc func(t *testing.T, src, dst, result nestedStruct)
		name       string
		mask       model.Mask
	}{
		{
			// Update one top-level field
			name: "Update A int",
			mask: []string{"A"},
			assertFunc: func(t *testing.T, src, dst, result nestedStruct) {
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
			mask: []string{"A", "B"},
			assertFunc: func(t *testing.T, src, dst, result nestedStruct) {
				require.Equal(t, src.A, result.A)

				result.A, result.B = dst.A, dst.B
				assert.Equal(t, dst, result)
			},
		},
		{
			// Update whole top-level struct
			name: "Update C struct",
			mask: []string{"C"},
			assertFunc: func(t *testing.T, src, dst, result nestedStruct) {
				require.Equal(t, src.C, result.C)

				result.C = dst.C
				assert.Equal(t, dst, result)
			},
		},
		{
			// Update top-level field of a nested struct
			name: "Update C.D struct.float",
			mask: []string{"C.D"},
			assertFunc: func(t *testing.T, src, dst, result nestedStruct) {
				require.Equal(t, src.C.D, result.C.D)

				result.C.D = dst.C.D
				assert.Equal(t, dst, result)
			},
		},
		{
			// Update a whole nested struct
			name: "Update C.F struct.struct",
			mask: []string{"C.F"},
			assertFunc: func(t *testing.T, src, dst, result nestedStruct) {
				require.Equal(t, src.C.F, result.C.F)

				result.C.F = dst.C.F
				assert.Equal(t, dst, result)
			},
		},
		{
			// Update slice of strings in a double nested struct (merge two slices)
			name: "Update C.F.G struct.struct.[]string",
			mask: []string{"C.F.G"},
			assertFunc: func(t *testing.T, src, dst, result nestedStruct) {
				merged := mergeSlices(dst.C.F.G, src.C.F.G)

				require.ElementsMatch(t, merged, result.C.F.G)

				result.C.F.G = dst.C.F.G
				assert.Equal(t, dst, result)
			},
		},
		{
			// Update slice of custom structs in a nested struct. (Merge two slices)
			name: "Update C.H struct.struct.[]struct",
			mask: []string{"C.H"},
			assertFunc: func(t *testing.T, src, dst, result nestedStruct) {
				merged := mergeSlices(dst.C.H, src.C.H)

				require.ElementsMatch(t, merged, result.C.H)

				result.C.H = dst.C.H
				assert.Equal(t, dst, result)
			},
		},
		{
			// Update map of strings in a nested field of struct. (Merge two maps)
			name: "Update J.K struct.map[string]string",
			mask: []string{"J.K"},
			assertFunc: func(t *testing.T, src, dst, result nestedStruct) {
				merged := mergeMaps(dst.J.K, src.J.K)

				require.Equal(t, merged, result.J.K)

				result.J.K = dst.J.K
				assert.Equal(t, dst, result)
			},
		},
		{
			// Try to update nested struct field with no protoName. (Nothing updates)
			name: "Try to Update C.E struct.string no protoName",
			mask: []string{"C.E"},
			assertFunc: func(t *testing.T, src, dst, result nestedStruct) {
				assert.Equal(t, dst, result)
			},
		},
		{
			// Update top level pointer to string
			name: "Update L *string",
			mask: []string{"L"},
			assertFunc: func(t *testing.T, src, dst, result nestedStruct) {
				require.Equal(t, src.L, result.L)

				result.L = dst.L
				assert.Equal(t, dst, result)
			},
		},
		{
			// Update all fields
			name: "Update All",
			mask: nil,
			assertFunc: func(t *testing.T, src, dst, result nestedStruct) {
				assert.Equal(t, src, result)
			},
		},
		{
			// No Paths in FieldMask. Updates nothing.
			name: "Update Nothing Empty Paths",
			mask: model.Mask{},
			assertFunc: func(t *testing.T, src, dst, result nestedStruct) {
				assert.Equal(t, dst, result)
			},
		},
		{
			// Random path in FieldMask. Updates nothing.
			name: "Update Nothing Random Path",
			mask: model.Mask{"random_path"},
			assertFunc: func(t *testing.T, src, dst, result nestedStruct) {
				assert.Equal(t, dst, result)
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := updateNestedStructFromMask(src, dst, tt.mask)
			tt.assertFunc(t, src, dst, *result)
		})
	}
}

func Test_UpdateServiceFromFieldMask(t *testing.T) {
	t.Parallel()

	// Generate random source and destination structs
	var srcService, dstService model.Service

	require.NoError(t, gofakeit.Struct(&srcService))
	require.NoError(t, gofakeit.Struct(&dstService))

	tests := []struct {
		assertFunc func(t *testing.T, srcService, dstService, result model.Service)
		name       string
		fieldMask  model.Mask
	}{
		{
			name:      "Update Name",
			fieldMask: []string{"name"},
			assertFunc: func(t *testing.T, srcService, dstService, result model.Service) {
				// Same ID updated name
				require.Equal(t, dstService.ID, result.ID)
				assert.Equal(t, srcService.Name, result.Name)
			},
		},
		{
			name:      "Update All",
			fieldMask: nil,
			assertFunc: func(t *testing.T, srcService, dstService, result model.Service) {
				// Same ID updated name, as ID cannot be updated, and service has only two fields.
				require.Equal(t, dstService.ID, result.ID)
				assert.Equal(t, srcService.Name, result.Name)
			},
		},
		{
			name:      "Nothing to Update Empty Paths",
			fieldMask: model.Mask{},
			assertFunc: func(t *testing.T, srcService, dstService, result model.Service) {
				// Same ID and name, as nothing was updated
				assert.Equal(t, dstService, result)
			},
		},
		{
			name:      "Nothing to Update Random Path",
			fieldMask: model.Mask{"random_path"},
			assertFunc: func(t *testing.T, srcService, dstService, result model.Service) {
				// Same ID and name, as nothing was updated
				assert.Equal(t, dstService, result)
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := updateServiceFromMask(srcService, dstService, tt.fieldMask)
			tt.assertFunc(t, srcService, dstService, *result)
		})
	}
}

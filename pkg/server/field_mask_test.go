package server

import (
	"encoding/json"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.expect.digital/translate/pkg/model"
)

// deepCopy makes a deep copy of src and returns it.
func deepCopy[T any](t *testing.T, src T) (dst T) { //nolint
	t.Helper()

	data, err := json.Marshal(src)
	require.NoError(t, err)

	require.NoError(t, json.Unmarshal(data, &dst))

	return
}

func Test_UpdateNestedStructFromMask(t *testing.T) {
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

	tests := []struct {
		assertFunc func(t *testing.T, src, dst, original nestedStruct)
		name       string
		mask       model.Mask
	}{
		{
			// Update one top-level field
			name: "Update A int",
			mask: []string{"A"},
			assertFunc: func(t *testing.T, src, dst, original nestedStruct) {
				// Check if field is updated
				require.Equal(t, src.A, dst.A)

				// Reset field to original value, and perform full check, to ensure that nothing else was changed
				dst.A = original.A
				assert.Equal(t, original, dst)
			},
		},
		{
			// Update two top-level fields
			name: "Update A and B int and string",
			mask: []string{"A", "B"},
			assertFunc: func(t *testing.T, src, dst, original nestedStruct) {
				require.Equal(t, src.A, dst.A)

				dst.A, dst.B = original.A, original.B
				assert.Equal(t, original, dst)
			},
		},
		{
			// Update whole top-level struct
			name: "Update C struct",
			mask: []string{"C"},
			assertFunc: func(t *testing.T, src, dst, original nestedStruct) {
				require.Equal(t, src.C, dst.C)

				dst.C = original.C
				assert.Equal(t, original, dst)
			},
		},
		{
			// Update top-level field of a nested struct
			name: "Update C.D struct.float",
			mask: []string{"C.D"},
			assertFunc: func(t *testing.T, src, dst, original nestedStruct) {
				require.InDelta(t, src.C.D, dst.C.D, 0.01)

				dst.C.D = original.C.D
				assert.Equal(t, original, dst)
			},
		},
		{
			// Update a whole nested struct
			name: "Update C.F struct.struct",
			mask: []string{"C.F"},
			assertFunc: func(t *testing.T, src, dst, original nestedStruct) {
				require.Equal(t, src.C.F, dst.C.F)

				dst.C.F = original.C.F
				assert.Equal(t, original, dst)
			},
		},
		{
			// Update slice of strings in a double nested struct (merge two slices)
			name: "Update C.F.G struct.struct.[]string",
			mask: []string{"C.F.G"},
			assertFunc: func(t *testing.T, src, dst, original nestedStruct) {
				// Check if all elements from src and dst are in result
				for _, srcElem := range src.C.F.G {
					require.Contains(t, dst.C.F.G, srcElem)
				}

				for _, dstElem := range dst.C.F.G {
					require.Contains(t, dst.C.F.G, dstElem)
				}

				dst.C.F.G = original.C.F.G
				assert.Equal(t, original, dst)
			},
		},
		{
			// Update slice of custom structs in a nested struct. (Merge two slices)
			name: "Update C.H struct.struct.[]struct",
			mask: []string{"C.H"},
			assertFunc: func(t *testing.T, src, dst, original nestedStruct) {
				// Check if all elements from src and dst are in result
				for _, srcElem := range src.C.H {
					require.Contains(t, dst.C.H, srcElem)
				}

				for _, dstElem := range dst.C.H {
					require.Contains(t, dst.C.H, dstElem)
				}

				dst.C.H = original.C.H
				assert.Equal(t, original, dst)
			},
		},
		{
			// Update map of strings in a nested field of struct. (Merge two maps)
			name: "Update J.K struct.map[string]string",
			mask: []string{"J.K"},
			assertFunc: func(t *testing.T, src, dst, original nestedStruct) {
				// Check if all keys from src and dst are in result
				for srcKey := range src.J.K {
					require.Contains(t, dst.J.K, srcKey)
				}

				for dstKey := range dst.J.K {
					require.Contains(t, dst.J.K, dstKey)
				}

				dst.J.K = original.J.K
				assert.Equal(t, original, dst)
			},
		},
		{
			// Try to update nested struct field with no protoName. (Nothing updates)
			name: "Try to Update C.E struct.string no protoName",
			mask: []string{"C.E"},
			assertFunc: func(t *testing.T, src, dst, original nestedStruct) {
				assert.Equal(t, original, dst)
			},
		},
		{
			// Update top level pointer to string
			name: "Update L *string",
			mask: []string{"L"},
			assertFunc: func(t *testing.T, src, dst, original nestedStruct) {
				require.Equal(t, src.L, dst.L)

				dst.L = original.L
				assert.Equal(t, original, dst)
			},
		},
		{
			// Update all fields
			name: "Update All",
			mask: nil,
			assertFunc: func(t *testing.T, src, dst, _ nestedStruct) {
				assert.Equal(t, src, dst)
			},
		},
		{
			// No Paths in FieldMask. Updates nothing.
			name: "Update Nothing Empty Paths",
			mask: model.Mask{},
			assertFunc: func(t *testing.T, _, dst, original nestedStruct) {
				assert.Equal(t, original, dst)
			},
		},
		{
			// Random path in FieldMask. Updates nothing.
			name: "Update Nothing Random Path",
			mask: model.Mask{"random_path"},
			assertFunc: func(t *testing.T, _, dst, original nestedStruct) {
				assert.Equal(t, original, dst)
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Make deep copies of structs for deterministic tests
			original := deepCopy(t, dst)
			dstCopy, srcCopy := deepCopy(t, dst), deepCopy(t, src)

			updateFromMask(&srcCopy, &dstCopy, tt.mask)
			tt.assertFunc(t, srcCopy, dstCopy, original)
		})
	}
}

func Test_UpdateServiceFromMask(t *testing.T) {
	t.Parallel()

	// Generate random source and destination structs
	var srcService, dstService model.Service

	require.NoError(t, gofakeit.Struct(&srcService))
	require.NoError(t, gofakeit.Struct(&dstService))

	tests := []struct {
		assertFunc func(t *testing.T, srcService, dstService, original model.Service)
		name       string
		fieldMask  model.Mask
	}{
		{
			name:      "Update Name",
			fieldMask: model.Mask{"name"},
			assertFunc: func(t *testing.T, srcService, dstService, original model.Service) {
				// Same ID updated name
				require.Equal(t, original.ID, dstService.ID)
				assert.Equal(t, srcService.Name, dstService.Name)
			},
		},
		{
			name:      "Update All",
			fieldMask: nil,
			assertFunc: func(t *testing.T, srcService, dstService, original model.Service) {
				// Same ID updated name, as ID cannot be updated, and service has only two fields.
				require.Equal(t, original.ID, dstService.ID)
				assert.Equal(t, srcService.Name, dstService.Name)
			},
		},
		{
			name:      "Nothing to Update Empty Paths",
			fieldMask: model.Mask{},
			assertFunc: func(t *testing.T, _, dstService, original model.Service) {
				// Same ID and name, as nothing was updated
				assert.Equal(t, dstService, original)
			},
		},
		{
			name:      "Nothing to Update Random Path",
			fieldMask: model.Mask{"random_path"},
			assertFunc: func(t *testing.T, _, dstService, original model.Service) {
				// Same ID and name, as nothing was updated
				assert.Equal(t, dstService, original)
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Make deep copies of structs for deterministic tests
			original := deepCopy(t, dstService)
			dstCopy, srcCopy := deepCopy(t, dstService), deepCopy(t, srcService)

			updateServiceFromMask(&srcCopy, &dstCopy, tt.fieldMask)
			tt.assertFunc(t, srcCopy, dstCopy, original)
		})
	}
}
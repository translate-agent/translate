package translate

import (
	"fmt"
	"reflect"
	"strings"

	"go.expect.digital/translate/pkg/model"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

// parseFieldMask parses and normalizes a field mask from a proto message and a list of paths.
func parseFieldMask(message proto.Message, paths []string) (model.Mask, error) {
	parsedMask, err := fieldmaskpb.New(message, paths...)
	if err != nil {
		return nil, fmt.Errorf("new fieldmaskpb: %w", err)
	}

	// Normalize sorts paths, removes duplicates, and removes sub-paths when possible.
	// e.g. if a field mask contains the paths foo.bar and foo,
	// the path foo.bar is redundant because it is already covered by the path foo
	parsedMask.Normalize()

	return parsedMask.Paths, nil
}

// updateModelFromFieldMask updates the dst with the values from src based on the fieldMask.
//
// Following scenarios are possible:
//   - fieldMask is nil: All fields are updated.
//   - fieldMask is not nil, but has no paths: No fields are updated.
//   - fieldMask contains paths, that does not exist in the model: The paths are ignored.
//   - fieldMask contains paths, that exist in model: Only those fields are updated.
//
// `protoName` tags are used to match fields from the fieldMask to fields in the model.
func updateModelFromFieldMask[T any](dst, src *T, mask model.Mask) *T {
	// If fieldMask is nil, update all fields
	if mask == nil {
		return src
	}

	for _, path := range mask {
		fields := strings.Split(path, ".")
		updateField(fields, reflect.ValueOf(dst).Elem(), reflect.ValueOf(src).Elem())
	}

	return dst
}

// updateField updates the dst value with the values from src, based on the fields slice.
func updateField(fields []string, dst, src reflect.Value) {
	if len(fields) == 0 {
		return
	}

	field := fields[0]

	for i := 0; i < dst.NumField(); i++ {
		tag := dst.Type().Field(i).Tag.Get("protoName")

		if field != tag {
			continue
		}

		dstField, srcField := dst.Field(i), src.Field(i)

		switch dst.Field(i).Kind() { //nolint:exhaustive
		case reflect.Struct:
			// If the field is a struct, and fields contains any sub-fields of a struct, recursively update the struct
			// If fields contains only 1 element, that means that the struct itself should be updated
			if len(fields) > 1 {
				updateField(fields[1:], dstField, srcField)
			} else {
				dstField.Set(srcField)
			}

		case reflect.Slice:
			//nolint:lll
			// If the field is a slice, append new values from src to existing slice in dst
			// https://github.com/protocolbuffers/protobuf/blob/9bbea4aa65bdaf5fc6c2583e045c07ff37ffb0e7/src/google/protobuf/field_mask.proto#L111
			dstField.Set(reflect.AppendSlice(dstField, srcField))

		case reflect.Map:
			// Same rule for maps as for slices
			for _, key := range srcField.MapKeys() {
				dstField.SetMapIndex(key, srcField.MapIndex(key))
			}

		default:
			// For all other field kinds, set value of field in dst to value of corresponding field in src
			dstField.Set(srcField)
		}
	}
}

// updateServiceFromFieldMask updates the dstService with the values from srcService based on the fieldMask.
func updateServiceFromFieldMask(
	dstService model.Service,
	srcService model.Service,
	mask model.Mask,
) *model.Service {
	// Set the ID of the srcService to the ID of the dstService, to prevent the ID from being updated
	srcService.ID = dstService.ID

	return updateModelFromFieldMask(&dstService, &srcService, mask)
}

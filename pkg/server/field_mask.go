package server

import (
	"fmt"
	"reflect"
	"strings"

	"go.expect.digital/translate/pkg/model"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

// parseFieldMask parses the field mask from the request and
// returns a model mask with removed duplicates and sorted paths.
func parseFieldMask(message proto.Message, paths []string) (model.Mask, error) {
	protoMask, err := fieldmaskpb.New(message, paths...)
	if err != nil {
		return nil, fmt.Errorf("new fieldmaskpb: %w", err)
	}

	// Normalize sorts paths, removes duplicates, and removes sub-paths when possible.
	// e.g. if a field mask contains the paths foo.bar and foo,
	// the path foo.bar is redundant because it is already covered by the path foo
	protoMask.Normalize()

	// Convert the proto mask to a model mask
	parsed, err := maskFromProto(message, protoMask)
	if err != nil {
		return nil, fmt.Errorf("transform proto mask to model mask: %w", err)
	}

	return parsed, nil
}

// updateFromMask updates the dst with the values from src based on the mask.
//
// Following scenarios are possible:
//   - mask is nil: All fields are updated.
//   - mask is not nil, but has no paths: No fields are updated.
//   - mask contains paths, that does not exist in the model: The paths are ignored.
//   - mask contains paths, that exist in model: Only those fields are updated.
//
// Example:
//
//	type Foo struct {
//		 Bar string
//		 Baz string
//	}
//
//	dst := Foo{Bar: "bar", Baz: "baz"}
//	src := Foo{Bar: "bar2", Baz: "baz2"}
//	mask := model.Mask{"Bar"}
//
//	updateFromMask(&src, &dst, mask)
//	fmt.Println(dst) // Foo{Bar: "bar2", Baz: "baz"}.
func updateFromMask[T any](src, dst *T, mask model.Mask) {
	// If mask is nil, update all fields
	if mask == nil {
		*dst = *src

		return
	}

	for _, path := range mask {
		fields := strings.Split(path, ".")
		updateField(reflect.ValueOf(src).Elem(), reflect.ValueOf(dst).Elem(), fields)
	}
}

// updateField updates the dst value with the values from src, based on the fields slice.
func updateField(src, dst reflect.Value, fields []string) {
	if len(fields) == 0 {
		return
	}

	field := fields[0]

	for i := 0; i < dst.NumField(); i++ {
		// Find corresponding field in dst
		if field != dst.Type().Field(i).Name {
			continue
		}

		srcField, dstField := src.Field(i), dst.Field(i)

		switch dst.Field(i).Kind() { //nolint:exhaustive
		case reflect.Struct:
			// If the field is a struct, and fields contains any sub-fields of a struct, recursively update the struct
			// If fields contains only 1 element, that means that the struct itself should be updated
			if len(fields) > 1 {
				updateField(srcField, dstField, fields[1:])
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

		return
	}
}

// updateServiceFromMask updates the dstService with the values from srcService based on the fieldMask.
func updateServiceFromMask(
	srcService *model.Service,
	dstService *model.Service,
	mask model.Mask,
) {
	// Set the ID of the srcService to the ID of the dstService, to prevent the ID
	// from being updated, when mask is nil or "ID" is in the mask
	srcService.ID = dstService.ID
	updateFromMask(srcService, dstService, mask)
}

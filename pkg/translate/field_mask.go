package translate

import (
	"reflect"
	"strings"

	"go.expect.digital/translate/pkg/model"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

// updateModelFromFieldMask updates the dst with the values from src based on the fieldMask.
// If fieldMask is nil, all fields are updated.
//
// `protoName` tags are used to match fields from the fieldMask to fields in the model.
func updateModelFromFieldMask[T any](fieldMask *fieldmaskpb.FieldMask, dst, src *T) *T {
	// If fieldMask is nil, update all fields
	if fieldMask == nil {
		return src
	}

	for _, path := range fieldMask.GetPaths() {
		fields := strings.Split(path, ".")
		updateField(fields, reflect.ValueOf(dst).Elem(), reflect.ValueOf(src).Elem())
	}

	return dst
}

// updateField updates the dst value with the values from src, based on the fields slice.
func updateField(fields []string, dst, src reflect.Value) {
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
	fieldMask *fieldmaskpb.FieldMask,
	dstService model.Service,
	srcService model.Service,
) *model.Service {
	// Set the ID of the srcService to the ID of the dstService, to prevent the ID from being updated
	srcService.ID = dstService.ID

	return updateModelFromFieldMask(fieldMask, &dstService, &srcService)
}

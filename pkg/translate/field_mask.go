package translate

import (
	"reflect"
	"strings"

	"go.expect.digital/translate/pkg/model"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

// updateModelFromFieldMask updates the targetModel with the values from updateModel based on the fieldMask.
// It returns a new *T value that is a copy of targetModel with the updates applied.
func updateModelFromFieldMask[T any](fieldMask *fieldmaskpb.FieldMask, dst, src *T) *T {
	// Create a new value that is a copy of targetModel
	v := reflect.ValueOf(dst).Elem()
	result := reflect.New(v.Type()).Elem()
	result.Set(v)

	// If fieldMask is nil, update all fields
	if fieldMask == nil {
		return src
	}

	for _, path := range fieldMask.GetPaths() {
		fields := strings.Split(path, ".")
		updateField(fields, result, reflect.ValueOf(src).Elem())
	}

	return result.Addr().Interface().(*T)
}

// updateField updates the targetValue with the values from updateValue based on the fields slice.
func updateField(fields []string, dst, src reflect.Value) {
	if len(fields) == 0 {
		return
	}

	for i := 0; i < dst.NumField(); i++ {
		tag := dst.Type().Field(i).Tag.Get("protoName")
		field := fields[0]

		if field != tag {
			continue
		}

		switch {
		case dst.Field(i).Kind() == reflect.Struct && len(fields) > 1:
			// If the field is a struct, call updateField recursively to update nested fields
			updateField(fields[1:], dst.Field(i), src.Field(i))

		case dst.Field(i).Kind() == reflect.Slice:
			//nolint:lll
			// If the field is a slice, append new values from updateValue to existing slice in targetValue
			// https://github.com/protocolbuffers/protobuf/blob/9bbea4aa65bdaf5fc6c2583e045c07ff37ffb0e7/src/google/protobuf/field_mask.proto#L111
			oldSlice := dst.Field(i)
			newSlice := src.Field(i)
			resultSlice := reflect.AppendSlice(oldSlice, newSlice)
			dst.Field(i).Set(resultSlice)

		default:
			// For all other field kinds, set value of field in targetValue to value of corresponding field in updateValue
			dst.Field(i).Set(src.Field(i))
		}
	}
}

// updateServiceFromFieldMask updates the targetService with the values from updateService based on the fieldMask.
// It returns a new *model.Service value that is a copy of targetService with the updates applied.
func updateServiceFromFieldMask(
	fieldMask *fieldmaskpb.FieldMask,
	dstService *model.Service,
	srcService *model.Service,
) *model.Service {
	// Set the ID of the srcService to the ID of the dstService, to prevent the ID from being updated
	srcService.ID = dstService.ID

	return updateModelFromFieldMask(fieldMask, dstService, srcService)
}

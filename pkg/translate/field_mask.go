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
//
// It returns a new *T value that is a copy of dst with the updates applied.
func updateModelFromFieldMask[T any](fieldMask *fieldmaskpb.FieldMask, dst, src *T) *T {
	// If fieldMask is nil, update all fields
	if fieldMask == nil {
		// Avoid returning a pointer to src.
		s := *src
		return &s
	}

	// Create a new value that is a copy of targetModel
	v := reflect.ValueOf(dst).Elem()
	result := reflect.New(v.Type()).Elem()
	result.Set(v)

	for _, path := range fieldMask.GetPaths() {
		fields := strings.Split(path, ".")
		updateField(fields, result, reflect.ValueOf(src).Elem())
	}

	return result.Addr().Interface().(*T)
}

// updateField updates the dst value with the values from src, based on the fields slice.
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
			// If the field is a slice, append new values from src to existing slice in dst
			// https://github.com/protocolbuffers/protobuf/blob/9bbea4aa65bdaf5fc6c2583e045c07ff37ffb0e7/src/google/protobuf/field_mask.proto#L111
			oldSlice := dst.Field(i)
			newSlice := src.Field(i)
			resultSlice := reflect.AppendSlice(oldSlice, newSlice)
			dst.Field(i).Set(resultSlice)

		default:
			// For all other field kinds, set value of field in dst to value of corresponding field in src
			dst.Field(i).Set(src.Field(i))
		}
	}
}

// updateServiceFromFieldMask updates the dstService with the values from srcService based on the fieldMask.
func updateServiceFromFieldMask(
	fieldMask *fieldmaskpb.FieldMask,
	dstService *model.Service,
	srcService *model.Service,
) *model.Service {
	// Set the ID of the srcService to the ID of the dstService, to prevent the ID from being updated
	srcService.ID = dstService.ID

	return updateModelFromFieldMask(fieldMask, dstService, srcService)
}

package model

import (
	"errors"
	"reflect"
	"slices"
	"strings"
)

// Mask is a list of paths that can be used to specify which fields in a resource should be updated or returned.
// Mask is not case sensitive. Paths are separated by "." and can be nested.
// Example:
//
//	 type Translation struct {
//		 Language language.Tag `json:"language"`
//		 Messages []Message    `json:"messages"`
//		 Original bool         `json:"original"`
//	 }
//
//	// valid masks
//	Mask{"language"}
//	Mask{"Language"}
//	Mask{"LANGUAGE"}
//	Mask{"messages"}
//	Mask{"language", "Original", "MESSAGES"}
//	...
type Mask []string

// update updates the dst with the values from src based on the mask.
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
//	update(&src, &dst, mask)
//	fmt.Println(dst) // Foo{Bar: "bar2", Baz: "baz"}.
func update[T any](src, dst *T, mask Mask) {
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
		if !strings.EqualFold(dst.Type().Field(i).Name, field) {
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

// ---------------------Model Implementations---------------------

// UpdateService updates the dst with the values from src based on the Mask.
// ID field is not allowed in the mask, as it is considered read-only.
// If the Mask is nil, all fields are updated, except the ID.
func UpdateService(src, dst *Service, mask Mask) error {
	// Prevent updating read-only fields like ID
	if slices.Contains(mask, "ID") {
		return errors.New("\"id\" is not allowed in field mask")
	}

	// When mask is nil dstService is updated with all fields from srcService
	// So we need to make sure that the ID is not updated
	src.ID = dst.ID
	update(src, dst, mask)

	return nil
}

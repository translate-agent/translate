package model

import (
	"errors"
	"reflect"
	"slices"
	"strings"
)

// Mask is a list of paths that can be used to specify which fields
// in a resource should be updated (PATCH) or returned (GET).
//
// Mask is case insensitive. In case of a struct, the paths are separated by a dot. e.g. "foo.bar.baz".
//
// Example:
//
//	type Example struct {
//		Field1 string
//		Field2 SubExample
//	}
//
//	type SubExample struct {
//		SubField1 string
//		SubField2 int
//	}
//
//	// valid masks for Example struct (in any case)
//	Mask{"Field1"} // field1, FIELD1, ...
//	Mask{"Field1", "Field2"}
//	Mask{"Field1", "Field2.SubField1"}
//	Mask{"Field1", "Field2.SubField1", "Field2.SubField2"}
//	Mask{"Field2.SubField1", "Field2.SubField2"}
//	...
type Mask []string

// Update updates the dst with the values from src based on the mask.
// When updating slices:
//   - replace the element if slice contains a struct with a field "ID"
//   - append in all other cases
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
//	Update(&src, &dst, mask)
//	fmt.Println(dst) // Foo{Bar: "bar2", Baz: "baz"}.
func Update[T any](src, dst *T, mask Mask) {
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

	for i := range dst.NumField() {
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
			updateSliceField(srcField, dstField)
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

// updateSliceField this function updates a destination slice by either replacing a struct with a matching ID
// or appending new values if no matching ID is found.
func updateSliceField(srcField, dstField reflect.Value) {
	appendOnly := func() bool {
		el := srcField.Type().Elem()

		if el.Kind() != reflect.Struct {
			return true
		}

		_, ok := el.FieldByName("ID")

		return !ok
	}

	if appendOnly() {
		dstField.Set(reflect.AppendSlice(dstField, srcField))
		return
	}

	// If found by id, replace the corresponding structure in the destination slice. Otherwise, append.
	for i := range srcField.Len() {
		srcID := srcField.Index(i).FieldByName("ID").Interface()

		idx := func(id interface{}) int {
			for j := range dstField.Len() {
				if dstField.Index(j).FieldByName("ID").Interface() == id {
					return j
				}
			}

			return -1
		}

		switch j := idx(srcID); j {
		default:
			dstField.Index(j).Set(srcField.Index(i))
		case -1:
			dstField.Set(reflect.Append(dstField, srcField.Index(i)))
		}
	}
}

// ---------------------Model Implementations---------------------

// UpdateService updates the dst with the values from src based on the Mask.
// ID field is not allowed in the mask, as it is considered read-only.
// If the Mask is nil, all fields are updated, except the ID.
func UpdateService(src, dst *Service, mask Mask) error {
	// Prevent updating read-only fields like ID
	if slices.ContainsFunc(mask, func(s string) bool {
		return strings.EqualFold(s, "ID")
	}) {
		return errors.New("\"id\" is not allowed in field mask")
	}

	// When mask is nil dstService is updated with all fields from srcService
	// So we need to make sure that the ID is not updated
	src.ID = dst.ID
	Update(src, dst, mask)

	return nil
}

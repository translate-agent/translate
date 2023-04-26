package translate

import (
	"fmt"
	"strings"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

// validateFieldMask validates that all the field paths in the given fieldMask are valid for the given protoMessage.
func validateFieldMask(fieldMask *fieldmaskpb.FieldMask, protoMessage proto.Message) error {
	// Get the message descriptor for the protoMessage
	messageDescriptor := protoMessage.ProtoReflect().Descriptor()

	fieldNames := getFieldNames(messageDescriptor)

	// validateFieldPath validates that the given pathsParts are valid for the given fieldNames.
	var validateFieldPath func([]string, map[string]protoreflect.FieldDescriptor) error

	validateFieldPath = func(pathParts []string, fieldNames map[string]protoreflect.FieldDescriptor) error {
		// If there are no more path parts to validate, return nil
		if len(pathParts) == 0 {
			return nil
		}

		// Get the first path part and check if it is a valid field name
		fieldName := pathParts[0]

		fieldDescriptor, ok := fieldNames[fieldName]
		if !ok {
			return fmt.Errorf("invalid field '%s'", fieldName)
		}

		// If there are more path parts and the current field is a nested message,
		// create a new map of field names for the nested message and call validateFieldPath recursively
		if length, msg := len(pathParts), fieldDescriptor.Message(); length > 1 && msg != nil {
			nestedFieldNames := getFieldNames(msg)
			return validateFieldPath(pathParts[1:], nestedFieldNames)
		}

		return nil
	}

	// Loop through each path in the field mask and split it into parts
	for _, path := range fieldMask.GetPaths() {
		pathParts := strings.Split(path, ".")
		if err := validateFieldPath(pathParts, fieldNames); err != nil {
			return fmt.Errorf("'%s' is not a valid %s field: %w", path, messageDescriptor.Name(), err)
		}
	}

	return nil
}

// getFieldNames returns a map of field names for the given message descriptor.
func getFieldNames(descriptor protoreflect.MessageDescriptor) map[string]protoreflect.FieldDescriptor {
	fields := descriptor.Fields()

	// Create a map to store the field names
	n := fields.Len()
	fieldNames := make(map[string]protoreflect.FieldDescriptor, n)

	// Loop through each field in the message descriptor and add its name to the map
	for i := 0; i < n; i++ {
		field := fields.Get(i)
		fieldNames[string(field.Name())] = field
	}

	return fieldNames
}

package translate

import (
	"fmt"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

func validateFieldMask(fieldMask *fieldmaskpb.FieldMask, protoMessage proto.Message) error {
	messageDescriptor := protoMessage.ProtoReflect().Descriptor()

	// Create a map to store the field names
	fieldNames := make(map[string]bool)

	// Loop through each field and add the name to the map
	fields := messageDescriptor.Fields()
	for i := 0; i < fields.Len(); i++ {
		fieldNames[string(fields.Get(i).Name())] = true
	}

	// Loop through each path in the field mask
	for _, path := range fieldMask.GetPaths() {
		// Check if the path is a valid field for the proto message
		if _, ok := fieldNames[path]; !ok {
			return fmt.Errorf("'%s' is not a valid %s field", path, messageDescriptor.Name())
		}
	}

	return nil
}

package translate

import (
	"fmt"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

// validateFieldMask validates that all the field paths in the given fieldMask are valid for the given protoMessage.
func validateFieldMask(fieldMask *fieldmaskpb.FieldMask, protoMessage proto.Message) error {
	// Check if the field mask is valid for the given proto message using the IsValid method
	if fieldMask.IsValid(protoMessage) {
		return nil
	}

	// As IsValid failed, the New function will always return an error, describing the invalid paths.
	_, err := fieldmaskpb.New(protoMessage, fieldMask.GetPaths()...)

	return fmt.Errorf("invalid field mask: %w", err)
}

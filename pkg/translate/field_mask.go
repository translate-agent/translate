package translate

import (
	"fmt"

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

package server

import (
	"errors"
	"fmt"

	"go.expect.digital/translate/pkg/convert"
	"go.expect.digital/translate/pkg/model"
	translatev1 "go.expect.digital/translate/pkg/pb/translate/v1"
)

var errUnspecifiedSchema = errors.New("unspecified schema")

/*
MessagesFromData converts in specific schema serialized data to model.Messages.
  - If original is true, then all messages are marked as TRANSLATED
  - If original is false, then all messages are marked as UNTRANSLATED or FUZZY (if schema supports fuzzy messages)

TODO: Add support for converting non original, but already translated messages and mark them as TRANSLATED
*/
func MessagesFromData(params *uploadParams) (*model.Messages, error) {
	var from func([]byte, bool) (model.Messages, error)

	switch params.schema {
	case translatev1.Schema_ARB:
		from = convert.FromArb
	case translatev1.Schema_GO:
		from = convert.FromGo
	case translatev1.Schema_JSON_NG_LOCALIZE:
		from = convert.FromNgLocalize
	case translatev1.Schema_JSON_NGX_TRANSLATE:
		from = convert.FromNgxTranslate
	case translatev1.Schema_POT:
		from = convert.FromPot
	case translatev1.Schema_XLIFF_2:
		from = convert.FromXliff2
	case translatev1.Schema_XLIFF_12:
		from = convert.FromXliff12
	case translatev1.Schema_UNSPECIFIED:
		return nil, errUnspecifiedSchema
	}

	messages, err := from(params.data, params.original)
	if err != nil {
		return nil, fmt.Errorf("convert from %s schema: %w", params.schema, err)
	}

	return &messages, nil
}

// MessagesToData converts model.Messages to specific schema serialized data.
func MessagesToData(schema translatev1.Schema, messages *model.Messages) ([]byte, error) {
	var to func(model.Messages) ([]byte, error)

	switch schema {
	case translatev1.Schema_ARB:
		to = convert.ToArb
	case translatev1.Schema_GO:
		to = convert.ToGo
	case translatev1.Schema_JSON_NG_LOCALIZE:
		to = convert.ToNgLocalize
	case translatev1.Schema_JSON_NGX_TRANSLATE:
		to = convert.ToNgxTranslate
	case translatev1.Schema_POT:
		to = convert.ToPot
	case translatev1.Schema_XLIFF_2:
		to = convert.ToXliff2
	case translatev1.Schema_XLIFF_12:
		to = convert.ToXliff12
	case translatev1.Schema_UNSPECIFIED:
		return nil, errUnspecifiedSchema
	}

	// Prevent nil pointer dereference.
	if messages == nil {
		messages = &model.Messages{}
	}

	data, err := to(*messages)
	if err != nil {
		return nil, fmt.Errorf("convert to %s schema: %w", schema, err)
	}

	return data, nil
}

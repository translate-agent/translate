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
TranslationFromData converts in specific schema serialized data to model.Translation.
  - If original is true, then all translation messages are marked as TRANSLATED
  - If original is false, then all translation messages are marked as UNTRANSLATED or FUZZY (if schema supports fuzzy translation)

TODO: Add support for converting non original, but already translated messages and mark them as TRANSLATED
*/
func TranslationFromData(params *uploadParams) (*model.Translation, error) {
	var from func([]byte, bool) (model.Translation, error)

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

	translation, err := from(params.data, params.original)
	if err != nil {
		return nil, fmt.Errorf("convert from %s schema: %w", params.schema, err)
	}

	return &translation, nil
}

// TranslationToData converts model.Translation to specific schema serialized data.
func TranslationToData(schema translatev1.Schema, translation *model.Translation) ([]byte, error) {
	var to func(model.Translation) ([]byte, error)

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
	if translation == nil {
		translation = &model.Translation{}
	}

	data, err := to(*translation)
	if err != nil {
		return nil, fmt.Errorf("convert to %s schema: %w", schema, err)
	}

	return data, nil
}

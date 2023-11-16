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
  - If original is true, then all messages are marked as TRANSLATED
  - If original is false, then all messages are marked as UNTRANSLATED or FUZZY (if schema supports fuzzy translation)

TODO: Add support for converting non original, but already translated messages and mark them as TRANSLATED.
*/
func TranslationFromData(params *uploadParams) (*model.Translation, error) {
	var (
		err         error
		original    bool
		translation model.Translation
	)

	if params.original != nil {
		original = *params.original
	}

	switch params.schema {
	case translatev1.Schema_ARB:
		translation, err = convert.FromArb(params.data, original)
	case translatev1.Schema_GO:
		translation, err = convert.FromGo(params.data, original)
	case translatev1.Schema_JSON_NG_LOCALIZE:
		translation, err = convert.FromNgLocalize(params.data, original)
	case translatev1.Schema_JSON_NGX_TRANSLATE:
		translation, err = convert.FromNgxTranslate(params.data, original)
	case translatev1.Schema_POT:
		translation, err = convert.FromPot(params.data, original)
	case translatev1.Schema_XLIFF_2:
		translation, err = convert.FromXliff2(params.data, params.original)
	case translatev1.Schema_XLIFF_12:
		translation, err = convert.FromXliff12(params.data, params.original)
	case translatev1.Schema_UNSPECIFIED:
		return nil, errUnspecifiedSchema
	}

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

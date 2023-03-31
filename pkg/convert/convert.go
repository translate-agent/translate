package convert

import (
	"fmt"

	"go.expect.digital/translate/pkg/model"
	translatev1 "go.expect.digital/translate/pkg/pb/translate/v1"
)

func From(schema translatev1.Schema, data []byte) (model.Messages, error) {
	var from func([]byte) (model.Messages, error)

	switch schema {
	case translatev1.Schema_ARB:
		from = FromArb
	case translatev1.Schema_GO:
		from = FromGo
	case translatev1.Schema_JSON_NG_LOCALISE:
		from = FromNgLocalize
	case translatev1.Schema_JSON_NGX_TRANSLATE:
		from = FromNgxTranslate
	case translatev1.Schema_POT:
		from = FromPot
	case translatev1.Schema_XLIFF_2:
		from = FromXliff2
	case translatev1.Schema_XLIFF_12:
		from = FromXliff12
	case translatev1.Schema_UNSPECIFIED:
		return model.Messages{}, fmt.Errorf("unspecified schema")
	}

	return from(data)
}

func To(schema translatev1.Schema, messages model.Messages) ([]byte, error) {
	var to func(model.Messages) ([]byte, error)

	switch schema {
	case translatev1.Schema_ARB:
		to = ToArb
	case translatev1.Schema_GO:
		to = ToGo
	case translatev1.Schema_JSON_NG_LOCALISE:
		to = ToNgLocalize
	case translatev1.Schema_JSON_NGX_TRANSLATE:
		to = ToNgxTranslate
	case translatev1.Schema_POT:
		to = ToPot
	case translatev1.Schema_XLIFF_2:
		to = ToXliff2
	case translatev1.Schema_XLIFF_12:
		to = ToXliff12
	case translatev1.Schema_UNSPECIFIED:
		return nil, fmt.Errorf("unspecified schema")
	}

	return to(messages)
}

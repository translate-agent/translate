package convert

import (
	"encoding/json"
	"errors"
	"fmt"

	"go.expect.digital/mf2/builder"
	ast "go.expect.digital/mf2/parse"

	"go.expect.digital/translate/pkg/model"
)

// FromNgxTranslate  parses the JSON-encoded byte slice representing messages in the ngx-translate format,
// recursively traverses the map, extracts the key-value pairs, converts the message strings,
// and constructs a model.Translation structure.
func FromNgxTranslate(b []byte, original *bool) (translation model.Translation, err error) {
	// if original is not provided default to false.
	if original == nil {
		original = ptr(false)
	}

	translation.Original = *original

	var dst map[string]any

	err = json.Unmarshal(b, &dst)
	if err != nil {
		return translation, fmt.Errorf("unmarshal from ngx-translate to model.Translation: %w", err)
	}

	var traverseMap func(key string, value any) error

	status := model.MessageStatusUntranslated
	if *original {
		status = model.MessageStatusTranslated
	}

	traverseMap = func(key string, value any) (err error) {
		switch v := value.(type) {
		default:
			return fmt.Errorf("unsupported value type %T for key %s", value, key)
		case string:
			msg, err := builder.NewBuilder().Text(v).Build() //nolint:govet
			if err != nil {
				return fmt.Errorf("convert string to MF2: %w", err)
			}

			translation.Messages = append(translation.Messages, model.Message{
				ID:      key,
				Message: msg,
				Status:  status,
			})
		case map[string]any:
			for subKey, subValue := range v {
				if key != "" {
					subKey = key + "." + subKey
				}

				err = traverseMap(subKey, subValue)
				if err != nil {
					return err
				}
			}
		}

		return err
	}

	err = traverseMap("", dst)
	if err != nil {
		return translation, fmt.Errorf("traverse ngx-translate: %w", err)
	}

	return translation, nil
}

// ToNgxTranslate converts a model.Translation structure into the ngx-translate format.
func ToNgxTranslate(translation model.Translation) ([]byte, error) {
	dst := make(map[string]string, len(translation.Messages))

	for _, msg := range translation.Messages {
		tree, err := ast.Parse(msg.Message)
		if err != nil {
			return nil, fmt.Errorf("parse mf2 message: %w", err)
		}

		switch mf2Msg := tree.Message.(type) {
		case nil:
			dst[msg.ID] = ""
		case ast.SimpleMessage:
			dst[msg.ID] = patternsToSimpleMsg(mf2Msg)
		case ast.ComplexMessage:
			return nil, errors.New("complex message not supported")
		}
	}

	b, err := json.Marshal(dst)
	if err != nil {
		return nil, fmt.Errorf("marshal from model.Translation to ngx-translate: %w", err)
	}

	return b, nil
}

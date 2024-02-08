package convert

import (
	"encoding/json"
	"fmt"

	ast "go.expect.digital/mf2/parse"

	"go.expect.digital/mf2"

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

	var dst map[string]interface{}

	if err = json.Unmarshal(b, &dst); err != nil {
		return translation, fmt.Errorf("unmarshal from ngx-translate to model.Translation: %w", err)
	}

	var traverseMap func(key string, value interface{}) error

	status := model.MessageStatusUntranslated
	if *original {
		status = model.MessageStatusTranslated
	}

	traverseMap = func(key string, value interface{}) (err error) {
		switch v := value.(type) {
		default:
			return fmt.Errorf("unsupported value type %T for key %s", value, key)
		case string:
			msg, err := mf2.NewBuilder().Text(v).Build() //nolint:govet
			if err != nil {
				return fmt.Errorf("convert string to MF2: %w", err)
			}

			translation.Messages = append(translation.Messages, model.Message{
				ID:      key,
				Message: msg,
				Status:  status,
			})
		case map[string]interface{}:
			for subKey, subValue := range v {
				if key != "" {
					subKey = key + "." + subKey
				}

				if err = traverseMap(subKey, subValue); err != nil {
					return err
				}
			}
		}

		return err
	}

	if err = traverseMap("", dst); err != nil {
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
			dst[msg.ID] = patternsToMsg(mf2Msg)
		case ast.ComplexMessage:
			return nil, fmt.Errorf("complex message not supported")
		}
	}

	b, err := json.Marshal(dst)
	if err != nil {
		return nil, fmt.Errorf("marshal from model.Translation to ngx-translate: %w", err)
	}

	return b, nil
}

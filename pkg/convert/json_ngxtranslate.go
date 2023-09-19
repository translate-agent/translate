package convert

import (
	"encoding/json"
	"fmt"

	"go.expect.digital/translate/pkg/model"
)

// FromNgxTranslate  parses the JSON-encoded byte slice representing messages in the ngx-translate format,
// recursively traverses the map, extracts the key-value pairs, converts the message strings,
// and constructs a model.Messages structure.
func FromNgxTranslate(b []byte, original bool) (messages model.Messages, err error) {
	messages.Original = original

	var dst map[string]interface{}

	if err = json.Unmarshal(b, &dst); err != nil {
		return messages, fmt.Errorf("unmarshal from ngx-translate to model.Messages: %w", err)
	}

	var traverseMap func(key string, value interface{}) error

	status := model.MessageStatusUntranslated
	if original {
		status = model.MessageStatusTranslated
	}

	traverseMap = func(key string, value interface{}) (err error) {
		switch v := value.(type) {
		default:
			return fmt.Errorf("unsupported value type %T for key %s", value, key)
		case string:
			messages.Messages = append(messages.Messages, model.Message{
				ID:      key,
				Message: convertToMessageFormatSingular(v),
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
		return messages, fmt.Errorf("traverse ngx-translate: %w", err)
	}

	return messages, nil
}

// ToNgxTranslate converts a model.Messages structure into the ngx-translate format.
func ToNgxTranslate(messages model.Messages) (b []byte, err error) {
	dst := make(map[string]string, len(messages.Messages))

	for _, msg := range messages.Messages {
		dst[msg.ID] = removeEnclosingBrackets(msg.Message)
	}

	b, err = json.Marshal(dst)
	if err != nil {
		return nil, fmt.Errorf("marshal from model.Messages to ngx-translate: %w", err)
	}

	return b, nil
}

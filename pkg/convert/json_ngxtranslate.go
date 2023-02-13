package convert

import (
	"encoding/json"
	"errors"
	"fmt"

	"go.expect.digital/translate/pkg/model"
)

func FromNgxTranslate(b []byte) (messages model.Messages, err error) {
	var dst map[string]interface{}

	if err = json.Unmarshal(b, &dst); err != nil {
		return messages, fmt.Errorf("unmarshal from ngx-translate to model.Messages: %w", err)
	}

	var traverseMap func(key string, value interface{}) error

	traverseMap = func(key string, value interface{}) (err error) {
		switch v := value.(type) {
		default:
			return errors.New("unsupported value type")
		case string:
			messages.Messages = append(messages.Messages, model.Message{ID: key, Message: v})
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

func ToNgxTranslate(messages model.Messages) (b []byte, err error) {
	dst := make(map[string]string, len(messages.Messages))

	for _, msg := range messages.Messages {
		dst[msg.ID] = msg.Message
	}

	b, err = json.Marshal(dst)
	if err != nil {
		return nil, fmt.Errorf("marshal to ngx-translate from model.Messages : %w", err)
	}

	return b, nil
}

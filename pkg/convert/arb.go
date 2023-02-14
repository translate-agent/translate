package convert

import (
	"encoding/json"
	"fmt"

	"github.com/mitchellh/mapstructure"
	"go.expect.digital/translate/pkg/model"
)

/* .arb files employ the key-value (key-translation) format, with separate files representing different languages.
Specification: https://docs.flutter.dev/development/accessibility-and-localization/internationalization

Example:

app_en.arb
{
	"title": "Hello World!",
	"@title" : {
		"description" : "Message to greet the World"
	},
	"greeting": "Welcome {user}!",
	"@greeting": {
		"placeholders": {
			"user":{
				"type":"string",
				"example":"Bob"
			}
		}
	}
}

app_fr.arb
{
  "title": "Bonjour le monde!",
  "greeting": "Bienvenue {user}!"
}

*/

// Converts a serialized data in ARB file format into model.Messages.
func FromArb(data []byte) (model.Messages, error) {
	var dst map[string]interface{}
	if err := json.Unmarshal(data, &dst); err != nil {
		return model.Messages{}, fmt.Errorf("unmarshal ARB serialized data: %w", err)
	}

	findDescription := func(key string) (string, error) {
		// if '@' prefix missing then no additional information is provided
		subKeyMap, ok := dst["@"+key]
		if !ok {
			return "", nil
		}

		var meta struct {
			Description string `json:"description"`
		}

		if err := mapstructure.Decode(subKeyMap, &meta); err != nil {
			return "", fmt.Errorf("decode metadata map: %w", err)
		}

		return meta.Description, nil
	}

	var messages model.Messages

	for key, value := range dst {
		// Ignore a key if it begins with '@' as it only supplies metadata for message not the message itself.
		if key[0] == '@' {
			continue
		}

		msg := model.Message{ID: key}

		// If a key does not have an '@' prefix and its value is not of type string, then file is not formatted correctly.
		var ok bool
		if msg.Message, ok = value.(string); !ok {
			return model.Messages{}, fmt.Errorf("unsupported value type '%T' for key '%s'", value, key)
		}

		var err error
		if msg.Description, err = findDescription(key); err != nil {
			return model.Messages{}, fmt.Errorf("find description of '%s': %w", key, err)
		}

		messages.Messages = append(messages.Messages, msg)
	}

	return messages, nil
}

// Converts model.Messages into a serialized data in ARB file format.
func ToArb(messages model.Messages) ([]byte, error) {
	dst := make(map[string]interface{}, len(messages.Messages)*2) //nolint:gomnd

	for _, msg := range messages.Messages {
		dst[msg.ID] = msg.Message
		if len(msg.Description) > 0 {
			dst["@"+msg.ID] = map[string]string{"description": msg.Description}
		}
	}

	result, err := json.Marshal(dst)
	if err != nil {
		return nil, fmt.Errorf("marshaling model.messages to Arb: %w", err)
	}

	return result, nil
}

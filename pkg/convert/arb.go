package convert

import (
	"encoding/json"
	"fmt"

	"go.expect.digital/translate/pkg/model"
)

/* .arb files employ the key-value (key-translation) format, with separate files representing different languages.

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
		return model.Messages{}, fmt.Errorf("unmarshal from ARB to model.Messages: %w", err)
	}

	findDescription := func(key string) (string, error) {
		// if '@' prefix missing then no additional information is provided
		subKey, ok := dst["@"+key]
		if !ok {
			return "", nil
		}

		keyMap, ok := subKey.(map[string]interface{})
		if !ok {
			return "", fmt.Errorf("unsupported value type '%T' for key '%s'", subKey, "@"+key)
		}

		description, ok := keyMap["description"]
		if !ok {
			return "", nil
		}

		if descriptionValue, ok := description.(string); ok {
			return descriptionValue, nil
		} else {
			return "", fmt.Errorf("unsupported value type '%T' for key 'description'", descriptionValue)
		}
	}

	var messages model.Messages

	for key, value := range dst {
		// if key starts with '@' we ignore it as it provides additional information not the message itself
		if key[0] == '@' {
			continue
		}

		// if there is no '@' prefix and value type is not string then file is not formatted correctly
		value, ok := value.(string)
		if !ok {
			return model.Messages{}, fmt.Errorf("unsupported value type '%T' for key '%s'", value, key)
		}

		description, err := findDescription(key)
		if err != nil {
			return model.Messages{}, fmt.Errorf("find description of '%s': %w", key, err)
		}

		messages.Messages = append(messages.Messages, model.Message{
			ID:          key,
			Message:     value,
			Description: description,
		})
	}

	return messages, nil
}

// Converts model.Messages into a serialized data in ARB file format.
func ToArb(messages model.Messages) ([]byte, error) {
	dst := make(map[string]interface{}, len(messages.Messages))

	for _, msg := range messages.Messages {
		dst[msg.ID] = msg.Message
		if len(msg.Description) != 0 {
			dst["@"+msg.ID] = map[string]string{"description": msg.Description}
		}
	}

	result, err := json.Marshal(dst)
	if err != nil {
		return nil, fmt.Errorf("marshaling model.messages to Arb: %w", err)
	}

	return result, nil
}

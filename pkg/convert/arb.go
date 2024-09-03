package convert

import (
	"encoding/json"
	"errors"
	"fmt"

	ast "go.expect.digital/mf2/parse"

	"go.expect.digital/mf2/builder"

	"github.com/mitchellh/mapstructure"
	"go.expect.digital/translate/pkg/model"
	"golang.org/x/text/language"
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

// FromArb converts a serialized data in ARB file format into model.Translation.
func FromArb(data []byte, original *bool) (model.Translation, error) {
	var dst map[string]interface{}
	if err := json.Unmarshal(data, &dst); err != nil {
		return model.Translation{}, fmt.Errorf("unmarshal ARB serialized data: %w", err)
	}

	// if original is not provided default to false.
	if original == nil {
		original = ptr(false)
	}

	findDescription := func(key string) (string, error) {
		// if '@' prefix missing then no additional information is provided.
		subKeyMap, ok := dst["@"+key]
		if !ok {
			return "", nil
		}

		var meta struct {
			Description string `json:"description" mapstructure:"description"`
		}

		if err := mapstructure.Decode(subKeyMap, &meta); err != nil {
			return "", fmt.Errorf("decode metadata map: %w", err)
		}

		return meta.Description, nil
	}

	//nolint:lll
	// ARB file can have optional '@@locale' top level key for source key language.
	// https://medium.com/@Albert221/how-to-internationalize-your-flutter-app-with-arb-files-today-full-blown-tutorial-476ee65ecaed
	findLocale := func() (language.Tag, error) {
		// if '@@locale' is missing then language it is not provided (Undetermined).
		locale, ok := dst["@@locale"]
		if !ok {
			return language.Tag{}, nil
		}

		// Check if @@locale key's value type is string.
		langString, ok := locale.(string)
		if !ok {
			return language.Tag{}, fmt.Errorf(`unsupported value type "%T" for key "@@locale"`, locale)
		}

		lang, err := language.Parse(langString)
		if err != nil {
			return language.Tag{}, fmt.Errorf("parse language: %w", err)
		}

		return lang, nil
	}

	lang, err := findLocale()
	if err != nil {
		return model.Translation{}, fmt.Errorf("find locale: %w", err)
	}

	status := model.MessageStatusUntranslated
	if *original {
		status = model.MessageStatusTranslated
	}

	translation := model.Translation{Language: lang, Original: *original}

	for key, value := range dst {
		// Ignore a key if it begins with '@' as it only supplies metadata for translation not the message itself.
		if key[0] == '@' {
			continue
		}

		msg := model.Message{ID: key, Status: status}

		// If a key does not have an '@' prefix and its value is not of type string, then file is not formatted correctly.
		var ok bool
		if msg.Message, ok = value.(string); !ok {
			return model.Translation{}, fmt.Errorf("unsupported value type '%T' for key '%s'", value, key)
		}

		if msg.Message, err = builder.NewBuilder().Text(msg.Message).Build(); err != nil {
			return model.Translation{}, fmt.Errorf("convert string to MF2: %w", err)
		}

		if msg.Description, err = findDescription(key); err != nil {
			return model.Translation{}, fmt.Errorf(`find description of "%s": %w`, key, err)
		}

		translation.Messages = append(translation.Messages, msg)
	}

	return translation, nil
}

// ToArb converts model.Translation into a serialized data in ARB file format.
func ToArb(translation model.Translation) ([]byte, error) {
	// dst length = number of messages + number of potential descriptions (same as number of messages) + locale.
	dst := make(map[string]interface{}, len(translation.Messages)*2+1)

	// "und" (Undetermined) language.Tag is also valid BCP47 tag.
	dst["@@locale"] = translation.Language

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

		if len(msg.Description) > 0 {
			dst["@"+msg.ID] = map[string]string{"description": msg.Description}
		}
	}

	result, err := json.Marshal(dst)
	if err != nil {
		return nil, fmt.Errorf("marshal dst map to ARB: %w", err)
	}

	return result, nil
}

func ptr[T any](v T) *T {
	return &v
}

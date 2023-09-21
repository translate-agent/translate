package model

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"sync"

	"golang.org/x/exp/slices"
	"golang.org/x/text/language"
)

type Messages struct {
	Language language.Tag
	Messages []Message
	Original bool
}

type MessagesSlice []Messages

// HasLanguage checks if MessagesSlice contains Messages with the given language.
func (ms MessagesSlice) HasLanguage(lang language.Tag) bool {
	return ms.LanguageIndex(lang) != -1
}

// LanguageIndex returns index of Messages with the given language. If not found, returns -1.
func (ms MessagesSlice) LanguageIndex(lang language.Tag) int {
	return slices.IndexFunc(ms, func(m Messages) bool {
		return m.Language == lang
	})
}

// Replace replaces Messages with the same language. If not found, appends to the slice.
func (ms *MessagesSlice) Replace(messages Messages) {
	switch idx := ms.LanguageIndex(messages.Language); idx {
	case -1:
		*ms = append(*ms, messages)
	default:
		(*ms)[idx] = messages
	}
}

// SplitOriginal returns a pointer to the original and other messages.
func (m MessagesSlice) SplitOriginal() (original *Messages, others MessagesSlice) {
	others = make(MessagesSlice, 0, len(m))

	for i := range m {
		if m[i].Original {
			original = &m[i]
		} else {
			others = append(others, m[i])
		}
	}

	return
}

/*
MarkUntranslated changes status of message in all languages except the original to
UNTRANSLATED if message.ID is in the ids slice.

Example:

	ids := { "1" }

	{ Language: en, Original: true, Messages: [ { ID: "1", Message: "Hello", Status: Translated  } ], ...
	{ Language: fr, Messages: [ { ID: "1", Message: "Bonjour", Status: Translated  } ], ... ] }
	{ Language: de, Messages: [ { ID: "1", Message: "Hallo", Status: Translated  } ], ... ]

	Result:
	{ Language: en, Original: true, Messages: [ { ID: "1", Message: "Hello", Status: Translated  } ], ...
	{ Language: fr, Messages: [ { ID: "1", Message: "Bonjour", Status: Untranslated  }, ... ] }
	{ Language: de, Messages: [ { ID: "1", Message: "Hallo", Status: Untranslated  } ], ... ]
*/
func (ms MessagesSlice) MarkUntranslated(ids []string) {
	n := len(ms)
	if len(ids) == 0 || n == 0 || (n == 1 && ms[0].Original) {
		return
	}

	slices.Sort(ids)

	for _, messages := range ms {
		if messages.Original {
			continue
		}

		for i := range messages.Messages {
			if _, found := slices.BinarySearch(ids, messages.Messages[i].ID); found {
				messages.Messages[i].Status = MessageStatusUntranslated
			}
		}
	}
}

/*
PopulateTranslations adds missing messages from the original language to other languages.
Example:

	Input:
	MessagesSlice{
		{
			Language: en,
			Original: true,
			Messages: [ { ID: "1", Message: "Hello" }, { ID: "2", Message: "World" } ],
		},
		{
			Language: fr,
			Original: false,
			Messages: [ { ID: "1", Message: "Bonjour" } ],
		},
	}

	Output:
	MessagesSlice{
		{
			Language: en,
			Original: true,
			Messages: [ { ID: "1", Message: "Hello" }, { ID: "2", Message: "World" } ],
		},
		{
			Language: fr,
			Original: false,
			Messages: [ { ID: "1", Message: "Bonjour" }, { ID: "2", Message: "World", Status: Untranslated } ],
		},
*/
func (ms MessagesSlice) PopulateTranslations() {
	origIdx := slices.IndexFunc(ms, func(m Messages) bool { return m.Original })
	if origIdx == -1 {
		return
	}

	var wg sync.WaitGroup

	for i := range ms {
		// Skip original messages
		if i == origIdx {
			continue
		}

		wg.Add(1)

		populate := func(i int) {
			defer wg.Done()

			// Current language's message lookup.
			lookup := make(map[string]struct{}, len(ms[i].Messages))
			for _, message := range ms[i].Messages {
				lookup[message.ID] = struct{}{}
			}

			// Add missing message from original language
			for _, origMsg := range ms[origIdx].Messages {
				if _, ok := lookup[origMsg.ID]; !ok {
					origMsg.Status = MessageStatusUntranslated
					ms[i].Messages = append(ms[i].Messages, origMsg)
				}
			}
		}

		go populate(i)
	}

	wg.Wait()
}

type Message struct {
	ID          string
	PluralID    string
	Message     string
	Description string
	Positions   Positions
	Status      MessageStatus
}

type MessageStatus int32

const (
	MessageStatusUntranslated MessageStatus = iota
	MessageStatusFuzzy
	MessageStatusTranslated
)

func (s MessageStatus) String() string {
	switch s {
	default:
		return ""
	case MessageStatusUntranslated:
		return "UNTRANSLATED"
	case MessageStatusFuzzy:
		return "FUZZY"
	case MessageStatusTranslated:
		return "TRANSLATED"
	}
}

// Value implements driver.Valuer interface.
func (s MessageStatus) Value() (driver.Value, error) {
	return s.String(), nil
}

// Scan implements sql.Scanner interface.
func (s *MessageStatus) Scan(value interface{}) error {
	switch v := value.(type) {
	default:
		return fmt.Errorf("unknown type %+v, expected string", v)
	case []byte:
		switch string(v) {
		case MessageStatusUntranslated.String():
			*s = MessageStatusUntranslated
		case MessageStatusFuzzy.String():
			*s = MessageStatusFuzzy
		case MessageStatusTranslated.String():
			*s = MessageStatusTranslated
		default:
			return fmt.Errorf("unknown message status: %+v", v)
		}
	}

	return nil
}

type Positions []string

// Value implements driver.Valuer interface.
func (p Positions) Value() (driver.Value, error) {
	if len(p) == 0 {
		return nil, nil
	}

	b, err := json.Marshal(p)
	if err != nil {
		return nil, fmt.Errorf("json marshal positions: %w", err)
	}

	return b, nil
}

// Scan implements sql.Scanner interface.
func (p *Positions) Scan(value interface{}) error {
	switch v := value.(type) {
	default:
		return fmt.Errorf("unknown type %+v, expected []byte", v)
	case nil:
		*p = nil
		return nil
	case []byte:
		if err := json.Unmarshal(v, &p); err != nil {
			return fmt.Errorf("json unmarshal positions: %w", err)
		}

		return nil
	}
}

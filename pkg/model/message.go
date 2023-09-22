package model

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"slices"
	"sync"

	"golang.org/x/text/language"
)

type Messages struct {
	Language language.Tag
	Messages []Message
	Original bool
}

// MessageIndex returns index of Message with the given ID. If not found, returns -1.
func (m *Messages) MessageIndex(id string) int {
	return slices.IndexFunc(m.Messages, func(msg Message) bool {
		return msg.ID == id
	})
}

/*
FindChangedMessageIDs returns a list of message IDs that have been altered in the new Messages e.g.
 1. The message.message has been changed
 2. The message with new ID has been added.
*/
func (m *Messages) FindChangedMessageIDs(new *Messages) []string {
	lookup := make(map[string]string, len(m.Messages))

	for _, msg := range m.Messages {
		lookup[msg.ID] = msg.Message
	}

	var ids []string

	for _, msg := range new.Messages {
		if oldMsg, ok := lookup[msg.ID]; !ok || oldMsg != msg.Message {
			ids = append(ids, msg.ID)
		}
	}

	return ids
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

// OriginalIndex returns index of Messages with the original flag set to true. If not found, returns -1.
func (ms MessagesSlice) OriginalIndex() int {
	return slices.IndexFunc(ms, func(m Messages) bool {
		return m.Original
	})
}

// Replace replaces Messages with the same language. If not found, appends it.
func (ms *MessagesSlice) Replace(messages Messages) {
	switch idx := ms.LanguageIndex(messages.Language); idx {
	case -1:
		*ms = append(*ms, messages)
	default:
		(*ms)[idx] = messages
	}
}

/*
MarkUntranslated changes status of message in all languages except the original to
UNTRANSLATED if message.ID is in the ids slice.

Example:

	Input:
	ids := { "1" }
	MessagesSlice{
		{
			Language: en,
			Original: true,
			Messages: [ { ID: "1", Message: "Hello", Status: Translated,  }, ... ],
		},
		{
			Language: fr,
			Original: false,
			Messages: [ { ID: "1", Message: "Bonjour", Status: Translated }, ... ],
		},
	}

	Output:
	MessagesSlice{
		{
			Language: en,
			Original: true,
			Messages: [ { ID: "1", Message: "Hello", Status: Translated  }, ... ],
		},
		{
			Language: fr,
			Original: false,
			Messages: [ { ID: "1", Message: "Bonjour", Status: Untranslated  }, ... ],
		}
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
		if ms[i].Original {
			continue
		}

		wg.Add(1)

		populate := func(i int) {
			defer wg.Done()

			lookup := make(map[string]struct{}, len(ms[i].Messages))
			for j := range ms[i].Messages {
				lookup[ms[i].Messages[j].ID] = struct{}{}
			}

			for j := range ms[origIdx].Messages {
				if _, ok := lookup[ms[origIdx].Messages[j].ID]; !ok {
					newMsg := ms[origIdx].Messages[j]
					newMsg.Status = MessageStatusUntranslated

					ms[i].Messages = append(ms[i].Messages, newMsg)
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

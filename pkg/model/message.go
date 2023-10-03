package model

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"slices"
	"sync"

	"golang.org/x/text/language"
)

type Translation struct {
	Language language.Tag
	Messages []Message
	Original bool
}

/*
FindChangedMessageIDs returns a list of message IDs that have been altered in the new Translation e.g.
 1. The message.message has been changed
 2. The message with new ID has been added.
*/
func (t *Translation) FindChangedMessageIDs(new *Translation) []string {
	lookup := make(map[string]int, len(t.Messages))
	for i := range t.Messages {
		lookup[t.Messages[i].ID] = i
	}

	var ids []string

	for _, msg := range new.Messages {
		if idx, ok := lookup[msg.ID]; !ok || t.Messages[idx].Message != msg.Message {
			ids = append(ids, msg.ID)
		}
	}

	return ids
}

type Translations []Translation

// HasLanguage checks if Translations contains Translation with the given language.
func (ts Translations) HasLanguage(lang language.Tag) bool {
	return ts.LanguageIndex(lang) != -1
}

// LanguageIndex returns index of Translation with the given language. If not found, returns -1.
func (ts Translations) LanguageIndex(lang language.Tag) int {
	return slices.IndexFunc(ts, func(m Translation) bool {
		return m.Language == lang
	})
}

// OriginalIndex returns index of Translation with the original flag set to true. If not found, returns -1.
func (ts Translations) OriginalIndex() int {
	return slices.IndexFunc(ts, func(m Translation) bool {
		return m.Original
	})
}

// Replace replaces Translation with the same language. If not found, appends it.
func (ts *Translations) Replace(translation Translation) {
	switch idx := ts.LanguageIndex(translation.Language); idx {
	case -1:
		*ts = append(*ts, translation)
	default:
		(*ts)[idx] = translation
	}
}

/*
MarkUntranslated changes status of message in all languages except the original to
UNTRANSLATED if message.ID is in the ids slice.

Example:

	Input:
	ids := { "1" }
	Translations{
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
	Translations{
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
func (ts Translations) MarkUntranslated(ids []string) {
	n := len(ts)
	if len(ids) == 0 || n == 0 || (n == 1 && ts[0].Original) {
		return
	}

	slices.Sort(ids)

	for _, translation := range ts {
		if translation.Original {
			continue
		}

		for i := range translation.Messages {
			if _, found := slices.BinarySearch(ids, translation.Messages[i].ID); found {
				translation.Messages[i].Status = MessageStatusUntranslated
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
func (ts Translations) PopulateTranslations() {
	origIdx := slices.IndexFunc(ts, func(m Translation) bool { return m.Original })
	if origIdx == -1 {
		return
	}

	var wg sync.WaitGroup

	for i := range ts {
		if ts[i].Original {
			continue
		}

		wg.Add(1)

		populate := func(i int) {
			defer wg.Done()

			lookup := make(map[string]struct{}, len(ts[i].Messages))
			for j := range ts[i].Messages {
				lookup[ts[i].Messages[j].ID] = struct{}{}
			}

			for j := range ts[origIdx].Messages {
				if _, ok := lookup[ts[origIdx].Messages[j].ID]; !ok {
					newMsg := ts[origIdx].Messages[j]
					newMsg.Status = MessageStatusUntranslated

					ts[i].Messages = append(ts[i].Messages, newMsg)
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
	MessageStatusTranslated MessageStatus = iota
	MessageStatusFuzzy
	MessageStatusUntranslated
)

func (s MessageStatus) String() string {
	switch s {
	default:
		return ""
	case MessageStatusTranslated:
		return "TRANSLATED"
	case MessageStatusFuzzy:
		return "FUZZY"
	case MessageStatusUntranslated:
		return "UNTRANSLATED"
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
		case MessageStatusTranslated.String():
			*s = MessageStatusTranslated
		case MessageStatusFuzzy.String():
			*s = MessageStatusFuzzy
		case MessageStatusUntranslated.String():
			*s = MessageStatusUntranslated
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

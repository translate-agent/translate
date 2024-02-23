package po

type PO struct {
	Headers  Headers
	Messages []Message
}

type Headers []Header

// Get returns the value of the header with the given key if exists, otherwise it returns an empty string.
func (h Headers) Get(key string) string {
	for _, header := range h {
		if header.Name == key {
			return header.Value
		}
	}

	return ""
}

type Header struct {
	Name  string
	Value string
}

type Message struct {
	MsgID       string
	MsgIDPlural string
	MsgStr      []string

	TranslatorComments []string
	ExtractedComments  []string
	References         []string
	Flags              []string
}

package model

import "golang.org/x/text/language"

type Messages struct {
	Language language.Tag
	Messages []Message
	Original bool
}

type Message struct {
	ID          string
	PluralID    string
	Message     string
	Description string
	Fuzzy       bool
}

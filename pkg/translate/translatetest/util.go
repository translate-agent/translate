package translatetest

import (
	"go.expect.digital/translate/pkg/model"
	"go.expect.digital/translate/pkg/testutil/rand"
	"golang.org/x/text/language"
)

// RandMessages returns a random messages model with the given count of messages and source language.
// The messages will not be fuzzy.
func RandMessages(msgCount uint, srcLang language.Tag) *model.Messages {
	msgOpts := []rand.ModelMessageOption{rand.WithFuzzy(false)}
	msgsOpts := []rand.ModelMessagesOption{rand.WithLanguage(srcLang)}

	return rand.ModelMessages(msgCount, msgOpts, msgsOpts...)
}

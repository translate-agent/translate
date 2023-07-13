package awstranslate

import (
	"context"
	"strings"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.expect.digital/translate/pkg/model"
	"go.expect.digital/translate/pkg/testutil/rand"
	"golang.org/x/text/language"
)

func Test_Translate(t *testing.T) {
	t.Parallel()

	// t.Skip()

	ctx := context.Background()

	viper.SetEnvPrefix("translate")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	viper.AutomaticEnv()

	client, err := NewTranslate(ctx, WithDefaultClient(ctx))
	require.NoError(t, err)

	targetLang := language.LatinAmericanSpanish

	messages := randMessages(3, targetLang)

	msgs, err := client.Translate(ctx, messages)

	require.NoError(t, err)
	assert.NotEmpty(t, msgs)
}

// randMessages returns a random messages model with the given count of messages and source language.
// The messages will not be fuzzy.
func randMessages(msgCount uint, targetLang language.Tag) *model.Messages {
	msgOpts := []rand.ModelMessageOption{rand.WithStatus(model.MessageStatusUntranslated), rand.WithMessage("收发短信")}
	msgsOpts := []rand.ModelMessagesOption{rand.WithLanguage(targetLang)}

	return rand.ModelMessages(msgCount, msgOpts, msgsOpts...)
}

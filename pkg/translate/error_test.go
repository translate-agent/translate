package translate

import (
	"fmt"
	"math/rand"
	"reflect"
	"testing"
	"testing/quick"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/assert"
)

func Test_GetOriginalErr(t *testing.T) {
	t.Parallel()

	createWrappedErr := func(n int, originalErr error) error {
		wrapped := originalErr

		for i := 0; i < n; i++ {
			wrapped = fmt.Errorf("%s: %w", gofakeit.SentenceSimple(), wrapped)
		}

		return wrapped
	}

	conf := &quick.Config{
		MaxCount: 100,
		Values: func(args []reflect.Value, _ *rand.Rand) {
			originalErr := gofakeit.Error()
			args[0] = reflect.ValueOf(originalErr)
			args[1] = reflect.ValueOf(createWrappedErr(gofakeit.IntRange(0, 20), originalErr))
		},
	}

	f := func(original, wrapped error) bool {
		return assert.Equal(t, original, getOriginalErr(wrapped))
	}

	assert.NoError(t, quick.Check(f, conf))
}

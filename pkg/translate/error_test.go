package translate

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"math/rand"
	"reflect"
	"testing"
	"testing/quick"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.expect.digital/translate/pkg/repo"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/runtime/protoiface"
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

func Test_NewStatusWithDetails(t *testing.T) {
	t.Parallel()

	randCodeAndMsg := func() (codes.Code, string) {
		return codes.Code(gofakeit.IntRange(1, 15)), gofakeit.SentenceSimple()
	}

	t.Run("With Details", func(t *testing.T) {
		t.Parallel()

		detailCount := gofakeit.IntRange(1, 10)
		details := make([]protoiface.MessageV1, 0, detailCount)

		for i := 0; i < detailCount; i++ {
			details = append(details, &errdetails.BadRequest_FieldViolation{
				Field:       gofakeit.Word(),
				Description: gofakeit.SentenceSimple(),
			})
		}

		c, msg := randCodeAndMsg()

		st, err := newStatusWithDetails(c, msg, details...)
		require.NoError(t, err)

		require.Equal(t, c, st.Code())
		require.Equal(t, msg, st.Message())
		assert.Len(t, st.Details(), detailCount)
	})

	t.Run("Without Details", func(t *testing.T) {
		t.Parallel()

		c, msg := randCodeAndMsg()

		st, err := newStatusWithDetails(c, msg)
		require.NoError(t, err)

		require.Equal(t, c, st.Code())
		require.Equal(t, msg, st.Message())
		assert.Empty(t, st.Details())
	})
}

func Test_RequestErrorToStatusErr(t *testing.T) {
	t.Parallel()

	t.Run("InvalidArgument", func(t *testing.T) {
		t.Parallel()
		err := &fieldViolationError{
			field: gofakeit.Word(),
			err:   gofakeit.Error(),
		}

		assert.ErrorContains(t, requestErrorToStatusErr(err), codes.InvalidArgument.String())
	})

	t.Run("Unknown", func(t *testing.T) {
		t.Parallel()
		err := gofakeit.Error()

		assert.ErrorContains(t, requestErrorToStatusErr(err), codes.Unknown.String())
	})
}

func Test_RepoErrorToStatusErr(t *testing.T) {
	t.Parallel()

	t.Run("NotFound", func(t *testing.T) {
		t.Parallel()
		err := &repo.NotFoundError{
			Entity: gofakeit.Word(),
		}

		assert.ErrorContains(t, repoErrorToStatusErr(err), codes.NotFound.String())
	})

	t.Run("Internal", func(t *testing.T) {
		t.Parallel()
		err := &repo.DefaultError{
			Entity:    gofakeit.Word(),
			Operation: gofakeit.Word(),
		}

		assert.ErrorContains(t, repoErrorToStatusErr(err), codes.Internal.String())
	})

	t.Run("Unknown", func(t *testing.T) {
		t.Parallel()
		err := gofakeit.Error()

		assert.ErrorContains(t, repoErrorToStatusErr(err), codes.Unknown.String())
	})
}

func Test_ConvertToErrorToStatusErr(t *testing.T) {
	t.Parallel()

	t.Run("JSON SyntaxError", func(t *testing.T) {
		t.Parallel()
		err := &convertError{
			err:   &json.SyntaxError{},
			field: gofakeit.Word(),
		}

		assert.ErrorContains(t, convertFromErrorToStatusErr(err), codes.InvalidArgument.String())
	})

	t.Run("XML SyntaxError", func(t *testing.T) {
		t.Parallel()
		err := &convertError{
			err:   &xml.SyntaxError{},
			field: gofakeit.Word(),
		}

		assert.ErrorContains(t, convertFromErrorToStatusErr(err), codes.InvalidArgument.String())
	})

	t.Run("Random Error", func(t *testing.T) {
		t.Parallel()
		err := &convertError{
			err:   gofakeit.Error(),
			field: gofakeit.Word(),
		}

		assert.ErrorContains(t, convertFromErrorToStatusErr(err), codes.InvalidArgument.String())
	})
}

func Test_ConvertFromErrorToStatusErr(t *testing.T) {
	t.Parallel()

	err := &convertError{
		err:   gofakeit.Error(),
		field: gofakeit.Word(),
	}

	assert.ErrorContains(t, convertToErrorToStatusErr(err), codes.Internal.String())
}

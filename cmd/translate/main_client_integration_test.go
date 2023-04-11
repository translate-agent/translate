package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.expect.digital/translate/cmd/client/cmd"
)

//nolint:paralleltest
func Test_ServiceLs(t *testing.T) {
	t.Run("OK", func(t *testing.T) {
		res, err := cmd.ExecuteWithParams([]string{
			"service", "ls",
			"-a", fmt.Sprintf("%s:%s", host, port),
			"-i", "true",
		})

		if !assert.NoError(t, err) {
			return
		}

		assert.Contains(t, string(res), "TOTAL")
	})

	t.Run("error, no transport security set", func(t *testing.T) {
		res, err := cmd.ExecuteWithParams([]string{
			"service", "ls",
			"-a", fmt.Sprintf("%s:%s", host, port),
		})

		assert.ErrorContains(t, err, "no transport security set")
		assert.Nil(t, res)
	})
}

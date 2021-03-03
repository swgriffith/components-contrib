// ------------------------------------------------------------
// Copyright (c) Microsoft Corporation and Dapr Contributors.
// Licensed under the MIT License.
// ------------------------------------------------------------

package localstorage

import (
	"fmt"
	"testing"

	"github.com/dapr/components-contrib/state"
	"github.com/dapr/dapr/pkg/logger"
	"github.com/stretchr/testify/assert"
)

func TestInit(t *testing.T) {
	m := state.Metadata{}
	s := NewLocalStorageStore(logger.NewLogger("logger"))
	t.Run("Init with valid metadata", func(t *testing.T) {
		m.Properties = map[string]string{
			"hostPath": "/temp",
		}
		err := s.Init(m)
		assert.Nil(t, err)
		assert.Equal(t, "/temp", s.hostPath)
	})

	t.Run("Init with missing metadata", func(t *testing.T) {
		m.Properties = map[string]string{
			"invalidValue": "a",
		}
		err := s.Init(m)
		assert.NotNil(t, err)
		assert.Equal(t, err, fmt.Errorf("missing or empty hostPath field from metadata"))
	})
}

func TestFileName(t *testing.T) {
	t.Run("Valid composite key", func(t *testing.T) {
		key := getFileName("app_id||key")
		assert.Equal(t, "key", key)
	})

	t.Run("No delimiter present", func(t *testing.T) {
		key := getFileName("key")
		assert.Equal(t, "key", key)
	})
}

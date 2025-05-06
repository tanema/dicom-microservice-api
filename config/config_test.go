package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad(t *testing.T) {
	t.Parallel()
	t.Run("it should load dev", func(t *testing.T) {
		t.Parallel()
		cfg, err := Load("dev")
		require.NoError(t, err)
		assert.Equal(t, "development", cfg.Env)
	})
	t.Run("it should load staging", func(t *testing.T) {
		t.Parallel()
		cfg, err := Load("stag")
		require.NoError(t, err)
		assert.Equal(t, "staging", cfg.Env)
	})
	t.Run("it should load prod", func(t *testing.T) {
		t.Parallel()
		cfg, err := Load("prod")
		require.NoError(t, err)
		assert.Equal(t, "production", cfg.Env)
	})
	t.Run("it should load test", func(t *testing.T) {
		t.Parallel()
		cfg, err := Load("test")
		require.NoError(t, err)
		assert.Equal(t, "testing", cfg.Env)
	})
	t.Run("it should fail otherwise", func(t *testing.T) {
		t.Parallel()
		_, err := Load("other")
		require.Error(t, err)
	})
}

package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/oscarhfnorris/git-hunk/internal/config"
)

func TestDefault(t *testing.T) {
	cfg := config.Default()
	assert.Equal(t, "gpt-4o-mini", cfg.Model)
	assert.False(t, cfg.AutoGenerate)
	assert.Empty(t, cfg.APIKey)
}

func TestLoad_FileNotExist(t *testing.T) {
	cfg, err := config.Load("/nonexistent/path/config.toml")
	require.NoError(t, err)
	assert.Equal(t, config.Default(), cfg)
}

func TestLoad_ValidFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	content := `api_key = "sk-test"
model = "gpt-4o"
auto_generate = true
`
	require.NoError(t, os.WriteFile(path, []byte(content), 0o600))

	cfg, err := config.Load(path)
	require.NoError(t, err)
	assert.Equal(t, "sk-test", cfg.APIKey)
	assert.Equal(t, "gpt-4o", cfg.Model)
	assert.True(t, cfg.AutoGenerate)
}

func TestLoad_InvalidTOML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	require.NoError(t, os.WriteFile(path, []byte("not [ valid toml !!!"), 0o600))

	_, err := config.Load(path)
	assert.Error(t, err)
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     config.Config
		wantErr bool
	}{
		{
			name:    "valid default",
			cfg:     config.Default(),
			wantErr: false,
		},
		{
			name:    "empty model",
			cfg:     config.Config{Model: ""},
			wantErr: true,
		},
		{
			name:    "custom valid",
			cfg:     config.Config{APIKey: "key", Model: "gpt-4o", AutoGenerate: true},
			wantErr: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.cfg.Validate()
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

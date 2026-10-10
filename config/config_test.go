package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadConfig(t *testing.T) {
	t.Setenv("ANNICT_CLIENT_ID", "test-annict-id")
	t.Setenv("ANNICT_CLIENT_SECRET", "test-annict-secret")
	t.Setenv("ANILIST_CLIENT_ID", "test-anilist-id")
	t.Setenv("ANILIST_CLIENT_SECRET", "test-anilist-secret")
	t.Setenv("TOKEN_DIRECTORY", "/tmp/test-token")

	cfg, err := LoadConfig()
	require.NoError(t, err)
	assert.Equal(t, "test-annict-id", cfg.AnnictClientID)
	assert.Equal(t, "/tmp/test-token", cfg.TokenDirectory)
}

func TestLoadConfig_FromEnvFile(t *testing.T) {
	dir := t.TempDir()
	envPath := filepath.Join(dir, "custom.env")
	content := `ANNICT_CLIENT_ID=custom-annict-id
ANNICT_CLIENT_SECRET=custom-annict-secret
ANILIST_CLIENT_ID=custom-anilist-id
ANILIST_CLIENT_SECRET=custom-anilist-secret
TOKEN_DIRECTORY=/app/custom-token
`
	require.NoError(t, os.WriteFile(envPath, []byte(content), 0600))

	require.NoError(t, os.Setenv("ANNICT_CLIENT_ID", "custom-annict-id"))
	require.NoError(t, os.Setenv("ANNICT_CLIENT_SECRET", "custom-annict-secret"))
	require.NoError(t, os.Setenv("ANILIST_CLIENT_ID", "custom-anilist-id"))
	require.NoError(t, os.Setenv("ANILIST_CLIENT_SECRET", "custom-anilist-secret"))

	cfg, err := LoadConfig()
	require.NoError(t, err)
	assert.NotEmpty(t, cfg.AnnictClientID)
}

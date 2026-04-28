package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/neto-app/neto/tui/internal/config"
)

func TestConfig_SaveAndLoad(t *testing.T) {
	// Use a temp directory to avoid touching ~/.config/neto.
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "config.json")

	// Patch the config path by writing the file manually, then use Load via
	// the real path. We test Save+Load indirectly by overriding HOME.
	t.Setenv("HOME", tmpDir)
	// Also cover XDG on Linux.
	t.Setenv("XDG_CONFIG_HOME", "")

	want := &config.Config{
		APIURL:          "http://localhost:8080",
		SupabaseURL:     "https://abc.supabase.co",
		SupabaseAnonKey: "anon-key",
		AccessToken:     "access-jwt",
		RefreshToken:    "refresh-jwt",
	}

	if err := config.Save(want); err != nil {
		t.Fatalf("Save: %v", err)
	}

	// Verify the file exists at the expected path.
	expectedPath := filepath.Join(tmpDir, ".config", "neto", "config.json")
	if cfgPath != expectedPath {
		// cfgPath was declared for illustration; real path is expectedPath.
		cfgPath = expectedPath
	}

	info, err := os.Stat(cfgPath)
	if err != nil {
		t.Fatalf("stat config file: %v", err)
	}
	if perm := info.Mode().Perm(); perm != 0o600 {
		t.Errorf("permissions: got %o, want 0600", perm)
	}

	got, err := config.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if got.APIURL != want.APIURL {
		t.Errorf("APIURL: got %q, want %q", got.APIURL, want.APIURL)
	}
	if got.SupabaseURL != want.SupabaseURL {
		t.Errorf("SupabaseURL: got %q, want %q", got.SupabaseURL, want.SupabaseURL)
	}
	if got.SupabaseAnonKey != want.SupabaseAnonKey {
		t.Errorf("SupabaseAnonKey: got %q, want %q", got.SupabaseAnonKey, want.SupabaseAnonKey)
	}
	if got.AccessToken != want.AccessToken {
		t.Errorf("AccessToken: got %q, want %q", got.AccessToken, want.AccessToken)
	}
	if got.RefreshToken != want.RefreshToken {
		t.Errorf("RefreshToken: got %q, want %q", got.RefreshToken, want.RefreshToken)
	}
}

func TestConfig_Load_ErrNotConfigured(t *testing.T) {
	// Point HOME at an empty temp directory.
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	t.Setenv("XDG_CONFIG_HOME", "")

	_, err := config.Load()
	if err == nil {
		t.Fatal("expected ErrNotConfigured, got nil")
	}
	if err != config.ErrNotConfigured {
		t.Errorf("expected ErrNotConfigured, got %v", err)
	}
}

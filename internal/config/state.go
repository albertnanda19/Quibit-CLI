package config

import (
	"os"
	"path/filepath"
	"strings"
)

func SplashDisabledByEnv() bool {
	v := strings.TrimSpace(os.Getenv("QUIBIT_NO_SPLASH"))
	if v == "" {
		return false
	}
	v = strings.ToLower(v)
	return v != "0" && v != "false"
}

func HasSeenSplash() bool {
	p, ok := splashMarkerPath()
	if !ok {
		return false
	}
	_, err := os.Stat(p)
	return err == nil
}

func MarkSplashSeen() error {
	p, ok := splashMarkerPath()
	if !ok {
		return nil
	}
	dir := filepath.Dir(p)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil
	}
	return os.WriteFile(p, []byte("seen\n"), 0o644)
}

func splashMarkerPath() (string, bool) {
	stateHome := strings.TrimSpace(os.Getenv("XDG_STATE_HOME"))
	if stateHome == "" {
		home, err := os.UserHomeDir()
		if err != nil || strings.TrimSpace(home) == "" {
			return "", false
		}
		stateHome = filepath.Join(home, ".local", "state")
	}
	return filepath.Join(stateHome, "quibit", "splash_seen"), true
}

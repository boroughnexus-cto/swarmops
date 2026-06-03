package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDeployScriptPath(t *testing.T) {
	// Valid deploy dir with scripts/deploy.sh present.
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "scripts"), 0o755); err != nil {
		t.Fatal(err)
	}
	script := filepath.Join(dir, "scripts", "deploy.sh")
	if err := os.WriteFile(script, []byte("#!/usr/bin/env bash\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	got, err := deployScriptPath(dir)
	if err != nil {
		t.Fatalf("deployScriptPath(valid) error: %v", err)
	}
	if got != script {
		t.Errorf("deployScriptPath = %q, want %q", got, script)
	}

	// Empty dir argument.
	if _, err := deployScriptPath(""); err == nil {
		t.Error("deployScriptPath(\"\") should error")
	}

	// Non-existent dir.
	if _, err := deployScriptPath(filepath.Join(dir, "nope")); err == nil {
		t.Error("deployScriptPath(missing dir) should error")
	}

	// Dir exists but no scripts/deploy.sh.
	bare := t.TempDir()
	if _, err := deployScriptPath(bare); err == nil {
		t.Error("deployScriptPath(dir without scripts/deploy.sh) should error")
	}

	// A file (not a dir) passed as deploy dir.
	if _, err := deployScriptPath(script); err == nil {
		t.Error("deployScriptPath(file) should error")
	}
}

func TestDeployDirPathHonorsEnv(t *testing.T) {
	t.Setenv("SWARMOPS_DEPLOY_DIR", "/custom/deploy/dir")
	if got := deployDirPath(); got != "/custom/deploy/dir" {
		t.Errorf("deployDirPath() with env = %q, want /custom/deploy/dir", got)
	}
	t.Setenv("SWARMOPS_DEPLOY_DIR", "")
	home, _ := os.UserHomeDir()
	if got := deployDirPath(); got != filepath.Join(home, "swarmops") {
		t.Errorf("deployDirPath() default = %q, want %q", got, filepath.Join(home, "swarmops"))
	}
}

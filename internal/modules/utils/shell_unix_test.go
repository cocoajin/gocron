//go:build !windows

package utils

import (
	"context"
	"testing"
	"time"
)

func TestShellCancelledBeforeStart(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := ExecShell(ctx, "echo should-not-run"); err == nil {
		t.Fatal("cancelled command ran")
	}
}
func TestShellTimeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	started := time.Now()
	if _, err := ExecShell(ctx, "sleep 5"); err == nil {
		t.Fatal("timeout reported success")
	}
	if time.Since(started) > 2*time.Second {
		t.Fatal("process group not terminated")
	}
}

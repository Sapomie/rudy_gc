package filetool

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestCopyFileAtomicCtxCancelledLeavesNoTargetOrTemporaryFile(t *testing.T) {
	dir := t.TempDir()
	source := filepath.Join(dir, "source.bin")
	target := filepath.Join(dir, "target.bin")
	if err := os.WriteFile(source, []byte("content"), 0o644); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := CopyFileAtomicCtx(ctx, source, target, nil); err == nil {
		t.Fatal("expected cancellation error")
	}
	if _, err := os.Stat(target); !os.IsNotExist(err) {
		t.Fatalf("target stat error = %v", err)
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0].Name() != "source.bin" {
		t.Fatalf("unexpected files after cancellation: %#v", entries)
	}
}

func TestCopyFileAtomicCtxPublishesCompleteFileWithSourcePermissions(t *testing.T) {
	dir := t.TempDir()
	source := filepath.Join(dir, "source.bin")
	target := filepath.Join(dir, "target.bin")
	if err := os.WriteFile(source, []byte("complete-content"), 0o640); err != nil {
		t.Fatal(err)
	}
	if err := CopyFileAtomicCtx(context.Background(), source, target, nil); err != nil {
		t.Fatal(err)
	}
	content, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "complete-content" {
		t.Fatalf("target content = %q", content)
	}
	info, err := os.Stat(target)
	if err != nil {
		t.Fatal(err)
	}
	if got := info.Mode().Perm(); got != 0o640 {
		t.Fatalf("target permissions = %o", got)
	}
}

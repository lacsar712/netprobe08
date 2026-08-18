package twamp

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestWrapModeDeniedIs(t *testing.T) {
	err := WrapModeDenied("negotiate", "encrypted")
	if !errors.Is(err, ErrMode) {
		t.Fatalf("lost ErrMode: %v", err)
	}
}

func TestDumpSessionLogPersists(t *testing.T) {
	path := filepath.Join(t.TempDir(), "session.log")
	body := "mode=authenticated pad=128\n"
	if err := DumpSessionLog(path, body); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != body {
		t.Fatalf("got %q", b)
	}
}

func TestExportSessionFileRejectsEscape(t *testing.T) {
	if _, err := ExportSessionFile(t.TempDir(), filepath.Join("..", "hosts")); err == nil {
		t.Fatal("expected path escape to be rejected")
	}
}

func TestCopyPaddingIndependent(t *testing.T) {
	src := []byte{1, 2, 3, 4, 5}
	got := CopyPadding(src, 3)
	got[0] = 9
	if src[0] != 1 {
		t.Fatal("CopyPadding aliased the TWAMP pad buffer")
	}
}

func TestSessionBagSetPad(t *testing.T) {
	bag := NewSessionBag()
	bag.SetPad("pe-a-to-pe-b", 64)
	if bag.Pad("pe-a-to-pe-b") != 64 {
		t.Fatal("pad not stored")
	}
}

func TestWaitReflectorHonorsCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(30 * time.Millisecond)
		cancel()
	}()
	start := time.Now()
	err := WaitReflector(ctx, 600*time.Millisecond)
	if err == nil {
		t.Fatal("expected cancel error")
	}
	if time.Since(start) > 250*time.Millisecond {
		t.Fatalf("WaitReflector ignored cancel, elapsed=%s", time.Since(start))
	}
}

func TestParseTWAMPJSONRejectsInvalid(t *testing.T) {
	if _, err := ParseTWAMPJSON([]byte("pad=64")); err == nil {
		t.Fatal("expected JSON error")
	}
}

func TestAfterWriteRejectsPadShrink(t *testing.T) {
	min := ""
	get := func() (string, error) { return min, nil }
	set := func(v string) error { min = v; return nil }
	if err := AfterWrite(get, set, "mode=authenticated pad=128"); err != nil {
		t.Fatal(err)
	}
	if err := AfterWrite(get, set, "mode=authenticated pad=32"); err == nil {
		t.Fatal("expected pad shrink below last advertised value to be rejected")
	}
}

func TestGrowPadNoWriteThrough(t *testing.T) {
	dst := make([]byte, 2, 8)
	copy(dst, []byte("AB"))
	got := GrowPad(dst, 'C')
	got[0] = 'X'
	if dst[0] != 'A' {
		t.Fatal("GrowPad wrote through into the pad buffer")
	}
}

func TestNilProbePeerName(t *testing.T) {
	var p *Probe
	if p.PeerName() != "" {
		t.Fatalf("got %q", p.PeerName())
	}
}

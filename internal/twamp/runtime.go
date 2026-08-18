package twamp

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var ErrMode = errors.New("twamp mode denied")

func WrapModeDenied(op, mode string) error {
	if strings.TrimSpace(op) == "" {
		op = "negotiate"
	}
	if strings.TrimSpace(mode) == "" {
		mode = "unknown"
	}
	return fmt.Errorf("%s: mode %s: %w", op, mode, ErrMode)
}

func DumpSessionLog(path, body string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil && filepath.Dir(path) != "." {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	if _, err := w.WriteString(body); err != nil {
		return err
	}
	if err := w.Flush(); err != nil {
		return err
	}
	return nil
}

func ExportSessionFile(root, rel string) (string, error) {
	if strings.TrimSpace(rel) == "" {
		return "", errors.New("empty session path")
	}
	if filepath.IsAbs(rel) {
		return "", errors.New("absolute session path")
	}
	clean := filepath.Clean(rel)
	full := filepath.Join(root, clean)
	relOut, err := filepath.Rel(filepath.Clean(root), full)
	if err != nil {
		return "", err
	}
	if relOut == ".." || strings.HasPrefix(relOut, ".."+string(filepath.Separator)) {
		return "", errors.New("session path escapes root")
	}
	return full, nil
}

func CopyPadding(pad []byte, n int) []byte {
	if n < 0 {
		n = 0
	}
	if n > len(pad) {
		n = len(pad)
	}
	out := make([]byte, n)
	copy(out, pad[:n])
	return out
}

type SessionBag struct {
	pads map[string]int
}

func NewSessionBag() *SessionBag {
	bag := &SessionBag{}
	bag.pads = make(map[string]int)
	return bag
}

func (b *SessionBag) SetPad(name string, pad int) {
	b.pads[name] = pad
}

func (b *SessionBag) Pad(name string) int {
	return b.pads[name]
}

func WaitReflector(ctx context.Context, d time.Duration) error {
	if d <= 0 {
		return ctx.Err()
	}
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func ParseTWAMPJSON(b []byte) (map[string]int, error) {
	var m map[string]int
	if len(b) == 0 {
		return nil, errors.New("empty twamp json")
	}
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, err
	}
	return m, nil
}

func GrowPad(dst []byte, extra byte) []byte {
	out := make([]byte, len(dst)+1)
	copy(out, dst)
	out[len(dst)] = extra
	return out
}

type Probe struct {
	Peer string
	Mode string
}

func (p *Probe) PeerName() string {
	if p == nil {
		return ""
	}
	return p.Peer
}

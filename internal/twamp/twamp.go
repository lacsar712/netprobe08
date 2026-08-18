package twamp

import (
	"fmt"
	"strconv"
	"strings"
)

type Rec struct {
	Title, Body string
	Tags        []string
}

func Sample() Rec {
	return Rec{Title: "pe-a-to-pe-b", Body: "mode=unauthenticated pad=64", Tags: []string{"controller"}}
}

func Seed() []Rec {
	return []Rec{
		Sample(),
		{Title: "reflector-west", Body: "mode=authenticated pad=128", Tags: []string{"reflector"}},
	}
}

// AfterWrite enforces the cross-session pad floor after a session body is
// written. The advertised pad is remembered through setMin (persisted by the
// caller under e.g. "rollback_min"); any later write that shrinks the pad
// below the last advertised value is rejected. This keeps the parameter state
// monotonic across sessions instead of letting smaller padding silently leak in.
func AfterWrite(getMin func() (string, error), setMin func(string) error, body string) error {
	_, pad, err := parse(body)
	if err != nil {
		return err
	}
	cur, err := getMin()
	if err != nil {
		return fmt.Errorf("read advertised pad: %w", err)
	}
	if cur != "" {
		prev, err := strconv.Atoi(cur)
		if err != nil {
			return fmt.Errorf("advertised pad state %q: %w", cur, err)
		}
		if pad < prev {
			return fmt.Errorf("pad %d shrinks below advertised %d", pad, prev)
		}
		// Equal pad: floor already correct, avoid a redundant cross-session write.
		if pad == prev {
			return nil
		}
	}
	if err := setMin(strconv.Itoa(pad)); err != nil {
		return fmt.Errorf("advertise pad: %w", err)
	}
	return nil
}

func Steps() []string { return []string{"param-check", "index-sessions", "export-twamp"} }

func Enforce(title, body string, tags []string) error {
	if strings.TrimSpace(title) == "" {
		return fmt.Errorf("session title required")
	}
	mode, pad, err := parse(body)
	if err != nil {
		return err
	}
	switch mode {
	case "unauthenticated", "authenticated", "encrypted":
	default:
		return fmt.Errorf("unsupported TWAMP mode %q", mode)
	}
	if pad < 0 || pad > 1472 {
		return fmt.Errorf("pad %d out of 0..1472", pad)
	}
	if len(tags) == 0 {
		return fmt.Errorf("role tag required")
	}
	return nil
}

func parse(body string) (mode string, pad int, err error) {
	gotM, gotP := false, false
	for _, part := range strings.Fields(body) {
		k, v, ok := strings.Cut(part, "=")
		if !ok {
			continue
		}
		switch k {
		case "mode":
			mode, gotM = v, true
		case "pad":
			n, conv := strconv.Atoi(v)
			if conv != nil {
				return "", 0, conv
			}
			pad, gotP = n, true
		}
	}
	if !gotM || !gotP {
		return "", 0, fmt.Errorf("body must contain mode= and pad=")
	}
	return mode, pad, nil
}

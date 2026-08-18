package twamp

import "testing"

// afterWriteState is a tiny stand-in for the persisted "rollback_min"
// setting, exercising AfterWrite the same way svc.Catalog wires it to the
// store: get returns the current string, set stores it.
type afterWriteState struct {
	min string
}

func (s *afterWriteState) get() (string, error) { return s.min, nil }
func (s *afterWriteState) set(v string) error   { s.min = v; return nil }

// TestAfterWriteRemembersAdvertisedPad is the regression for the pad-freeze
// bug: AfterWrite must persist the advertised pad so that a smaller pad on a
// later write (even across an independent state object) is rejected.
func TestAfterWriteRemembersAdvertisedPad(t *testing.T) {
	st := &afterWriteState{}
	if err := AfterWrite(st.get, st.set, "mode=authenticated pad=128"); err != nil {
		t.Fatalf("first advertise: %v", err)
	}
	if st.min != "128" {
		t.Fatalf("advertised pad not remembered: got %q want %q", st.min, "128")
	}
	if err := AfterWrite(st.get, st.set, "mode=authenticated pad=32"); err == nil {
		t.Fatal("pad shrink below advertised value was accepted")
	}
	if st.min != "128" {
		t.Fatalf("floor mutated after reject: got %q want %q", st.min, "128")
	}
}

// TestAfterWriteGrowsFloorOnLargerPad ensures a larger advertised pad raises
// the floor, so a subsequent value between the old and new floor is rejected.
func TestAfterWriteGrowsFloorOnLargerPad(t *testing.T) {
	st := &afterWriteState{}
	if err := AfterWrite(st.get, st.set, "mode=unauthenticated pad=64"); err != nil {
		t.Fatalf("advertise 64: %v", err)
	}
	if err := AfterWrite(st.get, st.set, "mode=unauthenticated pad=256"); err != nil {
		t.Fatalf("grow to 256: %v", err)
	}
	if st.min != "256" {
		t.Fatalf("floor not raised: got %q want %q", st.min, "256")
	}
	if err := AfterWrite(st.get, st.set, "mode=unauthenticated pad=128"); err == nil {
		t.Fatal("value below new floor was accepted")
	}
}

// TestAfterWriteAcceptsEqualPad ensures re-advertising the same pad is allowed
// and does not needlessly mutate the persisted floor.
func TestAfterWriteAcceptsEqualPad(t *testing.T) {
	st := &afterWriteState{}
	if err := AfterWrite(st.get, st.set, "mode=encrypted pad=512"); err != nil {
		t.Fatalf("advertise 512: %v", err)
	}
	if err := AfterWrite(st.get, st.set, "mode=encrypted pad=512"); err != nil {
		t.Fatalf("re-advertise 512: %v", err)
	}
	if st.min != "512" {
		t.Fatalf("floor changed on equal pad: got %q want %q", st.min, "512")
	}
}

// TestAfterWriteFirstWriteFromEmptyFloor ensures an empty stored floor is
// treated as "no advertisement yet" and accepts the first pad.
func TestAfterWriteFirstWriteFromEmptyFloor(t *testing.T) {
	st := &afterWriteState{min: ""}
	if err := AfterWrite(st.get, st.set, "mode=unauthenticated pad=0"); err != nil {
		t.Fatalf("first write from empty floor: %v", err)
	}
	if st.min != "0" {
		t.Fatalf("floor not set on first write: got %q want %q", st.min, "0")
	}
}

// TestAfterWriteRejectsBadBody ensures a body missing pad= is not silently
// accepted and never touches the persisted floor.
func TestAfterWriteRejectsBadBody(t *testing.T) {
	st := &afterWriteState{min: "128"}
	if err := AfterWrite(st.get, st.set, "mode=authenticated"); err == nil {
		t.Fatal("body without pad= was accepted")
	}
	if st.min != "128" {
		t.Fatalf("floor mutated on bad body: got %q want %q", st.min, "128")
	}
}

// TestAfterWriteFailsClosedOnGetError ensures that when the floor cannot be
// read, AfterWrite fails closed instead of accepting the write.
func TestAfterWriteFailsClosedOnGetError(t *testing.T) {
	get := func() (string, error) { return "", errSentinel }
	set := func(v string) error { return nil }
	if err := AfterWrite(get, set, "mode=authenticated pad=128"); err == nil {
		t.Fatal("expected error when getMin fails (fail-closed)")
	}
}

// TestAfterWritePropagatesSetError ensures a persistence failure from setMin is
// surfaced rather than swallowed.
func TestAfterWritePropagatesSetError(t *testing.T) {
	get := func() (string, error) { return "", nil }
	set := func(v string) error { return errSentinel }
	if err := AfterWrite(get, set, "mode=authenticated pad=128"); err == nil {
		t.Fatal("expected error when setMin fails")
	}
}

var errSentinel = newSentinelError()

func newSentinelError() error {
	return &sentinelErr{msg: "afterwrite: sentinel"}
}

type sentinelErr struct{ msg string }

func (e *sentinelErr) Error() string { return e.msg }

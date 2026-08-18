package twamp

import "testing"

func TestAccept(t *testing.T) {
	if err := Enforce("s", "mode=unauthenticated pad=64", []string{"controller"}); err != nil {
		t.Fatal(err)
	}
}

func TestRejectPad(t *testing.T) {
	if err := Enforce("s", "mode=unauthenticated pad=9000", []string{"controller"}); err == nil {
		t.Fatal("expected reject")
	}
}

package ids

import "testing"

func TestNew(t *testing.T) {
	id, err := New("usr")
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}
	if len(id) < 10 {
		t.Fatalf("unexpected id length: %s", id)
	}
}

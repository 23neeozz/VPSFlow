package token

import "testing"

func TestGenerateAndHash(t *testing.T) {
	value, err := GenerateOpaque(32)
	if err != nil {
		t.Fatalf("GenerateOpaque failed: %v", err)
	}
	if len(value) < 32 {
		t.Fatalf("token too short: %d", len(value))
	}

	hash := Hash(value)
	if hash == "" || hash == Hash("different") {
		t.Fatal("expected stable hash output")
	}
}

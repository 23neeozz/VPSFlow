package password

import "testing"

func TestHashAndVerify(t *testing.T) {
	hash, err := Hash("SecurePass123!")
	if err != nil {
		t.Fatalf("Hash failed: %v", err)
	}

	ok, err := Verify(hash, "SecurePass123!")
	if err != nil {
		t.Fatalf("Verify failed: %v", err)
	}
	if !ok {
		t.Fatal("expected password to verify")
	}

	ok, err = Verify(hash, "WrongPassword1!")
	if err != nil {
		t.Fatalf("Verify failed: %v", err)
	}
	if ok {
		t.Fatal("expected password mismatch")
	}
}

func TestValidate(t *testing.T) {
	if err := Validate("short"); err == nil {
		t.Fatal("expected validation error for short password")
	}
	if err := Validate("longpasswordonly"); err == nil {
		t.Fatal("expected validation error for missing complexity")
	}
	if err := Validate("SecurePass123!"); err != nil {
		t.Fatalf("expected valid password, got %v", err)
	}
}

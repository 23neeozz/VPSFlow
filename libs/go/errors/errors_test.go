package errors

import "testing"

func TestAppError(t *testing.T) {
	err := Wrap(CodeValidation, "invalid input", nil).WithDetails(map[string]any{"field": "email"})
	if !IsCode(err, CodeValidation) {
		t.Fatal("expected validation code")
	}
	if err.Details["field"] != "email" {
		t.Fatal("expected details preserved")
	}
}

func TestAsAppError(t *testing.T) {
	err := New(CodeNotFound, "resource missing")
	got, ok := AsAppError(err)
	if !ok || got.Code != CodeNotFound {
		t.Fatal("expected not found error")
	}
}

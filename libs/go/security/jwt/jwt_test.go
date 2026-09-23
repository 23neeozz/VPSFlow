package jwt

import (
	"testing"
	"time"
)

func TestIssueAndParseAccessToken(t *testing.T) {
	svc, err := NewService(Config{
		SigningKey: []byte("dev-only-change-in-production-min-32-chars"),
		AccessTTL:  time.Minute,
	})
	if err != nil {
		t.Fatalf("NewService failed: %v", err)
	}

	token, _, err := svc.IssueAccessToken("usr_1", "ses_1", "user@example.com")
	if err != nil {
		t.Fatalf("IssueAccessToken failed: %v", err)
	}

	claims, err := svc.ParseAccessToken(token)
	if err != nil {
		t.Fatalf("ParseAccessToken failed: %v", err)
	}
	if claims.UserID != "usr_1" || claims.SessionID != "ses_1" {
		t.Fatalf("unexpected claims: %+v", claims)
	}
}

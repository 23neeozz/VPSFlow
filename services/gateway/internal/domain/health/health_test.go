package health

import "testing"

func TestCheckerReportsHealthy(t *testing.T) {
	checker := NewChecker("gateway", "0.1.0")
	live := checker.Liveness()
	if live.Status != StatusHealthy {
		t.Fatalf("expected healthy liveness, got %s", live.Status)
	}
	ready := checker.Readiness()
	if ready.Status != StatusHealthy {
		t.Fatalf("expected healthy readiness, got %s", ready.Status)
	}
}

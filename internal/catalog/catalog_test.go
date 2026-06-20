package catalog

import "testing"

func TestRegisterAndGet(t *testing.T) {
	s := NewStore()
	_, err := s.Register(Service{Name: "api", Owner: "platform", Language: "go", Replicas: 2})
	if err != nil {
		t.Fatalf("register: %v", err)
	}
	got, err := s.Get("api")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Owner != "platform" {
		t.Errorf("owner = %q, want platform", got.Owner)
	}
	if _, err := s.Get("missing"); err != ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

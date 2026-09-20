package database

import "testing"

func TestCloseDB_NilPool(t *testing.T) {
	prev := DBPool
	DBPool = nil
	t.Cleanup(func() {
		DBPool = prev
	})

	// Should not panic
	CloseDB()
}

func TestInitDB_InvalidConfig(t *testing.T) {
	err := InitDB("://invalid-url")
	if err == nil {
		t.Fatal("expected error with malformed database url, got nil")
	}
}

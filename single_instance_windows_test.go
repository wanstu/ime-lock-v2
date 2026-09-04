//go:build windows

package main

import (
	"path/filepath"
	"testing"
)

func TestAcquireInstanceLock(t *testing.T) {
	lockPath := filepath.Join(t.TempDir(), "app.lock")
	release1, primary1, err := acquireInstanceLock(lockPath)
	if err != nil {
		t.Fatalf("first acquire error = %v", err)
	}
	if !primary1 {
		t.Fatal("first acquire should be primary")
	}
	defer release1()

	release2, primary2, err := acquireInstanceLock(lockPath)
	if err != nil {
		t.Fatalf("second acquire error = %v", err)
	}
	defer release2()
	if primary2 {
		t.Fatal("second acquire should not be primary")
	}

	release1()
	release3, primary3, err := acquireInstanceLock(lockPath)
	if err != nil {
		t.Fatalf("third acquire error = %v", err)
	}
	defer release3()
	if !primary3 {
		t.Fatal("lock should be acquirable after release")
	}
}

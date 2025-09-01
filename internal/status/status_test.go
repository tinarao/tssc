package status

import "testing"

func TestLocking(t *testing.T) {
	alias := "finland"
	err := Lock(alias)
	if err != nil {
		t.Fatal(err.Error())
	}
	defer Unlock()

	if !IsLocked() {
		t.Fatal("Not locked, but should be")
	}
}

package localfiles

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
)

func TestStoreRejectsTraversalAndUnapprovedTypes(t *testing.T) {
	store := New(t.TempDir(), 1024, ".json", ".txt")
	for _, name := range []string{"../outside.json", "payload.exe"} {
		if _, err := store.Save(context.Background(), name, []byte("data")); !errors.Is(err, ErrFileTypeDenied) {
			t.Fatalf("%s accepted: %v", name, err)
		}
	}
	got, err := store.Save(context.Background(), "template.json", []byte(`{"version":1}`))
	if err != nil || filepath.Base(got.Path) != "template.json" || got.SHA256 == "" {
		t.Fatalf("stored=%#v err=%v", got, err)
	}
}

package users_test

import (
	"testing"

	"novastream/models"
	"novastream/services/users"
)

func TestServiceInitialisesDefaultUser(t *testing.T) {
	svc, err := users.NewService(t.TempDir())
	if err != nil {
		t.Fatalf("failed to create service: %v", err)
	}

	list := svc.List()
	if len(list) != 1 {
		t.Fatalf("expected exactly one user, got %d", len(list))
	}

	if list[0].ID != models.DefaultUserID {
		t.Fatalf("expected default user id %q, got %q", models.DefaultUserID, list[0].ID)
	}
	if list[0].Name == "" {
		t.Fatalf("expected default user to have a name")
	}
}

func TestServiceCreateRenameAndDelete(t *testing.T) {
	svc, err := users.NewService(t.TempDir())
	if err != nil {
		t.Fatalf("failed to create service: %v", err)
	}

	created, err := svc.Create("Evening Watcher")
	if err != nil {
		t.Fatalf("create returned error: %v", err)
	}

	if created.ID == "" {
		t.Fatalf("expected created user to have id")
	}

	renamed, err := svc.Rename(created.ID, "Night Owl")
	if err != nil {
		t.Fatalf("rename returned error: %v", err)
	}

	if renamed.Name != "Night Owl" {
		t.Fatalf("expected renamed user to have updated name, got %q", renamed.Name)
	}

	if err := svc.Delete(created.ID); err != nil {
		t.Fatalf("delete returned error: %v", err)
	}

	if svc.Exists(created.ID) {
		t.Fatalf("expected user to be deleted")
	}
}

func TestDeleteDefaultUserFails(t *testing.T) {
	svc, err := users.NewService(t.TempDir())
	if err != nil {
		t.Fatalf("failed to create service: %v", err)
	}

	if err := svc.Delete(models.DefaultUserID); err == nil {
		t.Fatalf("expected delete to fail for default user")
	}
}

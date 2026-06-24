package config

import (
	"strings"
	"testing"
)

func TestExportRedactsSecrets(t *testing.T) {
	cfg := &Config{
		AdminPassword: "secret-admin",
		Mounts:        map[string]string{"/live": "hashed-mount-pw"},
		Users: map[string]*User{
			"admin": {Username: "admin", Password: "hashed-user-pw", Role: RoleSuperAdmin},
		},
		SMTP: &SMTPConfig{Enabled: true, Password: "smtp-secret"},
	}

	data, err := cfg.Export()
	if err != nil {
		t.Fatal(err)
	}
	out := string(data)
	for _, secret := range []string{"secret-admin", "hashed-mount-pw", "hashed-user-pw", "smtp-secret"} {
		if strings.Contains(out, secret) {
			t.Fatalf("export leaked secret %q", secret)
		}
	}
	if !strings.Contains(out, RedactedSecret) {
		t.Fatal("expected redacted placeholder in export")
	}
}

func TestMergeImportPreservesSecrets(t *testing.T) {
	orig := &Config{
		AdminPassword: "keep-me",
		Mounts:        map[string]string{"/live": "mount-hash"},
		Users: map[string]*User{
			"admin": {Username: "admin", Password: "user-hash", Role: RoleSuperAdmin},
		},
		PageTitle: "Original",
	}
	incoming := &Config{
		AdminPassword: RedactedSecret,
		Mounts:        map[string]string{"/live": RedactedSecret},
		Users: map[string]*User{
			"admin": {Username: "admin", Password: RedactedSecret, Role: RoleSuperAdmin},
		},
		PageTitle: "Imported",
	}
	if err := orig.MergeImport(incoming); err != nil {
		t.Fatal(err)
	}
	if orig.PageTitle != "Imported" {
		t.Fatalf("expected imported title, got %q", orig.PageTitle)
	}
	if orig.AdminPassword != "keep-me" {
		t.Fatalf("admin password not preserved")
	}
	if orig.Mounts["/live"] != "mount-hash" {
		t.Fatalf("mount password not preserved")
	}
	if orig.Users["admin"].Password != "user-hash" {
		t.Fatalf("user password not preserved")
	}
}

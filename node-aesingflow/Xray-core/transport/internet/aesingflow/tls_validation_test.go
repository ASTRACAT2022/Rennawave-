package aesingflow

import (
	"os"
	"testing"
)

func TestResolveCertificatePathRejectsWrongExtension(t *testing.T) {
	if _, err := resolveCertificatePath("/var/lib/remnawave/configs/xray/ssl/privkey.pem", ".key"); err == nil {
		t.Fatal("expected private key extension validation error")
	}
}

func TestValidateKeyPermissionsRejectsBroadMode(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/privkey.key"
	if err := os.WriteFile(path, []byte("private"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := validateKeyPermissions(path); err == nil {
		t.Fatal("expected broad private key permissions to be rejected")
	}
}

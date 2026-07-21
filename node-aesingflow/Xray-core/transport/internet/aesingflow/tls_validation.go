package aesingflow

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	xraytls "github.com/xtls/xray-core/transport/internet/tls"
)

const remnawaveSSLDir = "/var/lib/remnawave/configs/xray/ssl"

func validateTLS(config *xraytls.Config) error {
	if config == nil {
		return fmt.Errorf("streamSettings.security must be tls")
	}
	if config.MinVersion != "1.3" {
		return fmt.Errorf("AesingFlow requires tlsSettings.minVersion to be 1.3")
	}
	if !contains(config.NextProtocol, "aesingflow") {
		return fmt.Errorf("tlsSettings.alpn must include aesingflow")
	}
	if strings.TrimSpace(config.ServerName) == "" {
		return fmt.Errorf("tlsSettings.serverName is required for SAN validation")
	}
	if len(config.Certificate) == 0 {
		return fmt.Errorf("tlsSettings.certificates is required")
	}
	for _, certificate := range config.Certificate {
		if certificate == nil {
			return fmt.Errorf("tlsSettings contains an empty certificate")
		}
		certPath, err := resolveCertificatePath(certificate.CertificatePath, ".pem")
		if err != nil {
			return fmt.Errorf("certificateFile: %w", err)
		}
		keyPath, err := resolveCertificatePath(certificate.KeyPath, ".key")
		if err != nil {
			return fmt.Errorf("keyFile: %w", err)
		}
		if err := validateKeyPermissions(keyPath); err != nil {
			return err
		}
		certPEM, err := os.ReadFile(certPath)
		if err != nil {
			return fmt.Errorf("cannot read certificate file")
		}
		keyPEM, err := os.ReadFile(keyPath)
		if err != nil {
			return fmt.Errorf("cannot read private key file")
		}
		if block, _ := pem.Decode(certPEM); block == nil || block.Type != "CERTIFICATE" {
			return fmt.Errorf("certificate file is not valid PEM")
		}
		if block, _ := pem.Decode(keyPEM); block == nil || !strings.Contains(block.Type, "PRIVATE KEY") {
			return fmt.Errorf("private key file is not valid PEM")
		}
		pair, err := tls.X509KeyPair(certPEM, keyPEM)
		if err != nil {
			return fmt.Errorf("private key does not match certificate")
		}
		leaf, err := x509.ParseCertificate(pair.Certificate[0])
		if err != nil {
			return fmt.Errorf("certificate leaf is invalid")
		}
		if time.Now().After(leaf.NotAfter) {
			return fmt.Errorf("certificate has expired")
		}
		if !allowsServerAuth(leaf) {
			return fmt.Errorf("certificate is not valid for server authentication")
		}
		if err := leaf.VerifyHostname(config.ServerName); err != nil {
			return fmt.Errorf("certificate SAN does not match tlsSettings.serverName")
		}
	}
	return nil
}

func contains(values []string, required string) bool {
	for _, value := range values {
		if value == required {
			return true
		}
	}
	return false
}

func resolveCertificatePath(path, extension string) (string, error) {
	if filepath.Ext(path) != extension {
		return "", fmt.Errorf("must use %s extension", extension)
	}
	root, err := filepath.EvalSymlinks(remnawaveSSLDir)
	if err != nil {
		return "", fmt.Errorf("allowed Remnawave SSL directory is unavailable")
	}
	resolved, err := filepath.EvalSymlinks(path)
	if err != nil {
		return "", fmt.Errorf("file does not exist or contains an invalid symlink")
	}
	relative, err := filepath.Rel(root, resolved)
	if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) || filepath.IsAbs(relative) {
		return "", fmt.Errorf("path must remain inside %s", remnawaveSSLDir)
	}
	return resolved, nil
}

func validateKeyPermissions(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("cannot stat private key file")
	}
	if info.Mode().Perm()&0o077 != 0 {
		return fmt.Errorf("private key file permissions are too broad")
	}
	return nil
}

func allowsServerAuth(cert *x509.Certificate) bool {
	if len(cert.ExtKeyUsage) == 0 {
		return true
	}
	for _, usage := range cert.ExtKeyUsage {
		if usage == x509.ExtKeyUsageServerAuth || usage == x509.ExtKeyUsageAny {
			return true
		}
	}
	return false
}

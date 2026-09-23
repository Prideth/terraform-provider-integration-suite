package securitycontent

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/hex"
	"encoding/pem"
	"math/big"
	"testing"
	"time"
)

// generateTestCertificatePEM builds a synthetic, self-signed X.509
// certificate solely for this test — never a real-world certificate.
func generateTestCertificatePEM(t *testing.T, commonName string, notBefore, notAfter time.Time) ([]byte, *x509.Certificate) {
	t.Helper()

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("generating test key: %v", err)
	}

	template := &x509.Certificate{
		SerialNumber: big.NewInt(424242),
		Subject: pkix.Name{
			CommonName:         commonName,
			Organization:       []string{"Test Org"},
			OrganizationalUnit: []string{"Test Unit"},
			Country:            []string{"DE"},
		},
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	der, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		t.Fatalf("creating test certificate: %v", err)
	}

	cert, err := x509.ParseCertificate(der)
	if err != nil {
		t.Fatalf("parsing generated test certificate: %v", err)
	}

	pemBytes := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	return pemBytes, cert
}

func TestParseCertificatePEM(t *testing.T) {
	notBefore := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	notAfter := time.Date(2029, 1, 1, 0, 0, 0, 0, time.UTC)
	pemBytes, cert := generateTestCertificatePEM(t, "test.example.invalid", notBefore, notAfter)

	meta, err := ParseCertificatePEM(pemBytes)
	if err != nil {
		t.Fatalf("ParseCertificatePEM() error: %v", err)
	}

	wantSum := sha256.Sum256(cert.Raw)
	wantFingerprint := hex.EncodeToString(wantSum[:])
	if meta.SHA256Fingerprint != wantFingerprint {
		t.Errorf("SHA256Fingerprint = %q, want %q", meta.SHA256Fingerprint, wantFingerprint)
	}
	if meta.SerialNumber != "424242" {
		t.Errorf("SerialNumber = %q, want 424242", meta.SerialNumber)
	}
	if meta.SubjectDN == "" || meta.IssuerDN == "" {
		t.Error("SubjectDN/IssuerDN must not be empty")
	}
	if !meta.NotBefore.Equal(notBefore) {
		t.Errorf("NotBefore = %v, want %v", meta.NotBefore, notBefore)
	}
	if !meta.NotAfter.Equal(notAfter) {
		t.Errorf("NotAfter = %v, want %v", meta.NotAfter, notAfter)
	}
}

// TestParseCertificatePEM_FormattingIndependence proves the whole point of
// deriving a canonical fingerprint: two PEM encodings of the exact same
// certificate that differ only in line endings, wrapping, or a trailing
// newline produce the identical SHA-256 fingerprint.
func TestParseCertificatePEM_FormattingIndependence(t *testing.T) {
	pemBytes, _ := generateTestCertificatePEM(t, "formatting.example.invalid", time.Now(), time.Now().AddDate(1, 0, 0))

	original, err := ParseCertificatePEM(pemBytes)
	if err != nil {
		t.Fatalf("ParseCertificatePEM(original) error: %v", err)
	}

	// CRLF line endings and no trailing newline, simulating a file edited
	// on Windows or copy-pasted without a final blank line.
	crlf := []byte{}
	for _, b := range pemBytes {
		if b == '\n' {
			crlf = append(crlf, '\r', '\n')
		} else {
			crlf = append(crlf, b)
		}
	}
	for len(crlf) > 0 && (crlf[len(crlf)-1] == '\n' || crlf[len(crlf)-1] == '\r') {
		crlf = crlf[:len(crlf)-1]
	}

	reformatted, err := ParseCertificatePEM(crlf)
	if err != nil {
		t.Fatalf("ParseCertificatePEM(reformatted) error: %v", err)
	}

	if original.SHA256Fingerprint != reformatted.SHA256Fingerprint {
		t.Errorf("fingerprints differ purely due to PEM formatting: %q vs %q", original.SHA256Fingerprint, reformatted.SHA256Fingerprint)
	}
}

func TestParseCertificatePEM_RejectsNonPEM(t *testing.T) {
	if _, err := ParseCertificatePEM([]byte("not a certificate")); err == nil {
		t.Error("ParseCertificatePEM() error = nil, want an error for non-PEM content")
	}
}

func TestParseCertificatePEM_RejectsWrongPEMType(t *testing.T) {
	block := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: []byte("not really a key")})
	if _, err := ParseCertificatePEM(block); err == nil {
		t.Error("ParseCertificatePEM() error = nil, want an error for a non-CERTIFICATE PEM block")
	}
}

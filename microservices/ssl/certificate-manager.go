package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"log"
	"math/big"
	"net"
	"os"
	"time"
)

// CertificateManager handles TLS/SSL certificate operations
type CertificateManager struct {
	certDir string
}

// NewCertificateManager creates a new certificate manager
func NewCertificateManager(certDir string) *CertificateManager {
	if err := os.MkdirAll(certDir, 0755); err != nil {
		log.Printf("Failed to create cert directory: %v", err)
	}
	return &CertificateManager{certDir: certDir}
}

// GenerateSelfSignedCertificate generates a self-signed certificate
func (cm *CertificateManager) GenerateSelfSignedCertificate(hosts []string) error {
	// Generate private key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return fmt.Errorf("failed to generate private key: %v", err)
	}

	// Create certificate template
	template := x509.Certificate{
		SerialNumber: big.NewInt(time.Now().Unix()),
		Subject: pkix.Name{
			Organization: []string{"Financial Platform"},
			Country:      []string{"US"},
			Province:     []string{"CA"},
			Locality:     []string{"San Francisco"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(1, 0, 0), // Valid for 1 year
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	// Add IP addresses and hostnames
	for _, host := range hosts {
		if ip := net.ParseIP(host); ip != nil {
			template.IPAddresses = append(template.IPAddresses, ip)
		} else {
			template.DNSNames = append(template.DNSNames, host)
		}
	}

	// Create certificate
	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return fmt.Errorf("failed to create certificate: %v", err)
	}

	// Write certificate to file
	certOut, err := os.Create(fmt.Sprintf("%s/cert.pem", cm.certDir))
	if err != nil {
		return fmt.Errorf("failed to create cert.pem: %v", err)
	}
	defer certOut.Close()

	if err := pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes}); err != nil {
		return fmt.Errorf("failed to encode certificate: %v", err)
	}

	// Write private key to file
	keyOut, err := os.Create(fmt.Sprintf("%s/key.pem", cm.certDir))
	if err != nil {
		return fmt.Errorf("failed to create key.pem: %v", err)
	}
	defer keyOut.Close()

	privBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	if err := pem.Encode(keyOut, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: privBytes}); err != nil {
		return fmt.Errorf("failed to encode private key: %v", err)
	}

	log.Printf("Self-signed certificate generated successfully")
	return nil
}

// LoadTLSCertificate loads TLS certificate from files
func (cm *CertificateManager) LoadTLSCertificate() (*tls.Certificate, error) {
	cert, err := tls.LoadX509KeyPair(
		fmt.Sprintf("%s/cert.pem", cm.certDir),
		fmt.Sprintf("%s/key.pem", cm.certDir),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load certificate: %v", err)
	}
	return &cert, nil
}

// GetTLSConfig returns TLS configuration
func (cm *CertificateManager) GetTLSConfig() (*tls.Config, error) {
	cert, err := cm.LoadTLSCertificate()
	if err != nil {
		return nil, err
	}

	return &tls.Config{
		Certificates: []tls.Certificate{*cert},
		MinVersion:   tls.VersionTLS12,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
		},
	}, nil
}

// ValidateCertificate validates certificate expiration
func (cm *CertificateManager) ValidateCertificate() error {
	cert, err := cm.LoadTLSCertificate()
	if err != nil {
		return err
	}

	// Parse certificate
	x509Cert, err := x509.ParseCertificate(cert.Certificate[0])
	if err != nil {
		return fmt.Errorf("failed to parse certificate: %v", err)
	}

	// Check expiration
	if time.Now().After(x509Cert.NotAfter) {
		return fmt.Errorf("certificate has expired")
	}

	// Check if expiring soon (within 30 days)
	if time.Now().AddDate(0, 0, 30).After(x509Cert.NotAfter) {
		log.Printf("Warning: Certificate expires on %v", x509Cert.NotAfter)
	}

	return nil
}

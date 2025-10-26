package tls

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"os"
)

func NewTLSConfigServer(isEnabled bool, caCertFilepath, certFilepath, keyFilepath string) (*tls.Config, error) {
	if !isEnabled {
		return nil, nil //nolint:nilnil
	}

	cert, certPool, err := getTLSCertAndCertPool(caCertFilepath, certFilepath, keyFilepath)
	if err != nil {
		return nil, fmt.Errorf("failed to load tls-certificates: %w", err)
	}

	return &tls.Config{
		ClientCAs:    certPool,
		Certificates: []tls.Certificate{cert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		MinVersion:   tls.VersionTLS12,
	}, nil
}

func NewTLSConfigClient(isEnabled bool, caCertFilepath, certFilepath, keyFilepath string) (*tls.Config, error) {
	if !isEnabled {
		return nil, nil //nolint:nilnil
	}

	cert, certPool, err := getTLSCertAndCertPool(caCertFilepath, certFilepath, keyFilepath)
	if err != nil {
		return nil, fmt.Errorf("failed to load tls-certificates: %w", err)
	}

	return &tls.Config{
		RootCAs:      certPool,
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12,
	}, nil
}

func getTLSCertAndCertPool(caCertFilepath, certFilepath, keyFilepath string) (tls.Certificate, *x509.CertPool, error) {
	defaultTLSCert := tls.Certificate{}

	rootCABytes, err := os.ReadFile(caCertFilepath)
	if err != nil {
		return defaultTLSCert, nil, fmt.Errorf("failed to read ca file path (%s): %w", caCertFilepath, err)
	}

	cert, err := tls.LoadX509KeyPair(certFilepath, keyFilepath)
	if err != nil {
		return defaultTLSCert, nil, fmt.Errorf("failed to load X509KeyPair (%s, %s): %w", certFilepath, keyFilepath, err)
	}

	certPool := x509.NewCertPool()

	if ok := certPool.AppendCertsFromPEM(rootCABytes); !ok {
		return defaultTLSCert, nil, errors.New("failed to add root CA to cert pool") //nolint:err113
	}

	return cert, certPool, nil
}

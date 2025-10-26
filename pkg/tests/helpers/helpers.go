package helpers

import (
	"bytes"
	"context"
	cryptorand "crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"log/slog"
	"math/big"
	"net"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
)

func GenerateRSA2048() (*rsa.PrivateKey, error) {
	private, err := rsa.GenerateKey(cryptorand.Reader, 2048)
	if err != nil {
		return nil, err
	}

	return private, nil
}

func DockerPostgres(
	ctx context.Context,
	containerName,
	dbHost,
	dbPort,
	dbName,
	dbUser,
	dbPass string,
) (func() error, error) {
	pool, err := dockertest.NewPool("")
	if err != nil {
		return nil, fmt.Errorf("failed to connecting to docker: %w", err)
	}

	if err = pool.Client.Ping(); err != nil {
		return nil, fmt.Errorf("failed to pinging docker: %w", err)
	}
	if err = pool.RemoveContainerByName(containerName); err != nil {
		return nil, fmt.Errorf("failed to removing container: %w", err)
	}

	var (
		hcOpts  []func(*docker.HostConfig)
		runOpts = &dockertest.RunOptions{
			Repository: "postgres",
			Tag:        "17.5-alpine3.22", // пусть эта пока версия стоит, а то локально другая не поднимается
			Name:       containerName,
			Hostname:   dbHost,
			Env: []string{
				"POSTGRES_DB=" + dbName,
				"POSTGRES_USER=" + dbUser,
				"POSTGRES_PASSWORD=" + dbPass,
			},
			PortBindings: map[docker.Port][]docker.PortBinding{
				"5432/tcp": {{HostPort: dbPort}}, // 5433 -> 5432; выставим свой порт для удобства тестирования
			},
		}
	)

	resource, err := pool.RunWithOptions(runOpts, hcOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed to creating resource: %w", err)
	}

	fnRetry := func() error {
		slog.InfoContext(ctx, "ping db...")

		dsn := fmt.Sprintf(
			"postgres://%s:%s@%s?dbname=%s&sslmode=disable",
			dbUser,
			dbPass,
			net.JoinHostPort(dbHost, resource.GetPort("5432/tcp")),
			dbName,
		)

		poolLoc, err := pgxpool.New(ctx, dsn)
		if err != nil {
			return fmt.Errorf("failed to open db-conn: %w", err)
		}

		return poolLoc.Ping(ctx)
	}

	if err = pool.Retry(fnRetry); err != nil {
		return nil, fmt.Errorf("failed to connecting to docker: %w", err)
	}

	return func() error {
		return pool.Purge(resource)
	}, nil
}

// MakeCA creates caCert (*x509.Certificate), privateKey (*rsa.PrivateKey) and caCertBytes (asn1 bytes)
func MakeCA(
	subject pkix.Name,
	notBefore,
	notAfter time.Time,
) (*x509.Certificate, *rsa.PrivateKey, []byte, error) {
	caCert := &x509.Certificate{
		SerialNumber:          big.NewInt(2019),
		Subject:               subject,
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}

	// generate a private key for the CA
	caPrivKey, err := rsa.GenerateKey(cryptorand.Reader, 2048)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to generate CA private key: %w", err)
	}

	// create the CA certificate bytes
	caCertBytes, err := x509.CreateCertificate(cryptorand.Reader, caCert, caCert, &caPrivKey.PublicKey, caPrivKey)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to create CA certificate: %w", err)
	}

	return caCert, caPrivKey, caCertBytes, nil
}

// MakeCert creates a certificate (crtPEM, keyPEM)
func MakeCert(
	caCert *x509.Certificate,
	caKey *rsa.PrivateKey,
	subject pkix.Name,
	notBefore,
	notAfter time.Time,
) ([]byte, []byte, error) {
	cert := &x509.Certificate{
		SerialNumber: big.NewInt(1658),
		Subject:      subject,
		IPAddresses:  []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback},
		DNSNames:     []string{"localhost"},
		NotBefore:    notBefore,
		NotAfter:     notAfter,
		SubjectKeyId: []byte{1, 2, 3, 4, 6},
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:     x509.KeyUsageDigitalSignature,
	}

	certKey, err := rsa.GenerateKey(cryptorand.Reader, 2048)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate key: %w", err)
	}

	certBytes, err := x509.CreateCertificate(cryptorand.Reader, cert, caCert, &certKey.PublicKey, caKey)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create certificate: %w", err)
	}

	certPEM := new(bytes.Buffer)
	err = pem.Encode(certPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to encode certificate: %w", err)
	}

	certKeyPEM := new(bytes.Buffer)
	err = pem.Encode(certKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(certKey),
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to encode private key: %w", err)
	}

	return certPEM.Bytes(), certKeyPEM.Bytes(), nil
}

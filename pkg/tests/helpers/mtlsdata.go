package helpers

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"
	"time"
)

// MTLSData тестовые данные, клиентские и серверные сертификаты, а так же CA
type MTLSData struct {
	ServerTLSConfig *tls.Config
	ClientTLSConfig *tls.Config
	CABytes         []byte
	ServerCertBytes []byte
	ServerKeyBytes  []byte
	ClientCertBytes []byte
	ClientKeyBytes  []byte
}

func NewMTLSData() (*MTLSData, error) {
	m := MTLSData{}
	now := time.Now()
	oneYear := now.AddDate(1, 0, 0)
	tenYears := now.AddDate(10, 0, 0)
	subjectCA := pkix.Name{
		Country:       []string{"RU"},
		Organization:  []string{"CA WB"},
		Locality:      []string{"Moscow"},
		Province:      []string{"Moscow"},
		StreetAddress: []string{"Tverskaya st."},
		PostalCode:    []string{"123456"},
		CommonName:    "CA",
	}

	subjectServer := subjectCA
	subjectServer.CommonName = "MyServer"
	subjectServer.Organization = []string{"MyServer WB"}

	subjectClient := subjectCA
	subjectClient.CommonName = "MyClient"
	subjectClient.Organization = []string{"MyClient WB"}

	// make cert for CA
	caCert, caPrivKey, caCertBytes, err := MakeCA(subjectCA, now, tenYears)
	if err != nil {
		return nil, fmt.Errorf("failed to create CA cert: %w", err)
	}

	// make cert for server
	certPEMServerBytes, certKeyPEMServerBytes, err := MakeCert(
		caCert,
		caPrivKey,
		subjectServer,
		now,
		oneYear,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create certs bytes for server: %w", err)
	}

	// make cert for client
	certPEMClientBytes, certKeyPEMClientBytes, err := MakeCert(
		caCert,
		caPrivKey,
		subjectClient,
		now,
		oneYear,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create certs bytes for client: %w", err)
	}

	serverCert, err := tls.X509KeyPair(certPEMServerBytes, certKeyPEMServerBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to create cert for server: %w", err)
	}

	clientCert, err := tls.X509KeyPair(certPEMClientBytes, certKeyPEMClientBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to create cert for client: %w", err)
	}

	caPEM := new(bytes.Buffer)
	err = pem.Encode(caPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: caCertBytes,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to encode CAcert to PEM: %w", err)
	}

	caCertPool := x509.NewCertPool()

	if ok := caCertPool.AppendCertsFromPEM(caPEM.Bytes()); !ok {
		return nil, errors.New("failed to append CA cert to pool") //nolint:err113
	}

	serverTLSConfig := &tls.Config{
		ClientCAs:    caCertPool,
		Certificates: []tls.Certificate{serverCert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		MinVersion:   tls.VersionTLS12,
	}

	clientTLSConfig := &tls.Config{
		RootCAs:      caCertPool,
		Certificates: []tls.Certificate{clientCert},
		MinVersion:   tls.VersionTLS12,
	}

	m.ServerTLSConfig = serverTLSConfig
	m.ClientTLSConfig = clientTLSConfig
	m.CABytes = caPEM.Bytes()
	m.ServerCertBytes = certPEMServerBytes
	m.ServerKeyBytes = certKeyPEMServerBytes
	m.ClientCertBytes = certPEMClientBytes
	m.ClientKeyBytes = certKeyPEMClientBytes

	return &m, nil
}

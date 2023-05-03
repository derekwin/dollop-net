// Package tls provides tls config for dollop.
package tls

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"math/big"
	"net"
	"os"
	"time"
)

/*
Usage:
- Generate a private key for the CA:
openssl genrsa -des3 -out ca.key 4096

-


*/

// CreateServerTLSConfig creates server tls config.
func CreateServerTLSConfig(host string, certPath string, keyPath string, skipVerify bool) (*tls.Config, error) {
	// // ca pool
	// pool, err := getCACertPool(caCertPath)
	// if err != nil {
	// 	return nil, err
	// }

	// server certificate
	tlsCert, err := getCertAndKey(certPath, keyPath)
	if err != nil {
		return nil, err
	}

	if tlsCert == nil {
		tlsCert, err = generateCertificate(host)
		if err != nil {
			return nil, err
		}
	}

	clientAuth := tls.NoClientCert
	if !skipVerify {
		clientAuth = tls.RequireAndVerifyClientCert
	}

	return &tls.Config{
		Certificates: []tls.Certificate{*tlsCert},
		// ClientCAs:    pool,
		ClientAuth: clientAuth,
		NextProtos: []string{"dollop"},
	}, nil
}

// CreateClientTLSConfig creates client tls config.
func CreateClientTLSConfig(certPath string, keyPath string, skipVerify bool) (*tls.Config, error) {
	// ca pool
	// pool, err := getCACertPool(caCertPath)
	// if err != nil {
	// 	return nil, err
	// }

	// client certificate
	tlsCert, err := getCertAndKey(certPath, keyPath)
	if err != nil {
		return nil, err
	}

	certificates := []tls.Certificate{}
	if tlsCert != nil {
		certificates = append(certificates, *tlsCert)
	}

	return &tls.Config{
		InsecureSkipVerify: skipVerify,
		Certificates:       certificates,
		// RootCAs:            pool,
		NextProtos:         []string{"dollop"},
		ClientSessionCache: tls.NewLRUClientSessionCache(0),
	}, nil
}

func getCACertPool(caCertPath string) (*x509.CertPool, error) {
	var err error
	var caCert []byte

	caCert, err = os.ReadFile(caCertPath)
	if err != nil {
		return nil, err
	}

	if len(caCert) == 0 {
		return nil, errors.New("tls: cannot load CA cert")
	}

	pool := x509.NewCertPool()
	if ok := pool.AppendCertsFromPEM(caCert); !ok {
		return nil, errors.New("tls: cannot append CA cert to pool")
	}

	return pool, nil
}

func getCertAndKey(certPath string, keyPath string) (*tls.Certificate, error) {
	var err error
	var cert, key []byte

	// certificate
	cert, err = os.ReadFile(certPath)
	if err != nil {
		return nil, err
	}
	// private key
	key, err = os.ReadFile(keyPath)
	if err != nil {
		return nil, err
	}

	if len(cert) == 0 || len(key) == 0 {
		return nil, errors.New("tls: cannot load tls cert/key")
	}

	tlsCert, err := tls.X509KeyPair(cert, key)
	if err != nil {
		return nil, err
	}

	return &tlsCert, nil
}

func generateCertificate(host ...string) (*tls.Certificate, error) {
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err
	}

	notBefore := time.Now()
	notAfter := notBefore.Add(time.Hour * 24 * 365)

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return nil, err
	}

	template := x509.Certificate{
		SerialNumber:          serialNumber,
		Subject:               pkix.Name{Organization: []string{"dollop"}},
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		IsCA:                  true,
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		DNSNames:              []string{"localhost"},
	}

	for _, h := range host {
		if ip := net.ParseIP(h); ip != nil {
			template.IPAddresses = append(template.IPAddresses, ip)
		} else {
			template.DNSNames = append(template.DNSNames, h)
		}
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return nil, err
	}

	// create public key
	certOut := bytes.NewBuffer(nil)
	err = pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	if err != nil {
		return nil, err
	}

	// create private key
	keyOut := bytes.NewBuffer(nil)
	b, err := x509.MarshalECPrivateKey(priv)
	if err != nil {
		return nil, err
	}
	err = pem.Encode(keyOut, &pem.Block{Type: "EC PRIVATE KEY", Bytes: b})
	if err != nil {
		return nil, err
	}

	cert, err := tls.X509KeyPair(certOut.Bytes(), keyOut.Bytes())
	return &cert, err
}

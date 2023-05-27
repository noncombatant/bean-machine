// Adapted from https://golang.org/src/crypto/tls/generate_cert.go?m=text,
// which has the following copyright notice:
//
// Copyright 2009 The Go Authors. All rights reserved. Use of this source code
// is governed by a BSD-style license that can be found in the LICENSE file.

package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"time"
)

func PEMBlockForKey(key *ecdsa.PrivateKey) (*pem.Block, error) {
	bytes, e := x509.MarshalECPrivateKey(key)
	if e != nil {
		return nil, e
	}
	return &pem.Block{Type: "EC PRIVATE KEY", Bytes: bytes}, nil
}

func PEMBlockForCertificate(der []byte) *pem.Block {
	return &pem.Block{Type: "CERTIFICATE", Bytes: der}
}

// Generates a self-signed end-entity X.509 certificate valid for the given
// `hosts` (DNS names and/or IP addresses), the given organizationalUnit `ou`,
// for the validity period `notBefore` through `notAfter`.
//
// Returns a new ECDSA key and a DER-encoded X.509 certificate, or an error.
func GenerateCertificate(hosts []string, ou string, notBefore, notAfter time.Time) (*ecdsa.PrivateKey, []byte, error) {
	key, e := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if e != nil {
		return nil, nil, e
	}

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, e := rand.Int(rand.Reader, serialNumberLimit)
	if e != nil {
		return nil, nil, e
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{ou},
		},
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	for _, h := range hosts {
		if ip := net.ParseIP(h); ip != nil {
			template.IPAddresses = append(template.IPAddresses, ip)
		} else {
			template.DNSNames = append(template.DNSNames, h)
		}
	}

	der, e := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if e != nil {
		return nil, nil, e
	}
	return key, der, nil
}

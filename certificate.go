// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Adapted from https://golang.org/src/crypto/tls/generate_cert.go?m=text

package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"io"
	"log"
	"math/big"
	"net"
	"time"
)

func pemBlockForKey(priv *ecdsa.PrivateKey) *pem.Block {
	b, e := x509.MarshalECPrivateKey(priv)
	if e != nil {
		log.Fatal(e)
	}
	return &pem.Block{Type: "EC PRIVATE KEY", Bytes: b}
}

func generateCertificate(hosts []string, isCA bool, key, certificate io.Writer) {
	priv, e := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if e != nil {
		log.Fatal(e)
	}

	notBefore := time.Now()
	notAfter := notBefore.Add(365 * 24 * time.Hour)

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, e := rand.Int(rand.Reader, serialNumberLimit)
	if e != nil {
		log.Fatal(e)
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"Bean Machine Server"},
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

	if isCA {
		template.IsCA = true
		template.KeyUsage |= x509.KeyUsageCertSign
	}

	derBytes, e := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if e != nil {
		log.Fatal(e)
	}

	e = pem.Encode(certificate, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	if e != nil {
		log.Fatal(e)
	}
	e = pem.Encode(key, pemBlockForKey(priv))
	if e != nil {
		log.Fatal(e)
	}
}

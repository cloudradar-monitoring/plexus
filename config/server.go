package config

import (
	"crypto/rsa"
	"crypto/sha512"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
	"strings"
)

type Server struct {
	Certificate        *x509.Certificate
	CertificateHash    [sha512.Size384]byte
	Key                *rsa.PrivateKey
	PublicKeyPKCS1Hash [sha512.Size384]byte
	ServerID           string
}

func NewServer(crt, key string) (*Server, error) {
	server := &Server{}

	certBytes, err := ioutil.ReadFile(crt)
	if err != nil {
		return nil, fmt.Errorf("cannot read certificate %s: %s", crt, err)
	}
	certBlock, _ := pem.Decode(certBytes)
	cert, err := x509.ParseCertificate(certBlock.Bytes)
	if err != nil {
		return nil, fmt.Errorf("cannot parse certificate %s: %s", crt, err)
	}
	server.Certificate = cert
	server.CertificateHash = sha512.Sum384(cert.Raw)

	keyBytes, err := ioutil.ReadFile(key)
	if err != nil {
		return nil, fmt.Errorf("cannot read private key %s: %s", key, err)
	}
	keyBlock, _ := pem.Decode(keyBytes)
	privKey, err := parsePrivateRsaKey(keyBlock.Bytes)
	if err != nil {
		return nil, fmt.Errorf("cannot parse private key %s: %s", key, err)
	}
	publicKeyAsn1Der := x509.MarshalPKCS1PublicKey(&privKey.PublicKey)
	publicKeyHash := sha512.Sum384(publicKeyAsn1Der)
	server.ServerID = strings.ToUpper(hex.EncodeToString(publicKeyHash[:]))

	return server, nil
}

func parsePrivateRsaKey(block []byte) (*rsa.PrivateKey, error) {
	untyped, err := x509.ParsePKCS8PrivateKey(block)
	if err == nil {
		var ok bool
		key, ok := untyped.(*rsa.PrivateKey)
		if !ok {
			return nil, errors.New("private key must be a rsa private key")
		}
		return key, nil
	}
	return x509.ParsePKCS1PrivateKey(block)
}

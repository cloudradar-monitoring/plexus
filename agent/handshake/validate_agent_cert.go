package handshake

import (
	"bytes"
	"crypto"
	"crypto/rsa"
	"crypto/sha512"
	"crypto/x509"
	"encoding/binary"
	"fmt"
	"net"
)

func validateAgentCert(handshake *Handshake, payload *bytes.Buffer, conn net.Conn) error {
	var certLen int16
	if err := binary.Read(payload, binary.BigEndian, &certLen); err != nil {
		return fmt.Errorf("could not read certlen: %s", err)
	}
	certificate := make([]byte, certLen)
	if _, err := payload.Read(certificate); err != nil {
		return fmt.Errorf("could not read certificate: %s", err)
	}
	cert, err := x509.ParseCertificate(certificate)
	if err != nil {
		return fmt.Errorf("could not parse certificate: %s", err)
	}
	if !handshake.AgentNonceSet {
		return fmt.Errorf("agent nonce not set")
	}
	var res bytes.Buffer
	_, _ = res.Write(handshake.Server.CertificateHash[:])
	_, _ = res.Write(handshake.ServerNonce)
	_, _ = res.Write(handshake.AgentNonce)

	resHash := sha512.Sum384(res.Bytes())

	err = rsa.VerifyPKCS1v15(cert.PublicKey.(*rsa.PublicKey), crypto.SHA384, resHash[:], payload.Bytes())
	if err != nil {
		return fmt.Errorf("could not verify signature: %s", err)
	}
	handshake.Authenticated = true
	if handshake.HasAgentInfo {
		handshake.OnSuccess(&handshake.AgentInfo)
	}
	return nil
}

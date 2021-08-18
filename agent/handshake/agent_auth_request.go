package handshake

import (
	"bytes"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha512"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"net"

	"github.com/gobwas/ws/wsutil"
)

func agentAuthRequest(handshake *Handshake, payload *bytes.Buffer, conn net.Conn) error {
	if payload.Len() != 96 {
		return fmt.Errorf("malformed authentication request, must have length 96 but has %d", payload.Len())
	}

	agentSha386CertHash := make([]byte, 48)
	if _, err := payload.Read(agentSha386CertHash); err != nil {
		return fmt.Errorf("could not read agent cert hash: %s", err)
	}
	if !bytes.Equal(agentSha386CertHash, handshake.Server.CertificateHash[:]) {
		agentCertHex := hex.EncodeToString(agentSha386CertHash)
		serverCertHex := hex.EncodeToString(handshake.Server.CertificateHash[:])
		return fmt.Errorf("tls certificate hash mismatch: agent(%s) server(%s)", agentCertHex, serverCertHex)
	}
	res := &bytes.Buffer{}
	_ = binary.Write(res, binary.BigEndian, int16(1))
	_, _ = res.Write(handshake.Server.CertificateHash[:])
	_, _ = res.Write(handshake.ServerNonce)

	if err := wsutil.WriteServerBinary(conn, res.Bytes()); err != nil {
		return fmt.Errorf("could not send server nonce to agent: %s", err)
	}
	if _, err := payload.Read(handshake.AgentNonce); err != nil {
		return fmt.Errorf("colud not read agent nonce: %s", err)
	}
	handshake.AgentNonceSet = true

	res.Reset()
	_, _ = res.Write(agentSha386CertHash)
	_, _ = res.Write(handshake.AgentNonce)
	_, _ = res.Write(handshake.ServerNonce)

	resHash := sha512.Sum384(res.Bytes())
	res.Reset()
	signature, err := rsa.SignPKCS1v15(rand.Reader, handshake.Server.Key, crypto.SHA384, resHash[:])
	if err != nil {
		return fmt.Errorf("could not sign agent server nonce: %s", err)
	}

	res.Reset()
	_ = binary.Write(res, binary.BigEndian, int16(2))
	_ = binary.Write(res, binary.BigEndian, int16(len(handshake.Server.Certificate.Raw)))
	_, _ = res.Write(handshake.Server.Certificate.Raw)
	_, _ = res.Write(signature)

	if err := wsutil.WriteServerBinary(conn, res.Bytes()); err != nil {
		return fmt.Errorf("could not send signed agent / server nonce: %s", err)
	}
	return nil
}

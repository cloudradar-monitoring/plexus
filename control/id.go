package control

import (
	"encoding/base64"
	"encoding/hex"
	"errors"
	"strings"
)

func base64MeshIDToHexMeshID(id string) (string, error) {
	parts := strings.Split(id, "//")
	if len(parts) != 2 {
		return "", errors.New("invalid id")
	}
	id, err := base64IDToHex(parts[1])
	return "0x" + id, err
}

func base64IDToHex(id string) (string, error) {
	result := strings.ReplaceAll(id, "@", "+")
	result = strings.ReplaceAll(result, "$", "/")
	bin, err := base64.StdEncoding.DecodeString(result)
	if err != nil {
		return "", errors.New("invalid id")
	}
	return strings.ToUpper(hex.EncodeToString(bin)), nil
}

package verify

import (
	"crypto/sha512"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/rs/zerolog"

	"github.com/cloudradar-monitoring/plexus/config"
	"github.com/cloudradar-monitoring/plexus/control"
	"github.com/cloudradar-monitoring/plexus/logger/zerologger"
)

func Verify(configFile string) bool {
	cfg, errs := config.Get(configFile)
	_, _ = zerologger.Init(zerolog.Disabled, "")

	fmt.Print("Config: ")
	for _, err := range errs {
		if err.Level == zerolog.FatalLevel || err.Level == zerolog.PanicLevel {
			fmt.Println("Error")
			fmt.Println(">", err.Msg)
			return false
		}
	}
	fmt.Println("Ok!")

	fmt.Print("MeshCentral Server: ")
	mc, err := control.Connect(cfg.AsControlConfig(), zerologger.Get())
	if err != nil {
		fmt.Println("Error")
		fmt.Println(">", err)
		fmt.Println()
		meshcentralErr(&cfg)
		return false
	}
	defer mc.Close()

	plexusCertHash, err := parseCertificate(cfg.TLSCertFile)
	if err != nil {
		fmt.Println("Error")
		fmt.Println(">", err)
		return false
	}

	serverInfo, err := mc.ServerInfo()
	if err != nil {
		fmt.Println("Error")
		if closeReason := mc.CloseReason(); closeReason != nil {
			fmt.Println(">", closeReason)
		}
		fmt.Println(">", err)
		fmt.Println()
		meshcentralErr(&cfg)
		return false
	}
	fmt.Println("Ok!")
	fmt.Print("TLS Certificate: ")
	if serverInfo.TLSHash != plexusCertHash {
		fmt.Println("Error")
		fmt.Println("> TLS Cert Mismatch")
		fmt.Println("> Plexus     :", plexusCertHash)
		fmt.Println("> MeshCentral:", serverInfo.TLSHash)
		fmt.Println()
		fmt.Println("Plexus and MeshCentral must use the same TLS certificate.")
		return false
	}
	fmt.Println("Ok!")
	return true
}

func meshcentralErr(cfg *config.Config) {
	fmt.Println("Please verify the following things:")
	fmt.Println("* MeshCentral is running and listening on the url defined in PLEXUS_MESH_CENTRAL_URL: ", cfg.MeshCentralURL)
	fmt.Println("* A MeshCentral account with the username in PLEXUS_MESH_CENTRAL_USER", cfg.MeshCentralUser, "exists")
	fmt.Println("* The password in PLEXUS_MESH_CENTRAL_PASS is correct")
	fmt.Println()
	fmt.Println("If everything looks right, please create an new issue: https://github.com/cloudradar-monitoring/plexus/issues/new")
}

func parseCertificate(crt string) (string, error) {
	certBytes, err := ioutil.ReadFile(crt)
	if err != nil {
		return "", fmt.Errorf("cannot read certificate %s: %s", crt, err)
	}
	certBlock, _ := pem.Decode(certBytes)
	cert, err := x509.ParseCertificate(certBlock.Bytes)
	if err != nil {
		return "", fmt.Errorf("cannot parse certificate %s: %s", crt, err)
	}
	publicKeyHash := sha512.Sum384(cert.Raw)
	return strings.ToUpper(hex.EncodeToString(publicKeyHash[:])), nil
}

package restclient

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"io/ioutil"
	"math/big"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestNewClientWithOpts(t *testing.T) {
	r := require.New(t)
	httpClient := &http.Client{}
	client, err := NewClientWithOpts(
		WithHTTPClient(httpClient),
		WithTimeOut(time.Second*10),
		WithBaseURL("http://localhost:8080"),
		WithUsername("testuser"),
		WithPassword("testpasswod"),
	)
	r.Nil(err, "failed to create GD2 client")
	r.Equal(client.baseURL, "http://localhost:8080")
	r.Equal(client.password, "testpasswod")
	r.Equal(client.username, "testuser")
	r.Equal(time.Second*10, client.httpClient.Timeout)
	r.Nil(client.LastErrorResponse())
}

func TestNewClientWithOpts_Error(t *testing.T) {
	r := require.New(t)
	httpClient := &http.Client{}
	client, err := NewClientWithOpts(
		WithHTTPClient(httpClient),
		WithTLSConfig(&TLSOptions{InsecureSkipVerify: false, CaCertFile: "wrongFile.pem"}),
	)
	r.NotNil(err)
	r.NotEqual(client.username, "testuser")
}

func TestWithTLSConfig(t *testing.T) {
	r := require.New(t)

	certOut, err := ioutil.TempFile("", "cert")
	r.Nil(err, "failed to create temp file")
	defer os.Remove(certOut.Name())

	cert, err := generateCert()
	r.Nil(err, "failed to generate dummy certificate")

	err = pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: cert})
	r.Nil(err)

	tlsOpts := &TLSOptions{
		CaCertFile: certOut.Name(),
	}

	client, err := NewClientWithOpts(WithTLSConfig(tlsOpts))
	r.Nil(err)

	transport, ok := client.httpClient.Transport.(*http.Transport)
	r.True(ok)
	r.NotNil(transport.TLSClientConfig)
}

// generateCert will generate a dummy self-signed X.509 certificate.
func generateCert() ([]byte, error) {
	ca := &x509.Certificate{
		SerialNumber: big.NewInt(5687),
		Subject: pkix.Name{
			Organization: []string{"ORG"},
			Country:      []string{"India"},
			Locality:     []string{"City"},
			PostalCode:   []string{"568456"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0),
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}
	priv, _ := rsa.GenerateKey(rand.Reader, 2048)
	pub := &priv.PublicKey
	return x509.CreateCertificate(rand.Reader, ca, ca, pub, priv)
}

package restclient

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

func TestConfig_NewConfig(t *testing.T) {
	cfg := NewConfig()
	assert.IsType(t, &Config{}, cfg, "Object is not a config type")
}

func TestConfig_WithEndPoint(t *testing.T) {
	var c Config
	ep := "http://endpoint"
	a := c.WithEndPoint(ep)
	assert.Equal(t, ep, *a.EndPoint, "Endpoint not as expected")
}

func TestConfig_WithPassword(t *testing.T) {
	var c Config
	p := "test"
	a := c.WithPassword(p)
	assert.Equal(t, p, *a.Password, "Password not as expected")
}

func TestConfig_WithUserId(t *testing.T) {
	var c Config
	u := "user"
	a := c.WithUserId(u)
	assert.Equal(t, u, *a.UserId, "User not as expected")
}

func TestConfig_WithHTTPClient(t *testing.T) {
	tp := time.Second * 123
	hc := http.Client{Timeout: tp}
	var c Config
	a := c.WithHTTPClient(hc)
	assert.Equal(t, tp, a.HTTPClient.Timeout)
}

func TestConfig_WithCACert(t *testing.T) {
	s := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello")
	}))
	defer s.Close()

	//Get certifcate from test TLS server, output in PEM format to file
	certBytes := s.TLS.Certificates[0].Certificate[0]
	cert, _ := x509.ParseCertificate(certBytes)
	//Have to add test cert into a certPool to compare in the assertion as this is all we can get back from the TLSClientConfig of the http.Client and certPool has no public mechanism to extract certs from it
	certPool := x509.NewCertPool()
	certPool.AddCert(cert)

	var c Config
	a := c.WithCACert(cert)
	transport := a.HTTPClient.Transport
	assert.Equal(t, certPool, transport.(*http.Transport).TLSClientConfig.RootCAs, "Certificate not set to be trusted in HTTP Client")
}

func TestConfig_WithCAFilePath(t *testing.T) {
	s := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello")
	}))
	defer s.Close()

	//Get certifcate from test TLS server, output in PEM format to file
	certBytes := s.TLS.Certificates[0].Certificate[0]
	cert, _ := x509.ParseCertificate(certBytes)
	//Have to add test cert into a certPool to compare in the assertion as this is all we can get back from the TLSClientConfig of the http.Client and certPool has no public mechanism to extract certs from it
	certPool := x509.NewCertPool()
	certPool.AddCert(cert)

	//Get certifcate from test TLS server, output in PEM format to file
	certOut, _ := ioutil.TempFile(os.TempDir(), "prefix")
	defer os.Remove(certOut.Name())
	pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: certBytes})

	var c Config
	a := c.WithCAFilePath(certOut.Name())
	transport := a.HTTPClient.Transport
	assert.Equal(t, certPool, transport.(*http.Transport).TLSClientConfig.RootCAs, "Certificate not set to be trusted in HTTP Client")
}

func TestConfig_Load(t *testing.T) {
	completeCfg := `{
		"UserId": "testuser",
		"Password": "password",
		"EndPoint": "http://testurl"
	}`
	noUserIdCfg := `{
		"Password": "password",
		"EndPoint": "http://testurl"
	}`
	noRoleIdCfg := `{
		"UserId": "testuser",
		"Password": "password",
		"EndPoint": "http://testurl"
	}`
	var tests = []struct {
		cfgJson       *string
		userIdDefined bool
		roleIdDefined bool
	}{
		{&completeCfg, true, true},
		{&noUserIdCfg, false, true},
		{&noRoleIdCfg, true, false},
	}
	for _, test := range tests {
		testConfigFile, _ := ioutil.TempFile(os.TempDir(), "config")
		defer os.Remove(testConfigFile.Name())
		testConfigFile.WriteString(*test.cfgJson)

		cfg := Load(testConfigFile.Name())
		assert.IsType(t, &Config{}, cfg, "Object is not a config type")
		if test.userIdDefined {
			assert.Equal(t, "testuser", *cfg.UserId, "UserId not set on config correctly")
		} else {
			assert.Nil(t, cfg.UserId, "UserId pointer should be nil")
		}
	}
}

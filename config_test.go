package restclient

import (
	"crypto/x509"
	"encoding/json"
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
	invalidEp := "ftp://endpoint"

	a := c.WithEndPoint(ep)
	assert.Equal(t, ep, *a.EndPoint, "Endpoint not as expected")
	assert.Nil(t, a.configErr, "Configuration error is not nil when providing a endpoint")

	a = c.WithEndPoint(invalidEp)
	assert.NotNil(t, a.configErr, "An invalid endpoint did not create an error in the configuration")
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
	assert.Nil(t, a.configErr, "Configuration error is not nil when providing a valid certificate")
	transport := a.HTTPClient.Transport
	assert.Equal(t, certPool, transport.(*http.Transport).TLSClientConfig.RootCAs, "Certificate not set to be trusted in HTTP Client")
}

func TestConfig_WithCAFilePath(t *testing.T) {
	s := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello")
	}))
	defer s.Close()

	//Get certificate from test TLS server, output in PEM format to file
	certBytes := s.TLS.Certificates[0].Certificate[0]
	cert, _ := x509.ParseCertificate(certBytes)
	//Have to add test cert into a certPool to compare in the assertion as this is all we can get back from the TLSClientConfig of the http.Client and certPool has no public mechanism to extract certs from it
	certPool := x509.NewCertPool()
	certPool.AddCert(cert)

	//Get certificate from test TLS server, output in PEM format to file
	certOut, _ := ioutil.TempFile(os.TempDir(), "prefix")
	defer os.Remove(certOut.Name())
	pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: certBytes})

	var c Config
	a := c.WithCAFilePath(certOut.Name())
	assert.Nil(t, a.configErr, "Configuration error is not nil when providing a valid certificate file")
	transport := a.HTTPClient.Transport
	assert.Equal(t, certPool, transport.(*http.Transport).TLSClientConfig.RootCAs, "Certificate not set to be trusted in HTTP Client")

	invalidPEM, _ := ioutil.TempFile(os.TempDir(), "validcert")
	defer os.Remove(invalidPEM.Name())
	invalidPEM.Write([]byte("This is not valid PEM data"))
	a = c.WithCAFilePath(invalidPEM.Name())
	assert.NotNil(t, a.configErr, "An invalid CA file did not create an error in the configuration")
}

func TestConfig_Validate(t *testing.T) {
	s := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello")
	}))
	defer s.Close()

	//Get certificate from test TLS server, output in PEM format to file
	certBytes := s.TLS.Certificates[0].Certificate[0]
	cert, _ := x509.ParseCertificate(certBytes)
	//Have to add test cert into a certPool to compare in the assertion as this is all we can get back from the TLSClientConfig of the http.Client and certPool has no public mechanism to extract certs from it
	certPool := x509.NewCertPool()
	certPool.AddCert(cert)

	//Get certificate from test TLS server, output in PEM format to file
	certOut, _ := ioutil.TempFile(os.TempDir(), "prefix")
	defer os.Remove(certOut.Name())
	pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: certBytes})

	validSimple := `{
		"EndPoint": "http://testurl"
	}`
	validUserDetails := `{
		"UserId": "testuser",
		"Password": "password",
		"EndPoint": "http://testurl"
	}`
	validTLS := fmt.Sprintf(`{
		"UserId": "testuser",
		"Password": "password",
		"EndPoint": "https://testurl",
		"TrustCACert": "%s"
	}`, certOut.Name())
	invalidSimple := `{
		"EndPoint": "ftp://testurl"
	}`
	invalidNoEndpoint := `{
		"UserId": "testuser",
		"Password": "password"
	}`
	invalidUserDetails := `{
		"UserId": "testuser",
		"EndPoint": "http://testurl"
	}`
	invalidTLS := `{
		"UserId": "testuser",
		"Password": "password",
		"EndPoint": "https://testurl"
	}`
	var tests = []struct {
		cfgJson *string
		valid   bool
	}{
		{&validSimple, true},
		{&validUserDetails, true},
		{&validTLS, true},
		{&invalidNoEndpoint, false},
		{&invalidSimple, false},
		{&invalidUserDetails, false},
		{&invalidTLS, false},
	}
	for _, test := range tests {
		var c Config
		e := json.Unmarshal([]byte(*test.cfgJson), &c)
		if e != nil {
			t.Fatalf("Tests are broken, could not parse test data: %v\nerror: %v", *test.cfgJson, e)
		}
		err := c.Validate()
		if test.valid {
			assert.Nil(t, err, "Configuration was valid but Validate method returned an error")
		} else {
			assert.NotNil(t, err, "Configuration was not valid but Validation method did not return an error")
		}
	}
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

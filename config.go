package restclient

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"io/ioutil"
	"net/http"
)

type Config struct {
	UserId      *string
	Password    *string
	EndPoint    *string
	TrustCACert *string
	HTTPClient  *http.Client
}

func NewConfig() *Config {
	return &Config{
		HTTPClient: http.DefaultClient,
	}
}

func (c *Config) WithUserId(u string) *Config {
	c.UserId = &u
	return c
}

func (c *Config) WithPassword(p string) *Config {
	c.Password = &p
	return c
}

func (c *Config) WithEndPoint(e string) *Config {
	c.EndPoint = &e
	return c
}

func (c *Config) WithCACert(cert *x509.Certificate) *Config {
	// Set up our own certificate pool
	if len(cert.Raw) == 0 {
		panic("Certifcate provided is empty")
	}
	tlsConfig := &tls.Config{RootCAs: x509.NewCertPool()}
	transport := &http.Transport{TLSClientConfig: tlsConfig}
	if c.HTTPClient == nil {
		c.HTTPClient = http.DefaultClient
	}
	c.HTTPClient.Transport = transport
	tlsConfig.RootCAs.AddCert(cert)
	return c
}

func (c *Config) WithCAFilePath(caFilePath string) *Config {
	// Set up our own certificate pool
	c.TrustCACert = &caFilePath
	tlsConfig := &tls.Config{RootCAs: x509.NewCertPool()}
	transport := &http.Transport{TLSClientConfig: tlsConfig}
	if c.HTTPClient == nil {
		c.HTTPClient = http.DefaultClient
	}
	c.HTTPClient.Transport = transport
	// Load our trusted certificate path
	pemData, err := ioutil.ReadFile(caFilePath)
	if err != nil {
		panic(err)
	}
	ok := tlsConfig.RootCAs.AppendCertsFromPEM(pemData)
	if !ok {
		panic("Couldn't load PEM data")
	}

	return c
}

func (c *Config) WithHTTPClient(client http.Client) *Config {
	c.HTTPClient = &client
	return c
}

func Load(cfgPath string) *Config {
	j, err := ioutil.ReadFile(cfgPath)
	if err != nil {
		panic("Configuration file could not be openned: " + cfgPath + " " + err.Error())
	}

	var c Config
	err = json.Unmarshal(j, &c)
	if err != nil {
		panic("Configuration file could not be parsed: " + err.Error())
	}
	c.HTTPClient = http.DefaultClient
	if c.TrustCACert != nil {
		c.WithCAFilePath(*c.TrustCACert)
	}
	return &c
}

package restclient

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strings"
)

// A Config specifies the details needed to connect to a ReST service
type Config struct {
	UserId      *string
	Password    *string
	EndPoint    *string
	TrustCACert *string
	HTTPClient  *http.Client
}

// Create new, blank ReST client config
func NewConfig() *Config {
	return &Config{
		HTTPClient: http.DefaultClient,
	}
}

// Add a user ID to the config for basic authentication to the ReST service
func (c *Config) WithUserId(u string) *Config {
	c.UserId = &u
	return c
}

// Add a password to the config for basic authentication to the ReST service
func (c *Config) WithPassword(p string) *Config {
	c.Password = &p
	return c
}

//Specify the URL endpoint of the ReST service in the form http(s)://hostname:port
func (c *Config) WithEndPoint(e string) *Config {
	if strings.HasPrefix(e, "http://") || strings.HasPrefix(e, "https://") {
		c.EndPoint = &e
	}
	return c
}

// Add a trusted x509 certificate to the configuration.
// If the ReST service implements TLS/SSL then certificates signed by this CA certificate will be trusted.
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

// Add a trusted x509 certificate to the configuration by specifying a path to a PEM format certificate file.
// If the ReST service implements TLS/SSL then certificates signed by this CA certificate will be trusted.
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

//Override with a specific http.Client to be used for the connection to the ReST service.
func (c *Config) WithHTTPClient(client http.Client) *Config {
	c.HTTPClient = &client
	return c
}

func (c *Config) Validate() error {
	var msg string
	if c.EndPoint == nil {
		msg += "Endpoint not defined; "
	}
	if !strings.HasPrefix(*c.EndPoint, "http://") && !strings.HasPrefix(*c.EndPoint, "https://") {
		msg += "Endpoint is neither http:// nor https://"
	}
	if c.UserId != nil && c.Password == nil {
		msg += "UserId defined by no password set; "
	}
	if strings.HasPrefix(*c.EndPoint, "https://") && c.TrustCACert == nil {
		msg += "HTTPS endpoint defined but no trust certificate set; "
	}
	if msg != "" {
		return errors.New("ReST Client Config validation failed: " + msg)
	}
	return nil
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

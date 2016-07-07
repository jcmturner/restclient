package restclient

import (
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

const (
	testUserId   = "userA"
	testPassword = "pa55word"
)

func testServer(response string, tls bool, basicAuth bool) *httptest.Server {
	if tls {
		s := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !basicAuth || checkAuth(r) {
				testHandler(w, r, response)
			} else {
				w.WriteHeader(http.StatusUnauthorized)
			}
		}))
		return s
	} else {
		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !basicAuth || checkAuth(r) {
				testHandler(w, r, response)
			} else {
				w.WriteHeader(http.StatusUnauthorized)
			}
		}))
		return s
	}
}

func testHandler(w http.ResponseWriter, r *http.Request, response string) {
	w.Header().Set("Content-Type", "application/json")
	defer r.Body.Close()
	body, _ := ioutil.ReadAll(r.Body)
	w.Header().Set("X-Postvalue", string(body))
	w.Header().Set("X-Queryvalue", r.URL.Query().Get("queryKey"))
	fmt.Fprintln(w, response)
}

func checkAuth(r *http.Request) bool {
	s := strings.SplitN(r.Header.Get("Authorization"), " ", 2)
	if len(s) != 2 {
		return false
	}

	b, err := base64.StdEncoding.DecodeString(s[1])
	if err != nil {
		return false
	}

	pair := strings.SplitN(string(b), ":", 2)
	if len(pair) != 2 {
		return false
	}
	return pair[0] == testUserId && pair[1] == testPassword
}

func TestBuild(t *testing.T) {
	var tests = []struct {
		userid       string
		passwd       string
		url          string
		path         string
		method       string
		expectedPath string
	}{
		{"", "", "http://test", "/test/path", "GET", "/test/path"},
		{"testuser", "testpasswd", "http://test", "/test/path", "GET", "/test/path"},
		{"", "", "http://test", "", "GET", "/"},
		{"", "", "http://test", "test/path", "GET", "/test/path"},
		{"", "", "http://test/", "/test/path", "GET", "/test/path"},
		{"", "", "http://test", "/test/path", "POST", "/test/path"},
		{"", "", "http://test", "/test/path", "PUT", "/test/path"},
		{"", "", "http://test", "/test/path", "PATCH", "/test/path"},
	}
	for _, test := range tests {
		authHeader := "Basic " + base64.StdEncoding.EncodeToString([]byte(test.userid+":"+test.passwd))
		c := &Config{
			UserId:     &test.userid,
			Password:   &test.passwd,
			EndPoint:   &test.url,
			HTTPClient: http.DefaultClient,
		}
		var o *Operation
		switch test.method {
		case "GET":
			o = NewGetOperation()
		case "POST":
			o = NewPostOperation()
		case "PUT":
			o = NewPutOperation()
		case "PATCH":
			o = NewPatchOperation()
		}
		o.WithPath(test.path)
		r, err := BuildRequest(c, o)
		if err != nil {
			t.Errorf("Error building request: %v", err)
		}
		assert.IsType(t, &Request{}, r, "Object is not a Request type")
		assert.Equal(t, c, r.Config, "Request does not contain the right Config")
		assert.Equal(t, o, r.Operation, "Request does not contain the right Operation")
		assert.IsType(t, &http.Request{}, r.HTTPRequest, "HTTPRequest is not of the correct type")
		assert.Equal(t, test.expectedPath, r.HTTPRequest.URL.Path, "URL not set correctly in HTTPRequest")
		assert.Equal(t, authHeader, r.HTTPRequest.Header.Get("Authorization"), "Authorization header not set as expected")
		assert.Equal(t, test.method, r.HTTPRequest.Method, "Method not set as expected")
	}
}

func TestSend(t *testing.T) {
	//Set up test data
	var mtStr string
	validuser := testUserId
	validpasswd := testPassword
	invaliduser := "invalid" + validuser
	invalidpasswd := "invalid" + validpasswd
	//Post data
	p := `{ "postKey": "postData" }`
	//Query string
	q := "queryKey=queryData"
	//Response data
	const value1 = "value1"
	const value2 = "value2"
	j := fmt.Sprintf(
		`{
			"level1str": "%v",
			"level1bool": %v,
			"level1int": %v,
			"level1strarray": ["root", "hello"],
			"level2": {
				"level2str": "%v",
				"level2str2": "blah"
			}
		}`,
		value1, false, 2, value2)
	//Target struct for response data
	type rdatatype struct {
		Level1Str      string   `json:"level1str"`
		Level1Bool     bool     `json:"level1bool"`
		Level1Int      int      `json:"level1int"`
		Level1Strarray []string `json:"level1strarray"`
		Level2         struct {
			Level2Str  string `json:"level2str"`
			Level2Str2 string `json:"level2str2"`
		} `json:"level2"`
	}
	var rdata rdatatype

	//Define test combinations
	var tests = []struct {
		verb         string
		tls          bool
		basicAuth    bool
		userid       *string
		passwd       *string
		expectedCode int
		qdata        *string
		postdata     *string
	}{
		{"GET", false, false, &validuser, &validpasswd, http.StatusOK, nil, nil},
		{"GET", true, false, &validuser, &validpasswd, http.StatusOK, nil, nil},
		{"GET", false, true, &validuser, &validpasswd, http.StatusOK, nil, nil},
		{"GET", true, true, &validuser, &validpasswd, http.StatusOK, nil, nil},
		{"GET", false, false, nil, nil, http.StatusOK, &q, nil},
		{"POST", false, false, nil, nil, http.StatusOK, nil, &p},
		{"POST", false, false, nil, nil, http.StatusOK, &q, &p},
		{"GET", false, true, &mtStr, &mtStr, http.StatusUnauthorized, nil, nil},
		{"GET", false, true, &invaliduser, &validpasswd, http.StatusUnauthorized, nil, nil},
		{"GET", false, true, &invaliduser, &mtStr, http.StatusUnauthorized, nil, nil},
		{"GET", false, true, &validuser, &invalidpasswd, http.StatusUnauthorized, nil, nil},
		{"GET", false, true, &invaliduser, &invalidpasswd, http.StatusUnauthorized, nil, nil},
		{"GET", false, true, nil, &invalidpasswd, http.StatusUnauthorized, nil, nil},
		{"GET", false, true, &invaliduser, nil, http.StatusUnauthorized, nil, nil},
		{"GET", false, true, nil, nil, http.StatusUnauthorized, nil, nil},
		{"GET", true, true, &mtStr, &mtStr, http.StatusUnauthorized, nil, nil},
		{"GET", true, true, &invaliduser, &validpasswd, http.StatusUnauthorized, nil, nil},
		{"GET", true, true, &invaliduser, &mtStr, http.StatusUnauthorized, nil, nil},
		{"GET", true, true, &validuser, &invalidpasswd, http.StatusUnauthorized, nil, nil},
		{"GET", true, true, &invaliduser, &invalidpasswd, http.StatusUnauthorized, nil, nil},
		{"GET", true, true, nil, &invalidpasswd, http.StatusUnauthorized, nil, nil},
		{"GET", true, true, &invaliduser, nil, http.StatusUnauthorized, nil, nil},
		{"GET", true, true, nil, nil, http.StatusUnauthorized, nil, nil},
	}
	for _, test := range tests {
		s := testServer(j, test.tls, test.basicAuth)

		//Create config
		c := NewConfig().WithEndPoint(s.URL)
		if test.userid != nil {
			c.WithUserId(*test.userid)
		}
		if test.passwd != nil {
			c.WithPassword(*test.passwd)
		}
		if test.tls {
			//Get certifcate from test TLS server, output in PEM format to file
			certOut, _ := ioutil.TempFile(os.TempDir(), "prefix")
			defer os.Remove(certOut.Name())
			certBytes := s.TLS.Certificates[0].Certificate[0]
			pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: certBytes})

			c.WithCAFilePath(certOut.Name())
		}

		//Create operation
		var o *Operation
		switch test.verb {
		case "GET":
			o = NewGetOperation()
		case "POST":
			o = NewPostOperation()
		case "PUT":
			o = NewPutOperation()
		case "PATCH":
			o = NewPatchOperation()
		}
		o.WithPath("/test/path")
		if test.qdata != nil {
			o.WithQueryDataString(*test.qdata)
		}
		if test.postdata != nil {
			o.WithBodyDataString(*test.postdata)
		}

		//Test without a response data target first
		r, err := BuildRequest(c, o)
		if err != nil {
			t.Errorf("Error building request: %v", err)
		}
		code, _ := Send(r)
		assert.Equal(t, test.expectedCode, *code, "Expected to get HTTP 200 status returned from Send")
		if o.responsePtr != nil {
			t.Errorf("Target for the response data should be nil: %v", o.responsePtr)
		}

		//Test with response data target set
		o.WithResponseTarget(&rdata)
		r, err = BuildRequest(c, o)
		if err != nil {
			t.Errorf("Error building request: %v", err)
		}
		code, _ = Send(r)
		assert.Equal(t, test.expectedCode, *code, "Expected to get HTTP 200 status returned from Send")
		if test.expectedCode != http.StatusUnauthorized {
			assert.Equal(t, value1, rdata.Level1Str, "Response data not as expected")
			assert.Equal(t, value2, rdata.Level2.Level2Str, "Response data not as expected %v", rdata)
			if test.qdata != nil {
				assert.Equal(t, "queryData", r.HTTPResponse.Header.Get("X-Queryvalue"), "Server did not get the query data we sent. Should respond in the headers: %v", r.HTTPResponse.Header)
			}
			if test.postdata != nil {
				assert.Equal(t, *test.postdata, r.HTTPResponse.Header.Get("X-Postvalue"), "Server did not get the post data we sent. Should respond in the headers: %v", r.HTTPResponse.Header)
			}
		}

		//Test with server shutdown
		s.Close()
		code, err = Send(r)
		assert.Equal(t, http.StatusServiceUnavailable, *code, "Expected to get HTTP 503 status when server not running")
		assert.NotNil(t, err, "Expect to get an error when server not available")
	}
}

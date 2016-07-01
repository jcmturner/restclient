package restclient

import (
	"encoding/base64"
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func testServer(response string) *httptest.Server {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, response)
	}))
	return s
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
	s := testServer(j)
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
	var mtStr string
	c := &Config{
		UserId:     &mtStr,
		Password:   &mtStr,
		EndPoint:   &s.URL,
		HTTPClient: http.DefaultClient,
	}
	o := NewGetOperation().WithPath("/test/path").WithResponseTarget(&rdata)
	r, err := BuildRequest(c, o)
	if err != nil {
		t.Errorf("Error building request: %v", err)
	}
	code, _ := Send(r)
	assert.Equal(t, 200, *code, "Expected to get HTTP 200 status returned from Send")
	assert.Equal(t, value1, rdata.Level1Str, "Response data not as expected")
	assert.Equal(t, value2, rdata.Level2.Level2Str, "Response data not as expected %v", rdata)

	//Test behaviour when server not running
	s.Close()
	code, err = Send(r)
	assert.Equal(t, 503, *code, "Expected to get HTTP 503 status when server not running")
	assert.NotNil(t, err, "Expect to get an error when server not available")
}

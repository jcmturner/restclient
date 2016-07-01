package restclient

import (
	"github.com/stretchr/testify/assert"
	"net/url"
	"testing"
)

func TestNewGetOperation(t *testing.T) {
	o := NewGetOperation()
	assert.IsType(t, Operation{}, *o, "Did not return an Operation type")
	assert.Equal(t, "GET", o.httpMethod, "HTTP method on operation not correct")
}

func TestNewPatchOperation(t *testing.T) {
	o := NewPatchOperation()
	assert.IsType(t, Operation{}, *o, "Did not return an Operation type")
	assert.Equal(t, "PATCH", o.httpMethod, "HTTP method on operation not correct")
}

func TestNewPostOperation(t *testing.T) {
	o := NewPostOperation()
	assert.IsType(t, Operation{}, *o, "Did not return an Operation type")
	assert.Equal(t, "POST", o.httpMethod, "HTTP method on operation not correct")
}

func TestNewPutOperation(t *testing.T) {
	o := NewPutOperation()
	assert.IsType(t, Operation{}, *o, "Did not return an Operation type")
	assert.Equal(t, "PUT", o.httpMethod, "HTTP method on operation not correct")
}

func TestOperation_WithPath(t *testing.T) {
	o := NewGetOperation()
	o.WithPath("/some/path")
	assert.Equal(t, "/some/path", o.httpPath, "HTTP path not set correctly")

}

func TestOperation_WithResponseTarget(t *testing.T) {
	o := NewGetOperation()
	type test struct {
		key1 string
		key2 int
	}
	var testinst test
	o.WithResponseTarget(&testinst)
	if &testinst != o.responsePtr {
		t.Errorf("Pointer not stored as reponse target when passed pointer")
	}
}

func TestOperation_WithSendDataByteArray(t *testing.T) {
	var d = []byte{0, 1, 2, 3}
	o := NewGetOperation()
	o.WithSendDataByteArray(d)
	assert.Equal(t, d, o.sendData, "Send data not set correctly")
}

func TestOperation_WithSendDataString(t *testing.T) {
	s := "test string"
	o := NewGetOperation()
	o.WithSendDataString(s)
	assert.Equal(t, s, string(o.sendData), "Send data not set correctly")
}

func TestOperation_WithSendDataURLValues(t *testing.T) {
	u := url.Values{}
	u.Set("key1", "value1")
	u.Add("key2", "value2")
	o := NewGetOperation()
	o.WithSendDataURLValues(u)
	assert.Equal(t, u.Encode(), string(o.sendData), "Send data not set correctly")
}

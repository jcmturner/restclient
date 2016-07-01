package restclient

import (
	"net/url"
	"reflect"
)

// An Operation is the service API operation to be made.
// Data can be of one of these types: []byte, string, url.Values
type Operation struct {
	httpMethod  string
	httpPath    string
	sendData    []byte
	responsePtr interface{}
}

func NewGetOperation() (o *Operation) {
	o = &Operation{
		httpMethod: "GET",
	}
	return
}
func NewPostOperation() (o *Operation) {
	o = &Operation{
		httpMethod: "POST",
	}
	return
}
func NewPutOperation() (o *Operation) {
	o = &Operation{
		httpMethod: "PUT",
	}
	return
}
func NewPatchOperation() (o *Operation) {
	o = &Operation{
		httpMethod: "PATCH",
	}
	return
}

func (o *Operation) WithPath(p string) *Operation {
	o.httpPath = p
	return o
}

func (o *Operation) WithSendDataString(d string) *Operation {
	o.sendData = []byte(d)
	return o
}
func (o *Operation) WithSendDataByteArray(d []byte) *Operation {
	o.sendData = d
	return o
}
func (o *Operation) WithSendDataURLValues(d url.Values) *Operation {
	o.sendData = []byte(d.Encode())
	return o
}

func (o *Operation) WithResponseTarget(v interface{}) *Operation {
	//Checking the value is a pointer. Need some better error handling here and this just swallows
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return o
	}
	o.responsePtr = v
	return o
}

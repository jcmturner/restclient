package restclient

import (
	"encoding/json"
	"net/url"
	"reflect"
)

// An Operation is the ReST service API operation to be made.
// It encapsulates the HTTP method to be used, the path to be called and data to be sent, both post and query string data.
// It also contains a point to a struct that will be the target to hold any response data from the ReST service.
type Operation struct {
	httpMethod  string
	httpPath    string
	sendData    []byte
	queryData   string
	responsePtr interface{}
}

// Create an Operation that uses the GET verb against the ReST service.
func NewGetOperation() (o *Operation) {
	o = &Operation{
		httpMethod: "GET",
	}
	return
}

// Create an Operation that uses the POST verb against the ReST service.
func NewPostOperation() (o *Operation) {
	o = &Operation{
		httpMethod: "POST",
	}
	return
}

// Create an Operation that uses the PUT verb against the ReST service.
func NewPutOperation() (o *Operation) {
	o = &Operation{
		httpMethod: "PUT",
	}
	return
}

// Create an Operation that uses the PATCH verb against the ReST service.
func NewPatchOperation() (o *Operation) {
	o = &Operation{
		httpMethod: "PATCH",
	}
	return
}

// Define the path of the ReST service to call.
func (o *Operation) WithPath(p string) *Operation {
	o.httpPath = p
	return o
}

// Add some post data to the Operation by providing a string
func (o *Operation) WithBodyDataString(d string) *Operation {
	o.sendData = []byte(d)
	return o
}

// Add some post data to the Operation by providing a byte array
func (o *Operation) WithBodyDataByteArray(d []byte) *Operation {
	o.sendData = d
	return o
}

// Add some post data to the Operation by providing a url.Values type
func (o *Operation) WithBodyDataURLValues(d url.Values) *Operation {
	o.sendData = []byte(d.Encode())
	return o
}

// Add some post data to the Operation by providing a struct instance that will be marshaled into JSON
func (o *Operation) WithBodyDataStruct(d interface{}) *Operation {
	o.sendData, _ = json.Marshal(d)
	return o
}

// Add data to the query string of the Operation.
// This method is used to define this using a string.
// The string will need to be appropriately URL encoded
func (o *Operation) WithQueryDataString(d string) *Operation {
	//o.queryData = url.QueryEscape(d)
	o.queryData = d
	return o
}

// Add data to the query string of the Operation.
// This method is used to define this using a url.Values type.
func (o *Operation) WithQueryDataURLValues(d url.Values) *Operation {
	o.queryData = d.Encode()
	return o
}

// Define the pointer to a struct that will be used to hold the response data from the ReST call.
// When the request is sent to the ReST service any response will be marshalled into this struct.
func (o *Operation) WithResponseTarget(v interface{}) *Operation {
	//Checking the value is a pointer. Need some better error handling here and this just swallows
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return o
	}
	o.responsePtr = v
	return o
}

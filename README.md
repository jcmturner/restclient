# restclient library

This library can be used in Golang projects to simplify and standardise interacting with ReST services.

## How to use

### Configuration
First define how to access the ReST service. This is done using restclient.Config
Create a new config instance:
```
c := restclient.NewConfig()
```
You need to at least define an endpoint URL for the ReST service:
```
c.WithEndPoint("https://somehost:8080")
```
You can specify user name and password authentication details to this service (currently only basic authentication is supported):
```
c.WithUserId("userA").WithPassword("pa55word")
```
If the endpoint is connected to over TLS you can specify a signing certificate to trust for the connection. This can be done either by providing the path to a PEM format certificate file or a pointer to an x509.Certificate object.
```
c.WithCAFilePath("/path/to/trusted/cert.pem")
c.WithCACert(&x509.Certificate{})
```
A configuration can also be loaded from a file containing JSON formatted data. For example the JSON configuration file could contain:
```
{
  "EndPoint": "https://somehost:8080",
  "UserId": "userA",
  "Password": "pa55word",
  "TrustCACert": "/path/to/trusted/cert.pem"
}
```
It can be loaded with:
```
c.Load("/path/to/config.json")
```

### Operation
For each operation you want to perform against the ReST service create a new instance of restclient.Operation.
First create the Operation instance with the relevant method for the HTTP verb the ReST call requires:
```
o := restclient.NewGetOperation()
o := restclient.NewPostOperation()
o := restclient.NewPutOperation()
o := restclient.NewPatchOperation()
```
Define the path in the service the operation will call
```
o.WithPath("/some/api/path")
```
If a query string needs to be defined one of the following methods can be used. Note that if you are passing a string you need to first url encode it appropriately.
```
o.WithQueryDataString("something=value&somethingelse=value2")
o.WithQueryDataURLValues(url.Values{})
```
If posting data in the call is required it can be provided as either a string, byte array or url.Values with these methods.
```
o.WithSendDataString("somedatatosend")
o.WithSendDataByteArray([]byte{})
o.WithSendDataURLValues(url.Values{})
```
If the call returns data you want to retrieve, define a struct that a JSON response will parse into. Create an instance of this struct and provide the pointer to the Operation instance:
```
type AWSCredentials struct {
	SecretAccessKey string    `json:"SecretAccessKey"`
	SessionToken    string    `json:"SessionToken"`
	Expiration      time.Time `json:"Expiration"`
	AccessKeyID     string    `json:"AccessKeyId"`
}
var d AWSCredentials
o.WithResponseTarget(&d)
```

### Build the Request
With the  operation object and a config object created the next step is to build the request:
```
req, err := restclient.BuildRequest(c, o)
```
Note: It would be usual to have one config object and multiple operation objects and multiple request object built from these.

### Send the Request
Now we have the request we can send it to the service:
```
httpcode, err := restclient.Send(req)
```
Any response will be marshalled into the response target struct that you provided to the operation.

## Example use
To see an example of this library being used see: [https://github.com/jcmturner/aws-cli-wrapper]

## Improvements needed...
- Logging
- Tests for query string
- Go doc comments

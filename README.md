# restclient library

This library can be used in Golang projects to simplify and standardise interacting with ReST services.

## How to use

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
If posting data in the call is relevant this can be provided as a string, byte array or url.Values with these methods. Currently you can only provide post data, ability to define query strings is going to be added.
```
o.WithSendDataString("somedatatosend")
o.WithSendDataByteArray(bytearray)
o.WithSendDataURLValues(urlValuesType)
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

### Configuration
Now we have a defined operation we need to define where to send it. This is done using the restclient.Config
Create a new config instance:
```
c := restclient.NewConfig()
```
You need to at least define an endpoint URL for the ReST service:
```
c.WithEndPoint("https://somehost:8080")
```
You can specify user name and password authenticaiton details to this service (currently only basic authentication is supported):
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

### Build the Request
No we have an operation object and a config object we build the request:
```
req, err := restclient.BuildRequest(c, o)
```
Note: It would be usual to have one config object and multiple operation objects and multiple request object build from these.

### Send the Request
Now we have the request we can send it to the service:
```
httpcode, err := restclient.Send(req)
```
Any response will be marshalled into the response target struct that you provided to the operation.

## Improvements needed...
- Mechanism to define query string

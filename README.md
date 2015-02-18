go-pubcontrol
===============

Author: Konstantin Bokarius <kon@fanout.io>

A Go convenience library for publishing messages using the EPCP protocol.

License
-------

go-pubcontrol is offered under the MIT license. See the LICENSE file.

Installation
------------

```sh
go get github.com/fanout/go-pubcontrol
```

go-pubcontrol requires jwt-go 2.2.0 and lane 0.1.0. To ensure that the correct versions of both dependencies are installed follow these steps:

```sh
go get github.com/tools/godep
cd $GOPATH/src/github.com/fanout/go-pubcontrol
$GOPATH/bin/godep restore
```

Usage
-----

```go
package main

import "github.com/fanout/go-pubcontrol"
import "fmt"
import "encoding/base64"

type HttpResponseFormat struct {
    Body string
}
func (format HttpResponseFormat) Name() string {
    return "http-response"
}
func (format HttpResponseFormat) Export() map[string]interface{} {
    export := make(map[string]interface{})
    export["body"] = format.Body
    return export
}

func callback(result bool, err error) {
    if result {
        fmt.Println("Async publish successful")
    } else {
        fmt.Println("Async publish failed: " + err.Error())
    }
}

func main() {
    // PubControl can be initialized with or without an endpoint configuration.
    // Each endpoint can include optional JWT authentication info.
    // Multiple endpoints can be included in a single configuration.

    // Initialize PubControl with a single endpoint:
    decodedKey, err := base64.StdEncoding.DecodeString("<realmkey>")
    if err != nil {
        panic("Failed to base64 decode the key")
    }
    pub := pubcontrol.NewPubControl([]map[string]interface{} {
            map[string]interface{} {
            "uri": "https://api.fanout.io/realm/<myrealm>",
            "iss": "<myrealm>", 
            "key": decodedKey}})

    // Add new endpoints by applying an endpoint configuration:
    pub.ApplyConfig([]map[string]interface{} {
            map[string]interface{} { "uri": "<myendpoint_uri_1>" },
            map[string]interface{} { "uri": "<myendpoint_uri_2>" }})

    // Remove all configured endpoints:
    pub.RemoveAllClients()

    // Explicitly add an endpoint as a PubControlClient instance:
    client := pubcontrol.NewPubControlClient("<myendpoint_uri>")
    // Optionally set JWT auth: client.SetAuthJwt(<claim>, "<key>")
    // Optionally set basic auth: client.SetAuthBasic("<user>", "<password>")
    pub.AddClient(client)

    // Create an item to publish:
    format := &HttpResponseFormat{Body: "Test Go Publish!!"} 
    item := pubcontrol.NewItem([]pubcontrol.Formatter{format}, "", "")

    // Publish across all configured endpoints:
    err = pub.Publish("<channel>", item)
    if err != nil {
        panic("Sync publish failed with: " + err.Error())
    }
    err = pub.PublishAsync("<channel>", item, callback)
    if err != nil {
        panic("Async publish failed with: " + err.Error())
    }

    // Wait for all async publish calls to complete:
    pub.Finish()
}
```

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
```

Usage
-----

```go
package main

import "pubcontrol"
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
    pc := pubcontrol.NewPubControl([]map[string]interface{} {
            map[string]interface{} {
            "uri": "https://api.fanout.io/realm/<myrealm>",
            "iss": "<myrealm>", 
            "key": decodedKey}})

    // Add new endpoints by applying an endpoint configuration:
    pc.ApplyConfig([]map[string]interface{} {
            map[string]interface{} { "uri": "<myendpoint_uri_1>" },
            map[string]interface{} { "uri": "<myendpoint_uri_2>" }})

    // Remove all configured endpoints:
    pc.RemoveAllClients()

    // Explicitly add an endpoint as a PubControlClient instance:
    client := pubcontrol.NewPubControlClient("<myendpoint_uri>")
    // Optionally set JWT auth: client.SetAuthJwt(<claim>, "<key>")
    // Optionally set basic auth: client.SetAuthBasic("<user>", "<password>")
    pc.AddClient(client)

    // Create an item to publish:
    format := &HttpResponseFormat{Body: "Test Go Publish!!"} 
    item := pubcontrol.NewItem([]pubcontrol.Formatter{format}, "", "")

    // Publish across all configured endpoints:
    err = pc.Publish("<channel>", item)
    if err != nil {
        panic("Sync publish failed with: " + err.Error())
    }
    err = pc.PublishAsync("<channel>", item, callback)
    if err != nil {
        panic("Async publish failed with: " + err.Error())
    }

    // Wait for all async publish calls to complete:
    pc.Finish()
}
```

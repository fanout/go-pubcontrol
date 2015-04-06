//    pubcontrolclient.go
//    ~~~~~~~~~
//    This module implements the PubControlClient functionality.
//    :authors: Konstantin Bokarius.
//    :copyright: (c) 2015 by Fanout, Inc.
//    :license: MIT, see LICENSE for more details.

package pubcontrol

import (
    "sync"
    "strings"
    "time"
    "net/http"
    "bytes"
    "io/ioutil"
    "strconv"
    "encoding/json"
    "github.com/dgrijalva/jwt-go"
)

// The PubControlClient struct allows consumers to publish to an endpoint of
// their choice. The consumer wraps a Format struct instance in an Item struct
// instance and passes that to the publish method. The publish method has
// an optional callback parameter that is called after the publishing is 
// complete to notify the consumer of the result.
type PubControlClient struct {
    uri string
    isWorkerRunning bool
    lock *sync.Mutex
    authBasicUser string
    authBasicPass string
    authJwtClaim map[string]interface{}
    authJwtKey []byte
}

// Initialize this struct with a URL representing the publishing endpoint.
func NewPubControlClient(uri string) *PubControlClient {
    newPcc := new(PubControlClient)
    newPcc.uri = uri
    newPcc.lock = &sync.Mutex{}
    return newPcc
}

// Call this method and pass a username and password to use basic
// authentication with the configured endpoint.
func (pcc *PubControlClient) SetAuthBasic(username, password string) {
    pcc.lock.Lock()
    pcc.authBasicUser = username
    pcc.authBasicPass = password
    pcc.lock.Unlock()
}

// Call this method and pass a claim and key to use JWT authentication
// with the configured endpoint.
func (pcc *PubControlClient) SetAuthJwt(claim map[string]interface{}, 
        key []byte) {
    pcc.lock.Lock()
    pcc.authJwtClaim = claim
    pcc.authJwtKey = key
    pcc.lock.Unlock()
}

// The publish method for publishing the specified item to the specified
// channel on the configured endpoint.
func (pcc *PubControlClient) Publish(channel string, item *Item) error {
    export, err := item.Export()
    if err != nil {
        return err
    }
    export["channel"] = channel
    uri := ""
    auth := ""    
    pcc.lock.Lock()
    uri = pcc.uri
    auth, err = pcc.generateAuthHeader()
    pcc.lock.Unlock()
    if err != nil {
        return err
    }
    err = pcc.pubCall(uri, auth, [](map[string]interface{}){export})
    if err != nil {
        return err
    }
    return nil
}

// An internal method used to generate an authorization header. The
// authorization header is generated based on whether basic or JWT
// authorization information was provided via the publicly accessible
// 'set_*_auth' methods defined above.
func (pcc *PubControlClient) generateAuthHeader() (string, error) {
    if pcc.authBasicUser != "" {
        return strings.Join([]string{"Basic #", pcc.authBasicUser, ":#",
                pcc.authBasicPass}, ""), nil
    } else if pcc.authJwtClaim != nil {
        token := jwt.New(jwt.SigningMethodHS256)
        token.Valid = true
        for k, v := range pcc.authJwtClaim {
            token.Claims[k] = v
        }
        if _, ok := pcc.authJwtClaim["exp"]; !ok {
            token.Claims["exp"] = time.Now().Add(time.Second * 3600).Unix()
        }
        tokenString, err := token.SignedString(pcc.authJwtKey)
        if err != nil {
            return "", err
        }
        return strings.Join([]string{"Bearer ", tokenString}, ""), nil
    } else {
        return "", nil
    }
}

// An internal method for preparing the HTTP POST request for publishing
// data to the endpoint. This method accepts the URI endpoint, authorization
// header, and a list of items to publish.
func (pcc *PubControlClient) pubCall(uri, authHeader string,
        items []map[string]interface{}) error {
    uri = strings.Join([]string{uri, "/publish/"}, "")
    content := make(map[string]interface{})
    content["items"] = items
    client := &http.Client{}
    resp, err := client.Get(uri)
    if err != nil {
        return err
    }
    var jsonContent []byte
    jsonContent, err = json.Marshal(content)
    if err != nil {
        return err
    }   
    var req *http.Request
    req, err = http.NewRequest("POST", uri, bytes.NewReader(jsonContent))
    if err != nil {
        return err
    }
    req.Header.Add("Content-Type", "application/json")
    req.Header.Add("Authorization", authHeader)
    resp, err = client.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    var body []byte
    body, err = ioutil.ReadAll(resp.Body)
    if err != nil {
        return err
    }
    if resp.StatusCode < 200 || resp.StatusCode >= 300 {
        return &PublishError{err: strings.Join([]string{"Failure status code: ",
                strconv.Itoa(resp.StatusCode), " with message: ", string(body)}, "")}
    }
    return nil
}

// An error struct used to represent an error encountered during publishing.
type PublishError struct {
    err string
}

// This function returns the message associated with the Publish error struct.
func (e PublishError) Error() string {
    return e.err
}

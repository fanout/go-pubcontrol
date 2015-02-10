//    pubcontrolclient.go
//    ~~~~~~~~~
//    This module implements the PubControlClient functionality.
//    :authors: Konstantin Bokarius.
//    :copyright: (c) 2015 by Fanout, Inc.
//    :license: MIT, see LICENSE for more details.

package pubcontrol

// TODO: Figure out how to properly import lane and other packages.
import (
    "sync"
    "strings"
    "time"
    "net/http"
    "bytes"
    "io/ioutil"
    "strconv"
    "encoding/json"

    // TODO: Fix later.
    "github.com/oleiade/lane"
    "github.com/dgrijalva/jwt-go"

    // TODO: Remove later.
    "fmt"
)

type PubControlClient struct {
    uri string
    isWorkerRunning bool
    lock *sync.Mutex
    cond *sync.Cond
    condLock *sync.Mutex
    ReqQueue *lane.Deque
    authBasicUser string
    authBasicPass string
    authJwtClaim map[string]interface{}
    authJwtKey []byte
}

func NewPubControlClient(uri string) *PubControlClient {
    newPcc := new(PubControlClient)
    newPcc.uri = uri
    newPcc.lock = &sync.Mutex{}
    newPcc.ReqQueue = lane.NewDeque()
    return newPcc
}

func (pcc *PubControlClient) SetAuthBasic(username, password string) {
    pcc.lock.Lock()
    pcc.authBasicUser = username
    pcc.authBasicPass = password
    pcc.lock.Unlock()
}

func (pcc *PubControlClient) SetAuthJwt(claim map[string]interface{}, 
        key []byte) {
    pcc.lock.Lock()
    pcc.authJwtClaim = claim
    pcc.authJwtKey = key
    pcc.lock.Unlock()
}

func (pcc *PubControlClient) Publish(channel string, item *Item) error {
    export := item.Export()
    export["channel"] = channel
    uri := ""
    auth := ""    
    pcc.lock.Lock()
    uri = pcc.uri
    auth, err := pcc.generateAuthHeader()
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

func (pcc *PubControlClient) ensureThread() {
    if !pcc.isWorkerRunning {
        pcc.isWorkerRunning = true
        pcc.condLock = &sync.Mutex{}
        pcc.cond = sync.NewCond(pcc.condLock)
        go pcc.pubworker()
    }
}

func (pcc *PubControlClient) pubworker() {
}

func (pcc *PubControlClient) queueReq(req Request) {
    pcc.condLock.Lock()
    pcc.ReqQueue.Append(req)
    pcc.cond.Signal()
    pcc.condLock.Unlock()
}

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
    // TODO: Remove later.
    fmt.Println("JSON Content: ", string(jsonContent))    
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

type PublishError struct {
	err string
}

func (e PublishError) Error() string {
	return e.err
}
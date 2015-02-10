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

func (pcc *PubControlClient) ensureThread() {
    if !pcc.isWorkerRunning {
        pcc.isWorkerRunning = true
        pcc.condLock = &sync.Mutex{}
        pcc.cond = sync.NewCond(pcc.condLock)
        go pcc.pubworker()
    }
}

func (pcc *PubControlClient) pubworker() {
    for true {
        fmt.Println("In pubworker")     
        time.Sleep(time.Second * 1)
    }
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
        } else if !token.Valid {
            return "", &TokenError{err: "token is invalid"}
        }
        return strings.Join([]string{"Bearer ", tokenString}, ""), nil
    } else {
        return "", nil
    }
}

type TokenError struct {
	err string
}

func (e TokenError) Error() string {
	return e.err
}


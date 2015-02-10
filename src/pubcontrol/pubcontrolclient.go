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
    "github.com/oleiade/lane"
)

type PubControlClient struct {
    uri string
    lock *sync.Mutex
    cond *sync.Cond
    condLock *sync.Mutex
    ReqQueue *lane.Deque
    authBasicUser string
    authBasicPass string
    authJwtClaim map[string]string
    authJwtKey string
}

func NewPubControlClient(uri string) *PubControlClient {
    newPcc := new(PubControlClient)
    newPcc.uri = uri
    newPcc.lock = &sync.Mutex{}
    newPcc.ReqQueue = lane.NewDeque()
    return newPcc
}

func (pcc PubControlClient) SetAuthBasic(username, password string) {
    pcc.lock.Lock()
    pcc.authBasicUser = username
    pcc.authBasicPass = password
    pcc.lock.Unlock()
}

func (pcc PubControlClient) SetAuthJwt(claim map[string]string, 
        key string) {
    pcc.lock.Lock()
    pcc.authJwtClaim = claim
    pcc.authJwtKey = key
    pcc.lock.Unlock()
}

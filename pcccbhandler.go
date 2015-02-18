//    pcccbhandler.go
//    ~~~~~~~~~
//    This module implements the PubControlClientCallbackHandler struct
//    and features.
//    :authors: Konstantin Bokarius.
//    :copyright: (c) 2015 by Fanout, Inc.
//    :license: MIT, see LICENSE for more details.

package pubcontrol

type PubControlClientCallbackHandler struct {
    NumCalls int
    Success bool
    FirstError error
    Callback func(result bool, err error)
}

func NewPubControlClientCallbackHandler(numCalls int, 
        callback func(result bool, err error)) *PubControlClientCallbackHandler {
    pcccbhandler := new(PubControlClientCallbackHandler)
    pcccbhandler.NumCalls = numCalls
    pcccbhandler.Callback = callback
    pcccbhandler.Success = true
    pcccbhandler.FirstError = nil
    return pcccbhandler
}

func (pcccbhandler *PubControlClientCallbackHandler) Handler(success bool,
        err error) {
    if (!success && pcccbhandler.Success) {
        pcccbhandler.Success = false
        pcccbhandler.FirstError = err
    }
    pcccbhandler.NumCalls -= 1
    if pcccbhandler.NumCalls <= 0 {
        pcccbhandler.Callback(pcccbhandler.Success, pcccbhandler.FirstError)
    }
}

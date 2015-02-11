//    request.go
//    ~~~~~~~~~
//    This module implements the Request struct.
//    :authors: Konstantin Bokarius.
//    :copyright: (c) 2015 by Fanout, Inc.
//    :license: MIT, see LICENSE for more details.

package pubcontrol

type Request struct {
    Type string
    Uri string
    Auth string
    Export map[string]interface{}
    Callback func(result bool, err error)
}

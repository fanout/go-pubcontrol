//    pubcontrol.go
//    ~~~~~~~~~
//    This module implements the PubControl struct and features.
//    :authors: Konstantin Bokarius.
//    :copyright: (c) 2015 by Fanout, Inc.
//    :license: MIT, see LICENSE for more details.

package pubcontrol

import (
	"fmt"
	"runtime"
	"strings"
	"sync"
)

// The PubControl struct allows a consumer to manage a set of publishing
// endpoints and to publish to all of those endpoints via a single publish
// method call. A PubControl instance can be configured either using a
// hash or array of hashes containing configuration information or by
// manually adding PubControlClient instances.
type PubControl struct {
	clients       []*PubControlClient
	clientsRWLock sync.RWMutex
}

// Initialize with or without a configuration. A configuration can be applied
// after initialization via the apply_config method.
func NewPubControl(config []map[string]interface{}) *PubControl {
	pc := new(PubControl)
	pc.clients = make([]*PubControlClient, 0)
	if config != nil && len(config) > 0 {
		pc.ApplyConfig(config)
	}
	return pc
}

// Remove all of the configured PubControlClient instances.
func (pc *PubControl) RemoveAllClients() {
	pc.clientsRWLock.Lock()
	defer pc.clientsRWLock.Unlock()
	pc.clients = make([]*PubControlClient, 0)
}

// Add the specified PubControlClient instance.
func (pc *PubControl) AddClient(pcc *PubControlClient) {
	pc.clientsRWLock.Lock()
	defer pc.clientsRWLock.Unlock()
	pc.clients = append(pc.clients, pcc)
}

// Apply the specified configuration to this PubControl instance. The
// configuration object can either be a hash or an array of hashes where
// each hash corresponds to a single PubControlClient instance. Each hash
// will be parsed and a PubControlClient will be created either using just
// a URI or a URI and JWT authentication information.
func (pc *PubControl) ApplyConfig(config []map[string]interface{}) {
	pc.clientsRWLock.Lock()
	defer pc.clientsRWLock.Unlock()
	for _, entry := range config {
		if _, ok := entry["uri"]; !ok {
			continue
		}
		pcc := NewPubControlClient(entry["uri"].(string))
		if _, ok := entry["iss"]; ok {
			claim := make(map[string]interface{})
			claim["iss"] = entry["iss"]
			switch entry["key"].(type) {
			case string:
				pcc.SetAuthJwt(claim, []byte(entry["key"].(string)))
			case []byte:
				pcc.SetAuthJwt(claim, entry["key"].([]byte))
			}
            continue
		}
        if _, ok := entry["key"]; ok {
            switch entry["key"].(type) {
            case string:
                pcc.SetAuthBearer([]byte(entry["key"].(string)))
            case []byte:
                pcc.SetAuthBearer(entry["key"].([]byte))
        }
		pc.clients = append(pc.clients, pcc)
	}
}

// The publish method for publishing the specified item to the specified
// channel on the configured endpoints. Different endpoints are published to in parallel,
// with this function waiting for them to finish. Any errors (including panics) are aggregated
// into one error.
func (pc *PubControl) Publish(channel string, item *Item) error {
	pc.clientsRWLock.RLock()
	defer pc.clientsRWLock.RUnlock()
	wg := sync.WaitGroup{}
	errCh := make(chan string, len(pc.clients))

	for _, pcc := range pc.clients {
		wg.Add(1)
		client := pcc
		go func() {
			defer func() {
				if err := recover(); err != nil {
					stack := make([]byte, 1024*8)
					stack = stack[:runtime.Stack(stack, false)]
					errCh <- fmt.Sprintf("%s: PANIC: %v\n%s", client.uri, err, stack)
				}
				wg.Done()
			}()

			err := client.Publish(channel, item)
			if err != nil {
				errCh <- fmt.Sprintf("%s: %s", client.uri, strings.TrimSpace(err.Error()))
			}
		}()
	}
	wg.Wait()
	close(errCh)
	errs := make([]string, 0)
	for err := range errCh {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		return fmt.Errorf("%d/%d client(s) failed to publish to channel: %s Errors: [%s]",
			len(errs), len(pc.clients), channel, strings.Join(errs, "],["))
	}
	return nil
}

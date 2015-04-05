//    item.go
//    ~~~~~~~~~
//    This module implements the Item functionality.
//    :authors: Konstantin Bokarius.
//    :copyright: (c) 2015 by Fanout, Inc.
//    :license: MIT, see LICENSE for more details.

package pubcontrol

// The Item struct is a container used to contain one or more format
// implementation instances where each implementation instance is of a
// different type of format. An Item instance may not contain multiple
// implementations of the same type of format. An Item instance is then
// serialized into a hash that is used for publishing to clients.
type Item struct {
    id string
    prevId string
    formats []Formatter
}

// Initialize this struct with either a single Format implementation
// instance or an array of Format implementation instances. Optionally
// specify an ID and/or previous ID to be sent as part of the message
// published to the client.
func NewItem(formats []Formatter, id, prevId string) *Item {
    newItem := new(Item)
    newItem.id = id
    newItem.prevId = prevId
    newItem.formats = formats
    return newItem
}

// The export method serializes all of the formats, ID, and previous ID
// into a hash that is used for publishing to clients. If more than one
// instance of the same type of Format implementation was specified then
// an error will be raised.
func (item *Item) Export() map[string]interface{} {
    out := make(map[string]interface{})
    if item.id != "" {
        out["id"] = item.id
    }
    if item.prevId != "" {
        out["prev-id"] = item.prevId
    }
    for _,format := range item.formats {
        out[format.Name()] = format.Export()
    }
    return out
}

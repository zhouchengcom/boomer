package glocust

import (
	"log"

	"github.com/ugorji/go/codec"
	"github.com/vmihailenco/msgpack"
)

var (
	mh codec.MsgpackHandle
)

type message struct {
	_msgpack struct{}               `msgpack:",asArray"`
	Type     string                 `msgpack: "type"`
	Data     map[string]interface{} `msgpack: "data"`
	NodeID   string                 `msgpack: "node_id"`
}

func newMessage(t string, data map[string]interface{}, nodeID string) (msg *message) {
	return &message{
		Type:   t,
		Data:   data,
		NodeID: nodeID,
	}
}
func (m *message) serialize() (out []byte) {
	b, err := msgpack.Marshal(m)
	if err != nil {
		log.Fatal("[msgpack] encode fail")
	}
	out = b
	return
}

func newMessageFromBytes(raw []byte) *message {
	var newMsg = &message{}
	err := msgpack.Unmarshal(raw, newMsg)
	if err != nil {
		log.Fatal("[msgpack] decode fail")
	}
	return newMsg
}

package glocust

type client interface {
	recv()
	send()
}

var fromMaster = make(chan *message, 100)
var toMaster = make(chan *message, 100)
var disconnectedFromMaster = make(chan bool)

var rpc *string

// +build !goczmq

package glocust

import (
	"fmt"
	"log"
	"net"
	"strings"

	"github.com/zeromq/gomq"
	"github.com/zeromq/gomq/zmtp"
)

type gomqSocketClient struct {
	pushSocket *gomq.Socket
	pullSocket *gomq.Socket
}

func newClient() client {
	log.Println("Boomer is built with gomq support.")
	var message string
	var client client

	client = newZmqClient(*options.masterHost, *options.masterPort)
	message = fmt.Sprintf("Boomer is connected to master(%s:%d|%d) press Ctrl+c to quit.", *options.masterHost, *options.masterPort, *options.masterPort+1)

	log.Println(message)
	return client
}

func newGomqSocket(socketType zmtp.SocketType) *gomq.Socket {
	socket := gomq.NewSocket(false, socketType, nil, zmtp.NewSecurityNull())
	return socket
}

func getNetConn(addr string) net.Conn {
	parts := strings.Split(addr, "://")
	netConn, err := net.Dial(parts[0], parts[1])
	if err != nil {
		log.Fatal(err)
	}
	return netConn
}

func connectSock(socket *gomq.Socket, addr string) {
	netConn := getNetConn(addr)
	zmtpConn := zmtp.NewConnection(netConn)
	_, err := zmtpConn.Prepare(socket.SecurityMechanism(), socket.SocketType(), nil, false, nil)
	if err != nil {
		log.Fatal(err)
	}
	conn := gomq.NewConnection(netConn, zmtpConn)
	socket.AddConnection(conn)
	zmtpConn.Recv(socket.RecvChannel())
}

func newZmqClient(masterHost string, masterPort int) *gomqSocketClient {
	pushAddr := fmt.Sprintf("tcp://%s:%d", masterHost, masterPort)
	pullAddr := fmt.Sprintf("tcp://%s:%d", masterHost, masterPort+1)

	pushSocket := newGomqSocket(zmtp.PushSocketType)
	connectSock(pushSocket, pushAddr)

	pullSocket := newGomqSocket(zmtp.PullSocketType)
	connectSock(pullSocket, pullAddr)

	log.Println("ZMQ sockets connected")

	newClient := &gomqSocketClient{
		pushSocket: pushSocket,
		pullSocket: pullSocket,
	}
	go newClient.recv()
	go newClient.send()
	return newClient
}

func (c *gomqSocketClient) recv() {
	for {
		msg, err := c.pullSocket.Recv()
		if err != nil {
			log.Printf("Error reading: %v\n", err)
		} else {
			msgFromMaster := newMessageFromBytes(msg)
			fromMaster <- msgFromMaster
		}
	}

}

func (c *gomqSocketClient) send() {
	for {
		select {
		case msg := <-toMaster:
			c.sendMessage(msg)
			if msg.Type == "quit" {
				disconnectedFromMaster <- true
			}
		}
	}
}

func (c *gomqSocketClient) sendMessage(msg *message) {
	err := c.pushSocket.Send(msg.serialize())
	if err != nil {
		log.Printf("Error sending: %v\n", err)
	}
}

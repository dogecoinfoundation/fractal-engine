package simulator

import (
	"fmt"
)

type DogeNetNode struct {
	Name      string
	msgChan   chan Message
	peers     []*DogeNetNode
	groupName string
	listeners []MessageListener
}

func (n *DogeNetNode) AddListener(listener MessageListener) {
	n.listeners = append(n.listeners, listener)
}

func (n *DogeNetNode) Gossip(msg Message) {
	fmt.Println("Gossiping message to peers", n.peers)

	for _, peer := range n.peers {
		peer.msgChan <- msg
	}
}

func (n *DogeNetNode) Start() {
	go func() {
		for msg := range n.msgChan {
			for _, listener := range n.listeners {
				listener.OnMessage(msg, n.Name, "dogenet")
			}
		}
	}()
}

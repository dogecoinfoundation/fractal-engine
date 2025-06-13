package simulator

import "fmt"

type DogeCoreNode struct {
	Name      string
	msgChan   chan Message
	listeners []MessageListener
}

func (n *DogeCoreNode) AddListener(listener MessageListener) {
	n.listeners = append(n.listeners, listener)
}

func (n *DogeCoreNode) Send(msg Message) {
	n.msgChan <- msg
}

func (n *DogeCoreNode) Start() {
	go func() {
		for msg := range n.msgChan {
			for _, listener := range n.listeners {
				listener.OnMessage(msg, n.Name, "core")
			}
		}
	}()
}

type DogeCoreListener struct {
	MessageListener
	NodeGroup *NodeGroup
}

func (l *DogeCoreListener) OnMessage(msg Message, nodeName string, fromType string) {
	for _, m := range l.NodeGroup.metadataMempool.messages {
		fmt.Println("Checking message from", nodeName, fromType, msg.Title)
		if msg.Title == m.Title {
			fmt.Println("Confirming message from", nodeName, fromType, msg.Title)
			l.NodeGroup.metadataMempool.RemoveMessage(m)
			l.NodeGroup.AddMessage(msg)
		}
	}
}

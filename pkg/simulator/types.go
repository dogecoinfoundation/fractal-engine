package simulator

type MessageListener interface {
	OnMessage(msg Message, nodeName string, fromType string)
}

type Message struct {
	Title string
}

type NodeGroup struct {
	Name            string
	dogenet         *DogeNetNode
	core            *DogeCoreNode
	rpcServer       *RPCServerNode
	metadataMempool *Mempool
	l1Mempool       *Mempool
	messages        []*Message
}

func (n *NodeGroup) AddMessage(msg Message) {
	n.messages = append(n.messages, &msg)
}

func (n *NodeGroup) GetMessages() []*Message {
	return n.messages
}

package simulator

type RPCServerNode struct {
	Name      string
	msgChan   chan Message
	listeners []MessageListener
}

func (n *RPCServerNode) AddListener(listener MessageListener) {
	n.listeners = append(n.listeners, listener)
}

func (n *RPCServerNode) Send(msg Message) {
	n.msgChan <- msg
}

func (n *RPCServerNode) Start() {
	go func() {
		for msg := range n.msgChan {
			for _, listener := range n.listeners {
				listener.OnMessage(msg, n.Name, "rpc")
			}
		}
	}()
}

type RPCServerListener struct {
	MessageListener
	NodeGroup *NodeGroup
}

func (l *RPCServerListener) OnMessage(msg Message, nodeName string, fromType string) {
	l.NodeGroup.metadataMempool.AddMessage(msg)
}

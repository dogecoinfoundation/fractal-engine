package simulator

type FESimulator struct {
	nodes       []*NodeGroup
	MemPoolSize int
}

func NewFESimulator(memPoolSize int) *FESimulator {
	return &FESimulator{
		nodes:       make([]*NodeGroup, 0),
		MemPoolSize: memPoolSize,
	}
}

func (s *FESimulator) Start() {
	for _, node := range s.nodes {
		rpcListener := &RPCServerListener{
			NodeGroup: node,
		}

		dogeCoreListener := &DogeCoreListener{
			NodeGroup: node,
		}

		node.rpcServer.AddListener(rpcListener)
		node.core.AddListener(dogeCoreListener)

		node.dogenet.Start()
		node.core.Start()
		node.rpcServer.Start()
	}
}

func (s *FESimulator) AddListener(listener MessageListener) {
	for _, node := range s.nodes {
		node.dogenet.AddListener(listener)
		node.core.AddListener(listener)
		node.rpcServer.AddListener(listener)
	}
}

func (s *FESimulator) GetNode(name string) *NodeGroup {
	for _, node := range s.nodes {
		if node.Name == name {
			return node
		}
	}

	return nil
}

func (s *FESimulator) AddNode(name string) {
	s.nodes = append(s.nodes, &NodeGroup{
		Name: name,
		dogenet: &DogeNetNode{
			Name:    name,
			msgChan: make(chan Message),
			peers:   make([]*DogeNetNode, 0),
		},
		core: &DogeCoreNode{
			Name:    name,
			msgChan: make(chan Message),
		},
		rpcServer: &RPCServerNode{
			Name:    name,
			msgChan: make(chan Message),
		},
		metadataMempool: NewMempool(s.MemPoolSize),
		l1Mempool:       NewMempool(s.MemPoolSize),
		messages:        make([]*Message, 0),
	})
}

func (s *FESimulator) MakePeerGroup(nodeNames []string, groupName string) {
	group := make([]*DogeNetNode, len(nodeNames))

	for i, nodeName := range nodeNames {
		for _, node := range s.nodes {
			if node.Name == nodeName {
				node.dogenet.groupName = groupName
				group[i] = node.dogenet
			}
		}
	}

	for _, node := range s.nodes {
		for _, groupNode := range group {
			if node.Name != groupNode.Name && node.dogenet.groupName == groupName {
				node.dogenet.peers = append(node.dogenet.peers, groupNode)
			}
		}
	}

}

func (s *FESimulator) SimulateGossip(nodeName string, msg Message) {
	for _, node := range s.nodes {
		if node.Name == nodeName {
			node.dogenet.Gossip(msg)
		}
	}
}

func (s *FESimulator) SimulateL1Block(nodeName string, msg Message) {
	for _, node := range s.nodes {
		if node.Name == nodeName {
			node.core.Send(msg)
		}
	}
}

func (s *FESimulator) SimulateRPC(nodeName string, msg Message) {
	for _, node := range s.nodes {
		if node.Name == nodeName {
			node.rpcServer.Send(msg)
		}
	}
}

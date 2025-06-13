package main

import (
	"fmt"
	"os"
	"strings"

	"dogecoin.org/fractal-engine/pkg/simulator"
)

type MyListener struct {
}

func (l *MyListener) OnMessage(msg simulator.Message, nodeName string, fromType string) {
	fmt.Printf("Received message from [%s] %s: %s\n", nodeName, fromType, msg.Title)
}

type Instruction struct {
	Type     string
	NodeName string
	Message  simulator.Message
}

type Expectation struct {
	Type  string
	Node  string
	Value int
}

type Scenario struct {
	MempoolSize   int
	MempoolActive bool
	Nodes         []string
	Instructions  []Instruction
	Expectations  []Expectation
}

var SCENARIO_RPC_SAME_NODE_MEMPOOL_DISCARDS = &Scenario{
	MempoolSize:   3,
	MempoolActive: true,
	Nodes: []string{
		"vehicles:car,van,bike",
		"fruits:apple,banana,orange",
	},
	Instructions: []Instruction{
		{Type: "rpc", NodeName: "car", Message: simulator.Message{Title: "MyToken123"}},
		{Type: "rpc", NodeName: "car", Message: simulator.Message{Title: "fake1"}},
		{Type: "rpc", NodeName: "car", Message: simulator.Message{Title: "fake2"}},
		{Type: "rpc", NodeName: "car", Message: simulator.Message{Title: "fake3"}},
		{Type: "core", NodeName: "car", Message: simulator.Message{Title: "MyToken123"}},
	},
	Expectations: []Expectation{
		{
			Type:  "MessageCount",
			Node:  "car",
			Value: 1,
		},
	},
}

func RunScenario(scenario *Scenario) {
	mempoolSize := scenario.MempoolSize
	sim := simulator.NewFESimulator(mempoolSize)

	for _, node := range scenario.Nodes {
		nodeNames := strings.Split(strings.Split(node, ":")[1], ",")
		for _, nodeName := range nodeNames {
			sim.AddNode(nodeName)
		}
	}

	for _, node := range scenario.Nodes {
		nodeNames := strings.Split(strings.Split(node, ":")[1], ",")
		sim.MakePeerGroup(nodeNames, strings.Split(node, ":")[0])
	}

	sim.AddListener(&MyListener{})

	sim.Start()

	for _, instruction := range scenario.Instructions {
		if instruction.Type == "rpc" {
			sim.SimulateRPC(instruction.NodeName, instruction.Message)
		} else if instruction.Type == "core" {
			sim.SimulateL1Block(instruction.NodeName, instruction.Message)
		} else if instruction.Type == "dogenet" {
			sim.SimulateGossip(instruction.NodeName, instruction.Message)
		}
	}

	for _, expectation := range scenario.Expectations {
		if expectation.Type == "MessageCount" {
			node := sim.GetNode(expectation.Node)
			if len(node.GetMessages()) != expectation.Value {
				fmt.Printf("Expected %d messages in %s, got %d\n", expectation.Value, expectation.Node, len(node.GetMessages()))
				os.Exit(1)
			}
		}
	}
}

func main() {
	RunScenario(SCENARIO_RPC_SAME_NODE_MEMPOOL_DISCARDS)
}

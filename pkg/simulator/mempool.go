package simulator

import "fmt"

type Mempool struct {
	messages []*Message
	limit    int
}

func NewMempool(limit int) *Mempool {
	return &Mempool{
		messages: make([]*Message, 0),
		limit:    limit,
	}
}

func (m *Mempool) AddMessage(msg Message) {
	m.messages = append([]*Message{&msg}, m.messages...)

	if len(m.messages) >= m.limit+1 {
		m.messages = m.messages[:len(m.messages)-1]
	}

	fmt.Println("Added message to mempool", m.messages)
}

func (m *Mempool) GetMessage(index int) *Message {
	return m.messages[index]
}

func (m *Mempool) RemoveMessage(msg *Message) {
	for i, message := range m.messages {
		if message == msg {
			m.messages = append(m.messages[:i], m.messages[i+1:]...) // remove the message from the mempool
			break
		}
	}
}

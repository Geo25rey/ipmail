package ipmail

import (
	"bytes"
	"container/list"
	"io/ioutil"
	"ipmail/libipmail/crypto"
	"ipmail/libipmail/util"
	"os"
	"sync"
)

type MessageList interface {
	Add(message crypto.Message)
	Remove(message crypto.Message)
	ForEach(do func(message crypto.Message))
	FromId(id uint64) crypto.Message
	SaveToFile(file string) error
	Len() int
	FromIndex(idx int) crypto.Message
}

type messageList struct {
	mtx  sync.Mutex
	list *list.List
}

func NewMessageList() MessageList {
	result := messageList{
		mtx:  sync.Mutex{},
		list: list.New(),
	}
	return &result
}

func NewMessageListFromFile(file string,
	ipfs util.Cat, identity crypto.SelfIdentity, contacts crypto.ContactsIdentityList,
) MessageList {
	if identity == nil || contacts == nil {
		return nil
	}
	f, err := os.Open(file)
	if err != nil {
		return nil
	}
	defer f.Close()
	result := NewMessageList()
	buf, err := ioutil.ReadAll(f)
	buffer := bytes.NewBuffer(buf)
	for buffer.Len() > 0 {
		var msg crypto.Message
		msg, err = crypto.ReadMessage(buffer, ipfs, identity, contacts)
		if err != nil {
			break
		}
		result.Add(msg)
	}
	if buffer.Len() == 0 && err == nil {
		return result
	} else {
		return nil
	}
}

func (m *messageList) Add(message crypto.Message) {
	m.mtx.Lock()
	m.list.PushBack(message)
	m.mtx.Unlock()
}

func (m *messageList) Remove(message crypto.Message) {
	m.mtx.Lock()
	for elm := m.list.Back(); elm != nil; elm = elm.Prev() {
		compare := elm.Value.(crypto.Message)
		if compare.Id() == message.Id() {
			m.list.Remove(elm)
			break
		}
	}
	m.mtx.Unlock()
}

func (m *messageList) Len() int {
	m.mtx.Lock()
	defer m.mtx.Unlock()
	return m.list.Len()
}

func (m *messageList) ForEach(do func(message crypto.Message)) {
	m.mtx.Lock()
	for elm := m.list.Front(); elm != nil; elm = elm.Next() {
		do(elm.Value.(crypto.Message))
	}
	m.mtx.Unlock()
}

func (m *messageList) FromIndex(idx int) crypto.Message {
	m.mtx.Lock()
	defer m.mtx.Unlock()
	var result crypto.Message = nil
	for i, elm := 0, m.list.Front(); elm != nil; elm = elm.Next() {
		msg := elm.Value.(crypto.Message)
		if i == idx {
			result = msg
			break
		}
		i++
	}
	return result
}

func (m *messageList) FromId(id uint64) crypto.Message {
	var result crypto.Message = nil
	m.mtx.Lock()
	for elm := m.list.Front(); elm != nil; elm = elm.Next() {
		msg := elm.Value.(crypto.Message)
		if msg.Id() == id {
			result = msg
			break
		}
	}
	m.mtx.Unlock()
	return result
}

func (m *messageList) SaveToFile(file string) error {
	f, err := os.Create(file)
	if err != nil {
		return err
	}
	defer f.Close()
	m.ForEach(func(message crypto.Message) {
		if err != nil {
			return
		}
		err = message.Serialize(f)
	})
	return err
}

package ipmail

import (
	"bytes"
	"container/list"
	"io/ioutil"
	"ipmail/ipmail/crypto"
	"os"
	"sync"
)

type MessageList interface {
	Add(message crypto.Message)
	ForEach(do func(message crypto.Message))
	FromId(id uint64) crypto.Message
	SaveToFile(file string) error
}

type messageList struct {
	mtx sync.Mutex
	list *list.List
}

func NewMessageList() MessageList {
	result := messageList{
		mtx: sync.Mutex{},
		list: list.New(),
	}
	return &result
}

func NewMessageListFromFile(file string,
	identity crypto.SelfIdentity, contacts crypto.ContactsIdentityList,
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
		msg, err = crypto.ReadMessage(buffer, identity, contacts)
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

func (m *messageList) ForEach(do func(message crypto.Message))  {
	m.mtx.Lock()
	for elm := m.list.Front(); elm != nil; elm = elm.Next() {
		do(elm.Value.(crypto.Message))
	}
	m.mtx.Unlock()
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

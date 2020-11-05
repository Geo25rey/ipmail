package ipmail

import (
	iface "github.com/ipfs/interface-go-ipfs-core"
	"io"
	"sync"
)

type ReceiveFunction func(message iface.PubSubMessage)

type Receiver interface {
	OnMessage(function ReceiveFunction, isAsync bool)
	io.Closer
	subscriptionHandler()
}

type receiverImpl struct {
	subscription   iface.PubSubSubscription
	ipfs           *Ipfs
	waitForHandler sync.Mutex
	messageHandler ReceiveFunction
	isAsync        bool
}

func (r *receiverImpl) OnMessage(function ReceiveFunction, isAsync bool) {
	if function != nil {
		startHandler := r.messageHandler == nil
		r.messageHandler = function
		r.isAsync = isAsync
		if startHandler {
			r.waitForHandler.Unlock()
		}
	}
}

func (r *receiverImpl) Close() error {
	err := r.subscription.Close()
	r.waitForHandler.Unlock()
	return err
}

func (r *receiverImpl) subscriptionHandler() {
	var err error
	var message iface.PubSubMessage
	r.waitForHandler.Lock()
	if r.messageHandler == nil {
		r.waitForHandler.Unlock()
		return
	}
	for message, err = r.subscription.Next(r.ipfs.Context()); err == nil && message != nil; message,
		err = r.subscription.Next(r.ipfs.Context()) {
		if r.isAsync {
			go r.messageHandler(message)
		} else {
			r.messageHandler(message)
		}
	}
}

func NewReceiver(topic string, ipfs *Ipfs) (Receiver, error) {
	result := receiverImpl{}
	subscription, err := ipfs.Subscribe(topic)
	if err != nil {
		return nil, err
	}
	result.subscription = subscription
	result.ipfs = ipfs
	result.waitForHandler.Lock()
	go result.subscriptionHandler()
	return &result, nil
}

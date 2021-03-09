package ipmail

import (
	"bytes"
	_ "crypto/sha512"
	"errors"
	gpg "github.com/Geo25rey/crypto/openpgp"
	armor "github.com/Geo25rey/crypto/openpgp/armor"
	"github.com/ipfs/go-cid"
	"io"
	crypto2 "ipmail/libipmail/crypto"
	"ipmail/libipmail/util"
)

type Sender interface {
	Send(content io.Reader, sign bool, from *gpg.Entity, to ...*gpg.Entity) (cid.Cid, error)
	publishMessage(cid cid.Cid) error
}

type senderCtx struct {
	ipfs *Ipfs
}

func NewSender(ipfs *Ipfs) Sender {
	return &senderCtx{ipfs: ipfs}
}

func (this *senderCtx) Send(content io.Reader, sign bool, from *gpg.Entity, to ...*gpg.Entity) (cid.Cid, error) {
	for _, entity := range to {
		if entity == nil {
			return cid.Undef, errors.New("All message recipients must be in your contacts")
		}
	}

	var signer *gpg.Entity

	if sign {
		signer = from
	} else {
		signer = nil
	}

	buf := bytes.NewBuffer(make([]byte, 0))

	w2, err := armor.Encode(buf, crypto2.MessageEncoding, make(map[string]string))
	if err != nil {
		return cid.Undef, err
	}

	w3, err := gpg.Encrypt(w2, to, signer, nil, util.DefaultEncryptionConfig())
	if err != nil {
		return cid.Undef, err
	}

	_, err = io.Copy(w3, content)
	if err != nil {
		return cid.Undef, err
	}

	err = w3.Close()
	if err != nil {
		return cid.Undef, err
	}

	err = w2.Close()
	if err != nil {
		return cid.Undef, err
	}

	path, err := this.ipfs.AddFromReader(buf)
	if err != nil {
		return cid.Undef, err
	}

	return path.Cid(), this.publishMessage(path.Cid())
}

func (this *senderCtx) publishMessage(cid cid.Cid) error {
	toSend := make([]byte, 0)
	toSend = append(toSend, crypto2.MessageCidPrefix...)
	toSend = append(toSend, cid.Bytes()...)
	toSend = append(toSend, crypto2.MessageCidPostfix...)
	return this.ipfs.Publish(crypto2.MessageTopicName, toSend)
}

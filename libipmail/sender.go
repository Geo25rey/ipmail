package ipmail

import (
	_ "crypto/sha512"
	"errors"
	gpg "github.com/Geo25rey/crypto/openpgp"
	armor "github.com/Geo25rey/crypto/openpgp/armor"
	"github.com/ipfs/go-cid"
	"io"
	"io/ioutil"
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
	r, w := io.Pipe()
	defer r.Close()

	var signer *gpg.Entity

	if sign {
		signer = from
	} else {
		signer = nil
	}

	go func() {
		defer w.Close()
		w2, err := armor.Encode(w, crypto2.MessageEncoding, make(map[string]string))
		if err != nil {
			return
		}
		defer w2.Close()

		w3, err := gpg.Encrypt(w2, to, signer, nil, util.DefaultEncryptionConfig())
		if err != nil {
			return
		}
		defer w3.Close()

		_, err = io.Copy(w3, content)
		if err != nil {
			return
		}
	}()

	b, err := ioutil.ReadAll(r)
	if err != nil {
		return cid.Cid{}, err
	}
	path, err := this.ipfs.AddFromBytes(b)
	//path, err := this.ipfs.AddFromReader(r)
	if err != nil {
		return cid.Cid{}, err
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

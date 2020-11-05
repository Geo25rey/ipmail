package crypto

import (
	"errors"
	gpg "github.com/Geo25rey/crypto/openpgp"
	"github.com/Geo25rey/crypto/openpgp/armor"
	"github.com/Geo25rey/crypto/openpgp/packet"
	"github.com/libp2p/go-libp2p-core/peer"
	"io"
	"io/ioutil"
	"ipmail/ipmail/util"
	"reflect"
	"strconv"
	"strings"
)

type Message interface {
	FromName() string
	FromEmail() string
	Data() []byte
	String() string
	Id() uint64
	Serialize(writer io.Writer) error
	IsFrom(entity *gpg.Entity) bool
}

const (
	MessageEncoding = "9c9ek45n65o2radWERjoi"

	MessageTopicName = "Mail"

	MessageCidPrefix  = "2jf9viv549cjksdfjo932"
	MessageCidPostfix = "0anrnLKj34kvlPWnx1as"
)

type message struct {
	encryptedData []byte
	decryptedData []byte
	from          *packet.UserId
	fromEntity    *gpg.Entity
	id            uint64
	origin        peer.ID
}

func NewMessage(encryptedData []byte, id uint64, origin peer.ID,
	identity SelfIdentity, contacts ContactsIdentityList, prompt gpg.PromptFunction) Message {
	result := message{
		encryptedData: encryptedData,
		decryptedData: nil,
		from:          nil,
		id:            id,
		origin:        origin,
	}
	err := result.decrypt(identity, contacts, prompt)
	if err != nil {
		println(err.Error())
		return nil
	}
	return &result
}

func (m *message) decrypt(identity SelfIdentity, contacts ContactsIdentityList, prompt gpg.PromptFunction) error {
	if m.decryptedData == nil {
		r, w := io.Pipe()
		defer r.Close()
		go func() {
			defer w.Close()
			total := 0
			for total < len(m.encryptedData) {
				write, _ := w.Write(m.encryptedData[total:])
				total += write
			}
		}()
		decode, err := armor.Decode(r)
		if err != nil {
			return err
		}
		if strings.Compare(decode.Type, MessageEncoding) != 0 {
			return errors.New("data not encrypted as a message")
		}
		var keyring gpg.KeyRing = append(identity.EntityList(), contacts.ToArray()...) //append(contacts.ToArray(), identity.EntityList()...))
		readMessage, err := gpg.ReadMessage(decode.Body, keyring, prompt, util.DefaultEncryptionConfig())
		if err != nil {
			return err
		}
		body := readMessage.UnverifiedBody
		m.decryptedData, err = ioutil.ReadAll(body)
		if err != nil {
			return err
		}
		_, err = ioutil.ReadAll(decode.Body) // allow the writer in goroutine to close
		if err != nil {                      // fail if I/O error or armor verification fails
			return err
		}
		m.fromEntity = readMessage.SignedBy.Entity
		mapRange := reflect.ValueOf(m.fromEntity.Identities).MapRange()
		mapRange.Next()
		m.from = mapRange.Value().Interface().(*gpg.Identity).UserId
		return nil
	}
	return errors.New("already decrypted message")
}

func (m *message) FromName() string {
	return m.from.Name
}

func (m *message) FromEmail() string {
	return m.from.Email
}

func (m *message) IsFrom(entity *gpg.Entity) bool {
	return util.EntitiesEqual(m.fromEntity, entity)
}

func (m *message) Data() []byte {
	return append(make([]byte, len(m.decryptedData)), m.decryptedData...)
}

func (m *message) String() string {
	result := make([]byte, 0)
	result = append(result, "ID: "...)
	result = strconv.AppendUint(result, m.id, 10)
	result = append(result, " Name: "+m.FromName()...)
	return string(result)
}

func (m *message) Id() uint64 {
	return m.id
}

func (m *message) Serialize(w io.Writer) error {
	_, err := w.Write(util.Int64ToBytes(int64(len(m.encryptedData))))
	if err != nil {
		return err
	}
	_, err = w.Write(m.encryptedData)
	if err != nil {
		return err
	}
	_, err = w.Write(util.Int64ToBytes(int64(m.origin.Size())))
	if err != nil {
		return err
	}
	marshal, err := m.origin.Marshal()
	if err != nil {
		return err
	}
	_, err = w.Write(marshal)
	if err != nil {
		return err
	}
	_, err = w.Write(util.Uint64ToBytes(m.id))
	return err
}

func ReadMessage(r io.Reader, identity SelfIdentity, contacts ContactsIdentityList) (Message, error) {
	if contacts == nil {
		return nil, errors.New("contacts may not be nil")
	}
	result := message{}
	intBuf := make([]byte, 8)
	read, err := r.Read(intBuf)
	if err != nil {
		return nil, err
	}
	if read < len(intBuf) {
		return nil, errors.New("unexpected EOF")
	}
	dataSize, _ := util.BytesToInt64(intBuf)
	dataBuf := make([]byte, dataSize)
	read, err = r.Read(dataBuf)
	if err != nil {
		return nil, err
	}
	if read < len(dataBuf) {
		return nil, errors.New("unexpected EOF")
	}
	result.encryptedData = dataBuf
	read, err = r.Read(intBuf)
	if err != nil {
		return nil, err
	}
	if read < len(intBuf) {
		return nil, errors.New("unexpected EOF")
	}
	marshalSize, _ := util.BytesToInt64(intBuf)
	marshalBuf := make([]byte, marshalSize)
	read, err = r.Read(marshalBuf)
	if err != nil {
		return nil, err
	}
	if read < len(marshalBuf) {
		return nil, errors.New("unexpected EOF")
	}
	err = (&result.origin).Unmarshal(marshalBuf)
	if err != nil {
		return nil, err
	}
	read, err = r.Read(intBuf)
	if err != nil {
		return nil, err
	}
	if read < len(intBuf) {
		return nil, errors.New("unexpected EOF")
	}
	result.id, _ = util.BytesToUint64(intBuf)
	err = result.decrypt(identity, contacts, nil)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

package crypto

import (
	"bytes"
	"container/list"
	gpg "github.com/Geo25rey/crypto/openpgp"
	"github.com/Geo25rey/crypto/openpgp/packet"
	"io"
	"io/ioutil"
	"strings"
)

type IdentityList interface {
	GetByName(name string) IdentityList
	GetByEmail(email string) IdentityList
	GetByPublicKey(key packet.PublicKey) (IdentityList, error)
	GetAny() *gpg.Entity
	Add(entities ...*gpg.Entity)
	AddFromKeyRing(ring gpg.KeyRing)
	ToArray() []*gpg.Entity
	ForEach(do func(entity *gpg.Entity))
}

type identityList struct {
	list *list.List
}

func NewIdentityList(entities ...*gpg.Entity) IdentityList {
	result := identityList{}
	result.list = list.New()
	result.Add(entities...)
	return &result
}

func (i *identityList) ToArray() []*gpg.Entity {
	result := make([]*gpg.Entity, 0)
	for elm := i.list.Front(); elm != nil; elm = elm.Next() {
		entity := elm.Value.(*gpg.Entity)
		result = append(result, entity)
	}
	return result
}

func (i *identityList) Add(entities ...*gpg.Entity) {
	for _, entity := range entities {
		i.list.PushBack(entity)
	}
}

func (i *identityList) AddFromKeyRing(ring gpg.KeyRing) {
	keys := ring.DecryptionKeys()
	for _, key := range keys {
		i.Add(key.Entity)
	}
}

func (i *identityList) ForEach(do func(entity *gpg.Entity)) {
	for elm := i.list.Front(); elm != nil; elm = elm.Next() {
		do(elm.Value.(*gpg.Entity))
	}
}

func (i *identityList) GetAny() *gpg.Entity {
	if front := i.list.Front(); front != nil {
		return front.Value.(*gpg.Entity)
	}
	return nil
}

func (i *identityList) GetByName(name string) IdentityList {
	result := NewIdentityList()
	for elm := i.list.Front(); elm != nil; elm = elm.Next() {
		entity := elm.Value.(*gpg.Entity)
		for _, identity := range entity.Identities {
			if strings.Compare(identity.UserId.Name, name) == 0 {
				result.Add(entity)
				break
			}
		}
	}
	return result
}

func (i *identityList) GetByEmail(email string) IdentityList {
	result := NewIdentityList()
	for elm := i.list.Front(); elm != nil; elm = elm.Next() {
		entity := elm.Value.(*gpg.Entity)
		for _, identity := range entity.Identities {
			if strings.Compare(identity.UserId.Email, email) == 0 {
				result.Add(entity)
				break
			}
		}
	}
	return result
}

func (i *identityList) GetByPublicKey(key packet.PublicKey) (IdentityList, error) {
	result := NewIdentityList()
	r, w := io.Pipe()
	err := key.Serialize(w)
	if err != nil {
		return nil, err
	}
	compareTo, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	for elm := i.list.Front(); elm != nil; elm = elm.Next() {
		entity := elm.Value.(*gpg.Entity)
		err := entity.PrimaryKey.Serialize(w)
		if err != nil {
			return nil, err
		}
		comparing, err := ioutil.ReadAll(r)
		if err != nil {
			return nil, err
		}
		if bytes.Compare(comparing, compareTo) == 0 {
			result.Add(entity)
		}
	}
	return result, nil
}

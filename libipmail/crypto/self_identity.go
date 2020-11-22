package crypto

import (
	gpg "github.com/Geo25rey/crypto/openpgp"
	"ipmail/libipmail/util"
	"os"
)

type SelfIdentity interface {
	DefaultIdentity() *gpg.Entity
	EntityList() gpg.EntityList
	SaveToFile(file string) error
}

type selfIdentity struct {
	identities      IdentityList
	defaultIdentity *gpg.Entity
}

func NewSelfIdentity(name string, comment string, email string) (SelfIdentity, error) {
	result := selfIdentity{}
	var err error
	result.defaultIdentity, err = gpg.NewEntity(name, comment, email, util.DefaultEncryptionConfig())
	if err != nil {
		return nil, err
	}
	result.identities = NewIdentityList(result.defaultIdentity)
	return &result, nil
}

func (s *selfIdentity) DefaultIdentity() *gpg.Entity {
	if s.defaultIdentity != nil {
		return s.defaultIdentity
	}
	return s.identities.GetAny()
}

func (s *selfIdentity) EntityList() gpg.EntityList {
	return s.identities.ToArray()
}

func NewSelfIdentityFromFile(path string) SelfIdentity {
	file, err := os.Open(path)
	if err != nil {
		return nil
	}
	entities, err := util.LoadEntities(file)
	if err != nil {
		return nil
	}
	return &selfIdentity{
		defaultIdentity: entities[0],
		identities:      NewIdentityList(entities...),
	}
}

func (s *selfIdentity) SaveToFile(path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	return util.SaveEntitiesPrivate(file, append(s.EntityList(), s.defaultIdentity)...)
}

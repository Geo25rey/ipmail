package crypto

import (
	gpg "github.com/Geo25rey/crypto/openpgp"
	"ipmail/libipmail/util"
	"os"
)

type ContactsIdentityList interface {
	IdentityList
	SaveToFile(file string) error
}

type contactsIdentityList struct {
	IdentityList
}

func NewContactsIdentityList(entities gpg.EntityList) ContactsIdentityList {
	result := contactsIdentityList{}
	result.IdentityList = NewIdentityList(entities...)
	return &result
}

func NewContactsIdentityListFromFile(file string) (ContactsIdentityList, error) {
	result := contactsIdentityList{}
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	entities, err := util.LoadEntities(f)
	if err != nil {
		return nil, err
	}
	result.IdentityList = NewIdentityList(entities...)
	return &result, nil
}

func (c *contactsIdentityList) SaveToFile(file string) error {
	f, err := os.Create(file)
	if err != nil {
		return err
	}
	return util.SaveEntities(f, c.ToArray()...)
}

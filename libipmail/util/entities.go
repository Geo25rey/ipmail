package util

import (
	"bytes"
	"crypto"
	"encoding/base64"
	"fmt"
	gpg "github.com/Geo25rey/crypto/openpgp"
	"github.com/Geo25rey/crypto/openpgp/packet"
	"github.com/ipfs/go-cid"
	"github.com/ipfs/interface-go-ipfs-core/path"
	"io"
	"io/ioutil"
	"os"
	"reflect"
	"strings"
)

type Cat interface {
	Cat(resolved path.Resolved) ([]byte, error)
}

func ParseEntity(str string, ipfs Cat) (*gpg.Entity, error) {
	var b []byte = nil
	if strings.HasPrefix(str, "file:") {
		str = strings.TrimPrefix(str, "file:")
		f, err := os.Open(str)
		if err != nil {
			return nil, err
		}
		defer f.Close()
		b, err = ioutil.ReadAll(f)
		if err != nil {
			return nil, err
		}
	} else if strings.HasPrefix(str, "ipfs:") {
		parse, err := cid.Parse([]byte(str[5:]))
		if err != nil {
			return nil, err
		}
		b, err = ipfs.Cat(path.IpfsPath(parse))
		if err != nil {
			return nil, err
		}
	} else if strings.HasPrefix(str, "bin:") {
		b = []byte(strings.TrimPrefix(str, "bin:"))
	} else if strings.HasPrefix(str, "base64:") {
		str = strings.TrimPrefix(str, "base64:")
		buf := bytes.NewBufferString(str)
		var err error
		b, err = ioutil.ReadAll(base64.NewDecoder(base64.StdEncoding, buf))
		if err != nil {
			return nil, err
		}
	} else {
		return nil, fmt.Errorf("\"%s\" has an invalid prefix", str)
	}
	buf := bytes.NewBuffer(b)
	reader := packet.NewReader(buf)
	return gpg.ReadEntity(reader)
}

func EntityToString(entity *gpg.Entity) string {
	mapRange := reflect.ValueOf(entity.Identities).MapRange()
	mapRange.Next()
	user := mapRange.Value().Interface().(*gpg.Identity).UserId
	result := user.Name
	if len(user.Comment) > 0 {
		result += " (" + user.Comment + ")"
	}
	if len(user.Email) > 0 {
		return result + " <" + user.Email + ">"
	}
	return result
}

func SaveEntities(w io.Writer, entities ...*gpg.Entity) error {
	entityList := gpg.EntityList(entities)
	for _, v := range entityList {
		err := v.Serialize(w)
		if err != nil {
			return err
		}
	}
	return nil
}

func SaveEntitiesPrivate(w io.Writer, entities ...*gpg.Entity) error {
	entityList := gpg.EntityList(entities)
	for _, v := range entityList {
		err := v.SerializePrivate(w, DefaultEncryptionConfig())
		if err != nil {
			return err
		}
	}
	return nil
}

func LoadEntities(r io.Reader) (gpg.EntityList, error) {
	result := make(gpg.EntityList, 0)
	var err error
	for true {
		var entity *gpg.Entity
		entity, err = gpg.ReadEntity(packet.NewReader(r))
		if err != nil {
			break
		}
		result = append(result, entity)
	}
	return result, nil
}

func EntitiesEqual(entities ...*gpg.Entity) bool {
	if len(entities) < 2 {
		return true
	}

	compareTo := entities[0]
	var b []byte = nil
	if compareTo != nil {
		r, w := io.Pipe()
		defer r.Close()
		go func() {
			defer w.Close()
			compareTo.Serialize(w)
		}()
		b, _ = ioutil.ReadAll(r)
	}

	for _, entity := range entities[1:] {
		if (compareTo == nil && entity != nil) || (compareTo != nil && entity == nil) {
			return false
		}

		b1, _ := func() ([]byte, error) {
			r1, w1 := io.Pipe()
			defer r1.Close()
			go func() {
				defer w1.Close()
				entity.Serialize(w1)
			}()
			return ioutil.ReadAll(r1)
		}()
		if !bytes.Equal(b, b1) {
			return false
		}
	}

	return true
}

func DefaultEncryptionConfig() *packet.Config {
	return &packet.Config{
		DefaultHash:            crypto.SHA512,
		DefaultCipher:          packet.CipherAES256,
		DefaultCompressionAlgo: packet.CompressionNone,
		RSABits:                4096,
	}
}

package main

import (
	"bytes"
	"crypto"
	gpg "github.com/Geo25rey/crypto/openpgp"
	"github.com/Geo25rey/crypto/openpgp/armor"
	"github.com/Geo25rey/crypto/openpgp/packet"
	"io"
	"io/ioutil"
	"strings"
	"testing"
)

func TestSerialEncryption(t *testing.T) {
	DefaultEncryptionConfig := func() *packet.Config {
		return &packet.Config{
			DefaultHash:            crypto.SHA512,
			DefaultCipher:          packet.CipherAES256,
			DefaultCompressionAlgo: packet.CompressionNone,
			RSABits:                4096,
		}
	}

	const MessageEncoding = "Message"

	entity, _ := gpg.NewEntity("Name", "", "", DefaultEncryptionConfig())

	// Remove Private Keys to share
	buf := bytes.NewBuffer(make([]byte, 0))
	err := entity.Serialize(buf)
	publicEntity, err := gpg.ReadEntity(packet.NewReader(buf))
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	messageSent := "Some message"
	content := bytes.NewBuffer([]byte(messageSent))

	println("Creating Message")
	// Create an Encrypted Message to send
	to := gpg.EntityList{publicEntity}
	buf = bytes.NewBuffer(make([]byte, 0))

	w2, err := armor.Encode(buf, MessageEncoding, make(map[string]string))
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	w3, err := gpg.Encrypt(w2, to, nil, nil, DefaultEncryptionConfig())
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	_, err = io.Copy(w3, content)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	err = w3.Close()
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	err = w2.Close()
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	decode, err := armor.Decode(buf)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	messageDetails, err := gpg.ReadMessage(decode.Body, gpg.EntityList{entity}, nil, DefaultEncryptionConfig())
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	messageRead, err := ioutil.ReadAll(messageDetails.UnverifiedBody)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if strings.Compare(messageSent, string(messageRead)) != 0 {
		t.Errorf("Message sent isn't the same as message read\n'%s' != '%s'", messageSent, string(messageRead))
		t.FailNow()
	}
}

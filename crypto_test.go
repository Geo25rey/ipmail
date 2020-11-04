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

func TestSerialEncryption(t *testing.T)  {
	DefaultEncryptionConfig := func () *packet.Config {
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
	r, w := io.Pipe()
	defer r.Close()
	go func() {
		defer w.Close()
		entity.Serialize(w)
	}()
	publicEntity, err := gpg.ReadEntity(packet.NewReader(r))
	if err != nil {
		t.Error(err)
	}

	messageSent := "Some message"
	content := bytes.NewBuffer([]byte(messageSent))

	println("Creating Message")
	// Create an Encrypted Message to send
	to := gpg.EntityList{publicEntity}
	r, w = io.Pipe()
	defer r.Close()

	go func() {
		defer w.Close()
		w2, err := armor.Encode(w, MessageEncoding, make(map[string]string))
		if err != nil {
			t.Error(err)
		}
		defer w2.Close()

		w3, err := gpg.Encrypt(w2, to, nil, nil, DefaultEncryptionConfig())
		if err != nil {
			t.Error(err)
			return
		}
		defer w3.Close()

		_, err = io.Copy(w3, content)
		if err != nil {
			t.Error(err)
		}
	}()

	decode, err := armor.Decode(r)
	if err != nil {
		t.Error(err)
	}
	messageDetails, err := gpg.ReadMessage(decode.Body, gpg.EntityList{entity}, nil, DefaultEncryptionConfig())
	if err != nil {
		t.Error(err)
	}
	messageRead, err := ioutil.ReadAll(messageDetails.UnverifiedBody)
	if err != nil {
		t.Error(err)
	}
	if strings.Compare(messageSent, string(messageRead)) != 0 {
		t.Errorf("Message sent isn't the same as message read\n'%s' != '%s'", messageSent, string(messageRead))
	}
}

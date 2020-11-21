package views

import (
	"bytes"
	"container/list"
	"fyne.io/fyne"
	"fyne.io/fyne/canvas"
	"fyne.io/fyne/container"
	"fyne.io/fyne/widget"
	"github.com/ipfs/go-cid"
	"github.com/skip2/go-qrcode"
	"image"
	"ipmail/ipmail/crypto"
	"time"
)

func generateQRCode(index int, hashList *list.List) (fyne.CanvasObject, error, cid.Cid) {
	var qr *qrcode.QRCode = nil
	var err error = nil
	id := cid.Undef
	for i, hash := 0, hashList.Front(); hash != nil; i, hash = i+1, hash.Next() {
		if hash.Value == nil && i != index {
			continue
		}
		id = hash.Value.(cid.Cid)
		qr, err = qrcode.New(id.String(), qrcode.Low)
		if err != nil {
			qr = nil
			break
		}
	}
	if qr == nil {
		return nil, err, id
	}
	png, err := qr.PNG(256)
	if err != nil {
		return nil, err, id
	}
	buf := bytes.NewBuffer(png)
	decode, _, err := image.Decode(buf)
	if err != nil {
		return nil, err, id
	}
	img := canvas.NewImageFromImage(decode)
	img.Resize(fyne.NewSize(256, 256))
	img.SetMinSize(fyne.NewSize(256, 256))
	img.FillMode = canvas.ImageFillContain
	return img, nil, id
}

func MakeContactsList(contacts crypto.ContactsIdentityList, hashList *list.List) fyne.CanvasObject {
	var l *widget.List
	l = widget.NewList(func() int {
		return len(contacts.ToArray())
	}, func() fyne.CanvasObject {
		vbox := container.NewVBox(
			widget.NewLabel(""),
			widget.NewLabel(""),
			container.NewVBox(),
			widget.NewLabel(""),
		)
		return widget.NewAccordion(
			widget.NewAccordionItem("",
				vbox,
			),
		)
	}, func(index widget.ListItemID, c fyne.CanvasObject) {
		item := c.(*widget.Accordion).Items[0]
		box := item.Detail.(*fyne.Container)
		publicKeyLabel := box.Objects[0].(*widget.Label)
		invalidLabel := box.Objects[3].(*widget.Label)
		cidLabel := box.Objects[1].(*widget.Label)
		contact := contacts.ToArray()[index]
		for _, identity := range contact.Identities {
			id := identity.UserId
			if len(id.Name) == len(id.Comment) && len(id.Name) == len(id.Email) && len(id.Name) == 0 {
				item.Title = "Unnamed Contact"
			} else {
				item.Title = id.Id
			}
			publicKeyLabel.Text = string(append([]byte("Public Key: "), contact.PrimaryKey.KeyIdString()...))
			if identity.SelfSignature.KeyExpired(time.Now()) {
				invalidLabel.Text = "Key has expired"
			} else {
				invalidLabel.Text = ""
			}
			img, err, cID := generateQRCode(index, hashList)
			if cID != cid.Undef {
				cidLabel.Text = string(append([]byte("Content ID: "), cID.String()...))
			}
			if err == nil {
				box.Objects[2] = img
			} else {
				box.Objects[2] = container.NewVBox()
			}
			break // TODO show more than just the default identity
		}
		c.Refresh()
	})
	return l
}

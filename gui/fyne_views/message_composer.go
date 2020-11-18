package views

import (
	"bytes"
	"fyne.io/fyne"
	"fyne.io/fyne/container"
	"fyne.io/fyne/dialog"
	"fyne.io/fyne/theme"
	"fyne.io/fyne/widget"
	"ipmail/ipmail"
	"ipmail/ipmail/crypto"
)

func makeToolBar(w fyne.Window, send func()) *widget.Toolbar {
	return widget.NewToolbar(widget.NewToolbarAction(theme.CancelIcon(), func() {
		w.Close()
	}), widget.NewToolbarSpacer(),
		widget.NewToolbarAction(theme.MailSendIcon(), send))
}

func MakeMessageComposer(w fyne.Window,
	identity crypto.SelfIdentity, contacts crypto.ContactsIdentityList, sender ipmail.Sender) fyne.CanvasObject {
	subject := widget.NewEntry()
	subject.PlaceHolder = "Subject"
	recipient := widget.NewEntry()
	// TODO add error checking for recipient
	// TODO add support for multiple recipients
	// TODO add recipient auto complete
	recipient.PlaceHolder = "Recipient"
	body := widget.NewMultiLineEntry()
	// TODO add image support using canvas
	//textCanvas := canvas.NewText("", color.Black)
	//var r io.Reader
	//imageCanvas := canvas.NewImageFromImage(image.Decode(r))
	toolbar := makeToolBar(w, func() {
		r := bytes.NewReader(append(
			append(
				append([]byte("subject:"), subject.Text...),
				'\n'),
			body.Text...))
		_, err := sender.Send(r, true, identity.DefaultIdentity(),
			identity.DefaultIdentity(), contacts.GetByName(recipient.Text).GetAny())
		if err != nil {
			errDialog := dialog.NewError(err, w)
			errDialog.Show()
		} else {
			w.Close()
		}
	})
	return container.NewVBox(toolbar, subject, recipient, body)
}

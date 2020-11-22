package views

import (
	"fyne.io/fyne"
	"fyne.io/fyne/container"
	"fyne.io/fyne/theme"
	"fyne.io/fyne/widget"
	"ipmail/libipmail"
	"ipmail/libipmail/crypto"
)

func MakeContactRequestsManager(
	contacts crypto.ContactsIdentityList, requests ipmail.MessageList,
) fyne.CanvasObject {
	var list *widget.List
	list = widget.NewList(
		func() int {
			return requests.Len()
		},
		func() fyne.CanvasObject {
			return container.NewHBox(widget.NewLabel("Template Object"),
				widget.NewButtonWithIcon("", theme.CheckButtonCheckedIcon(), nil),
				widget.NewButtonWithIcon("", theme.CancelIcon(), nil))
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			msg := requests.FromIndex(id)
			objs := item.(*fyne.Container).Objects
			objs[0].(*widget.Label).SetText(msg.String())
			objs[1].(*widget.Button).OnTapped = func() {
				// accepted
				contacts.Add(msg.From())
				requests.Remove(msg)
				(*(&list)).Refresh()
			}
			objs[2].(*widget.Button).OnTapped = func() {
				// rejected
				requests.Remove(msg)
				(*(&list)).Refresh()
			}
		},
	)
	list.Show()
	return list
}

package views

import (
	"fyne.io/fyne"
	"fyne.io/fyne/container"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/theme"
	"fyne.io/fyne/widget"
	"ipmail/ipmail"
)

func MakeContent(messages ipmail.MessageList) fyne.CanvasObject {
	icon := widget.NewIcon(nil)
	label := widget.NewLabel("Select An Item From The List")
	hbox := container.NewHBox(icon, label)

	list := widget.NewList(
		func() int {
			return messages.Len()
		},
		func() fyne.CanvasObject {
			return container.NewHBox(widget.NewIcon(theme.MailComposeIcon()),
				widget.NewLabel("Template Object"))
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			item.(*fyne.Container).Objects[1].(*widget.Label).SetText(messages.FromIndex(id).String())
		},
	)
	list.OnSelected = func(id widget.ListItemID) {
		label.SetText(string(messages.FromIndex(id).Data()))
		icon.SetResource(theme.DocumentIcon())
	}
	list.OnUnselected = func(id widget.ListItemID) {
		label.SetText("Select An Item From The List")
		icon.SetResource(nil)
	}
	return container.NewHSplit(list, fyne.NewContainerWithLayout(layout.NewCenterLayout(), hbox))
}

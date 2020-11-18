package views

import (
	"fyne.io/fyne"
	"fyne.io/fyne/widget"
)

type SelectFunction func()

func MakeNav(lists map[string]SelectFunction) fyne.CanvasObject {
	tree := &widget.Tree{}
	keys := make([]string, 0, len(lists))
	for k := range lists {
		keys = append(keys, k)
	}
	tree.ChildUIDs = func(id string) []string {
		if tree.IsBranch(id) {
			return keys
		}
		return make([]string, 0)
	}
	tree.IsBranch = func(id string) bool {
		return len(id) == 0
	}
	tree.CreateNode = func(branch bool) fyne.CanvasObject {
		return widget.NewLabel("Template Object")
	}
	tree.UpdateNode = func(uid string, branch bool, node fyne.CanvasObject) {
		node.(*widget.Label).SetText(uid)
	}
	tree.OnSelected = func(id string) {
		lists[id]()
	}
	tree.ExtendBaseWidget(tree)

	return tree
}

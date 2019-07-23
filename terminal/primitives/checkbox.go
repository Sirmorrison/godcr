package primitives

import (
	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
)

type Checkbox struct {
	*tview.Checkbox
	text string
}

func NewCheckbox(text string) *Checkbox {
	return &Checkbox{
		Checkbox: tview.NewCheckbox().SetLabel(text).SetLabelColor(tcell.ColorWhite),
	}
}

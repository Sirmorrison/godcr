package pages

import (
	"github.com/gdamore/tcell"
	"github.com/raedahgroup/godcr/terminal/helpers"
	"github.com/raedahgroup/godcr/terminal/primitives"
	"github.com/rivo/tview"
)

func settingsPage() tview.Primitive {
	body := tview.NewFlex().SetDirection(tview.FlexRow)

	body.AddItem(primitives.NewLeftAlignedTextView("Settings"), 2, 1, false)

	body.AddItem(primitives.NewLeftAlignedTextView("Settings page coming soon").SetTextColor(helpers.DecredGreenColor), 0, 1, false)

	body.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			commonPageData.clearAllPageContent()
			return nil
		}

		return event
	})

	commonPageData.app.SetFocus(body)
	return body
}

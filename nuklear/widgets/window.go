package widgets

import (
	"image"

	"fmt"
	"github.com/aarzilli/nucular"
	"github.com/aarzilli/nucular/rect"
	"github.com/raedahgroup/godcr/nuklear/helpers"
	"golang.org/x/image/font"
	"image/color"
)

const defaultPageContentPadding = 10

type Window struct {
	*nucular.Window
}

func GroupWindow(uniqueWindowTitle string, parentWindow *nucular.Window, flags nucular.WindowFlags, windowReady func(*Window)) {
	if nw := parentWindow.GroupBegin(uniqueWindowTitle, flags); nw != nil {
		window := &Window{nw}
		windowReady(window)
		window.DoneAddingWidgets()
	}
}

func NoScrollGroupWindow(uniqueWindowTitle string, parentWindow *nucular.Window, windowReady func(*Window)) {
	GroupWindow(uniqueWindowTitle, parentWindow, nucular.WindowNoScrollbar, windowReady)
}

func ScrollableGroupWindow(uniqueWindowTitle string, parentWindow *nucular.Window, windowReady func(*Window)) {
	GroupWindow(uniqueWindowTitle, parentWindow, 0, windowReady)
}

func PageContentWindowDefaultPadding(pageTitle string, parentWindow *nucular.Window, windowReady func(contentWindow *Window)) {
	PageContentWindowWithPadding(pageTitle, parentWindow, defaultPageContentPadding, 0, windowReady)
}

func PageContentWindowWithPadding(pageTitle string, parentWindow *nucular.Window, xPadding, yPadding int, windowReady func(contentWindow *Window)) {
	NoScrollGroupWindow(pageTitle+"-page", parentWindow, func(pageWindow *Window) {
		pageWindow.Master().Style().GroupWindow.Padding = image.Point{X: 0, Y: 0}

		pageTitleHeight := pageWindow.SingleLineLabelHeight() + (defaultPageContentPadding * 2)
		pageTitleArea := pageWindow.Row(pageTitleHeight).SpaceBegin(1)
		pageWindow.LayoutSpacePushScaled(rect.Rect{
			X: defaultPageContentPadding,
			Y: 0,
			W: pageTitleArea.W,
			H: pageTitleHeight,
		})
		pageWindow.SetFont(helpers.PageHeaderFont)
		pageWindow.Label(pageTitle, LeftCenterAlign)

		pageWindow.PageContentWindow(pageTitle+"-page-content", xPadding, yPadding, windowReady)
	})
}

func (window *Window) PageContentWindow(uniqueWindowTitle string, xPadding, yPadding int, windowReady func(*Window)) {
	// create a rect for this page content window to prevent helpers from spilling into other windows
	pageContentArea := window.Row(0).SpaceBegin(1)
	window.LayoutSpacePushScaled(rect.Rect{
		X: 0,
		Y: 0,
		W: pageContentArea.W,
		H: pageContentArea.H,
	})

	window.Master().Style().Font = helpers.PageContentFont
	window.Master().Style().GroupWindow.Padding = image.Point{X: xPadding, Y: yPadding}

	// create group window
	ScrollableGroupWindow(uniqueWindowTitle, window.Window, windowReady)
}

func (window *Window) Font() font.Face {
	return window.Master().Style().Font
}

func (window *Window) SetFont(font font.Face) {
	window.Master().Style().Font = font
}

func (window *Window) UseFontAndResetToPrevious(font fontFace, fontReadyForUse func()) {
	currentFont := window.Master().Style().Font
	if currentFont != font {
		window.SetFont(font)
		defer window.SetFont(currentFont)
	}

	fontReadyForUse()
}

func (window *Window) DisplayErrorMessage(message string, err error) {
	window.DisplayMessage(fmt.Sprintf("%s: %s", message, err.Error()), helpers.DecredOrangeColor)
}

func (window *Window) DisplayMessage(message string, color color.RGBA) {
	window.AddWrappedLabelWithColor(message, LeftCenterAlign, color)
}

func (window *Window) DisplayIsLoadingMessage() {
	window.AddColoredLabel("Fetching data...", helpers.DecredOrangeColor, LeftCenterAlign)
}

func (window *Window) DoneAddingWidgets() {
	window.GroupEnd()
}

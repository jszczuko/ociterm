package gui

import (
	"fmt"
	"log"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type GuiController struct {
	pages       *tview.Pages
	topPanel    *guiTopPanel
	application *tview.Application
}

func NewGuiController(app *tview.Application) *GuiController {
	return &GuiController{
		pages:       tview.NewPages().AddPage("main", tview.NewTable(), false, true),
		topPanel:    NewTopPanel(),
		application: app,
	}
}

const (
	n_loading = "loading"
	n_main    = "main"
)

func (controller *GuiController) SetLoading() {
	if !controller.pages.HasPage(n_loading) {
		form := tview.NewTextView()
		form.SetTitle(n_loading)
		form.SetBorder(true)
		form.SetText("Loading ....")
		form.SetTextAlign(tview.AlignCenter)

		g := tview.NewGrid().
			SetColumns(0, 20, 0).
			SetRows(0, 3, 0).
			AddItem(form, 1, 1, 1, 1, 0, 0, true)
		controller.pages.AddAndSwitchToPage(n_loading, g, true).ShowPage(n_main)
	}
}

func (controller *GuiController) RemoveLoading() {
	if controller.pages.HasPage(n_loading) {
		controller.pages.RemovePage(n_loading).ShowPage(n_main)
	}
}

func (controller *GuiController) RefreshGUI() {
	controller.application.Draw()
}

func (controller *GuiController) GetGUITopPanel() *guiTopPanel {
	return controller.topPanel
}

func (controller *GuiController) GetGUIPages() *tview.Pages {
	return controller.pages
}

func (controller *GuiController) GetSetFocusFunc(primitive tview.Primitive) func() {
	return func() {
		controller.SetFocus(primitive)
	}
}

func (controller *GuiController) SetFocus(primitive tview.Primitive) {
	controller.application.SetFocus(primitive)
}

func (controller *GuiController) GetBackFocusFunc() func() {
	return func() {
		controller.SetFocus(controller.topPanel.profileInput)
	}
}

func WrapButton(button *tview.Button) *tview.Grid {
	return tview.NewGrid().
		SetColumns(0, 10, 0).
		SetRows(0, 1, 0).
		AddItem(button, 1, 1, 1, 1, 0, 0, false)
}

func (controller *GuiController) BindDefaultDoneFunc(onTab, onBackTab, onEsc tview.Primitive) func(key tcell.Key) {
	return func(key tcell.Key) {
		if tcell.KeyTAB == key && onTab != nil {
			controller.GetSetFocusFunc(onTab)()
		}
		if tcell.KeyEscape == key {
			if onEsc == nil {
				controller.GetBackFocusFunc()()
			} else {
				controller.GetSetFocusFunc(onEsc)()
			}
		}
		if tcell.KeyBacktab == key && onBackTab != nil {
			controller.GetSetFocusFunc(onBackTab)()
		}
	}
}

func (controller *GuiController) AddPage(name string, item tview.Primitive, resize bool) {
	controller.pages.AddAndSwitchToPage(name, item, resize).ShowPage(n_main)
}

func (controller *GuiController) RemovePage(removePage string, showPage string) error {
	if controller.pages.HasPage(removePage) {
		if controller.pages.HasPage(showPage) {
			controller.pages.RemovePage(removePage).ShowPage(showPage)
			return nil
		} else {
			return fmt.Errorf("page %v does not exists in pages", showPage)
		}
	} else {
		return fmt.Errorf("page %v does not exists in pages", removePage)
	}
}

func (controller *GuiController) LogError(message string, modal bool) {
	log.Print("ERROR " + message)
	if modal {
		modalName := "ModalErrorWindow"
		log.Print(modalName)
		modal := tview.NewModal().SetText(message).
			AddButtons([]string{"OK"}).
			SetDoneFunc(func(buttonIndex int, buttonLabel string) {
				controller.RemovePage(modalName, n_main)
			})
		controller.AddPage(modalName, modal, false)
	}
}

// Interface defining gui panel.
// Designed for lists of entities.
//
// - method Show(pages *tview.Pages) adds page with name defined by GetPanelName();
//
// - method GetPanelName() string returns unique name of the panel;
//
// - method Remove(pages *tview.Pages) removes panel from pages.
type GUIPanel interface {
	Show(pages *tview.Pages)
	Remove(pages *tview.Pages)
	GetPanelName() string
	GetInfo() string
}

func getFreeTagTable(tags map[string]string) *tview.Table {
	table := tview.NewTable()

	table.SetCell(0, 0, tview.NewTableCell("TAG").SetAlign(tview.AlignCenter).SetSelectable(false))
	table.SetCell(0, 1, tview.NewTableCell("VALUE").SetAlign(tview.AlignCenter).SetSelectable(false))
	row := 1
	for k, v := range tags {
		table.SetCell(row, 0, tview.NewTableCell(k))
		table.SetCell(row, 1, tview.NewTableCell(v))
		row += 1
	}
	table.SetBorder(true).SetTitle("Free Tags")
	return table
}

func getDefinedTagTable(tags map[string]map[string]interface{}) *tview.Table {
	table := tview.NewTable()

	table.SetCell(0, 0, tview.NewTableCell("TAG").SetAlign(tview.AlignCenter).SetSelectable(false))
	table.SetCell(0, 1, tview.NewTableCell("VALUE").SetAlign(tview.AlignCenter).SetSelectable(false))
	row := 1
	for _, ts := range tags {
		for v, t := range ts {
			table.SetCell(row, 0, tview.NewTableCell(v))
			table.SetCell(row, 1, tview.NewTableCell(t.(string)).SetSelectable(false))
			row += 1
		}
	}
	table.SetBorder(true).SetTitle("Defined Tags")
	return table
}

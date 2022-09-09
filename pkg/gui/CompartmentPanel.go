package gui

import (
	"log"
	"sort"
	"strconv"

	"github.com/gdamore/tcell/v2"
	oci "github.com/jszczuko/ociterm/pkg/oci"
	"github.com/oracle/oci-go-sdk/v52/identity"
	"github.com/rivo/tview"
)

type compartmentsPage struct {
	page         *string
	compartments *[]identity.Compartment
	nextPage     *string
}

type compartmentsGUI struct {
	mainGrid            *tview.Grid
	limitInput          *tview.InputField
	accessLevelDropDown *tview.DropDown
	sortByDropDown      *tview.DropDown
	sortOrderDropDown   *tview.DropDown
	lifecycleDropDown   *tview.DropDown
	refreshButton       *tview.Button
	nextPageButton      *tview.Button
	previousPageButton  *tview.Button
	mainTable           *tview.Table
}

type CompartmentPanel struct {
	guiController     *GuiController
	ociController     *oci.OCIController
	gui               *compartmentsGUI
	compartmentsPages []compartmentsPage
	currentPageIdx    int
	tenancyId         string
	compartmentId     string
	limit             int
	accessLevel       map[string]identity.ListCompartmentsAccessLevelEnum
	sortBy            map[string]identity.ListCompartmentsSortByEnum
	sortOrder         map[string]identity.ListCompartmentsSortOrderEnum
	lifecycleState    map[string]identity.CompartmentLifecycleStateEnum
}

func NewCompartmentPanel(TenancyId string, CompartmentId string, OciController *oci.OCIController, GuiController *GuiController) *CompartmentPanel {
	var res CompartmentPanel = CompartmentPanel{
		guiController:     GuiController,
		ociController:     OciController,
		compartmentsPages: make([]compartmentsPage, 0),
		currentPageIdx:    -1,
		tenancyId:         TenancyId,
		compartmentId:     CompartmentId,
		limit:             25,
		accessLevel: map[string]identity.ListCompartmentsAccessLevelEnum{
			"ANY":        identity.ListCompartmentsAccessLevelAny,
			"ACCESSIBLE": identity.ListCompartmentsAccessLevelAccessible,
		},
		sortBy: map[string]identity.ListCompartmentsSortByEnum{
			"TIMECREATED": identity.ListCompartmentsSortByTimecreated,
			"NAME":        identity.ListCompartmentsSortByName,
		},
		sortOrder: map[string]identity.ListCompartmentsSortOrderEnum{
			"ASC":  identity.ListCompartmentsSortOrderAsc,
			"DESC": identity.ListCompartmentsSortOrderDesc,
		},
		lifecycleState: map[string]identity.CompartmentLifecycleStateEnum{
			"ALL":      "",
			"CREATING": identity.CompartmentLifecycleStateCreating,
			"ACTIVE":   identity.CompartmentLifecycleStateActive,
			"INACTIVE": identity.CompartmentLifecycleStateInactive,
			"DELETING": identity.CompartmentLifecycleStateDeleting,
			"DELETED":  identity.CompartmentLifecycleStateDeleted,
		},
		gui: newCompartmentsGUI(),
	}
	res.createGUI()
	return &res
}

func NewCompartmentAsGUIPanel(TenancyId string, CompartmentId string, OciController *oci.OCIController, GuiController *GuiController) *GUIPanel {
	var inter interface{}
	var gui GUIPanel
	inter = NewCompartmentPanel(TenancyId, CompartmentId, OciController, GuiController)
	gui = inter.(GUIPanel)
	return &gui
}

func newCompartmentsGUI() *compartmentsGUI {
	var res compartmentsGUI = compartmentsGUI{
		mainGrid:            tview.NewGrid(),
		limitInput:          tview.NewInputField(),
		accessLevelDropDown: tview.NewDropDown(),
		sortByDropDown:      tview.NewDropDown(),
		sortOrderDropDown:   tview.NewDropDown(),
		lifecycleDropDown:   tview.NewDropDown(),
		refreshButton:       tview.NewButton("Refresh"),
		nextPageButton:      tview.NewButton("Page >>>"),
		previousPageButton:  tview.NewButton("<<< Page"),
		mainTable:           tview.NewTable(),
	}

	return &res
}

func (panel *CompartmentPanel) createGUI() {
	panel.gui.mainGrid.SetColumns(0, 20, 20, 20, 20, 20, 20, 20, 20, 0)
	panel.gui.mainGrid.SetRows(0, 3, 30)
	panel.gui.mainGrid.AddItem(WrapButton(panel.gui.previousPageButton), 1, 1, 1, 1, 0, 0, false)
	panel.gui.accessLevelDropDown.SetBorder(true).SetTitle("Access Level")
	panel.gui.mainGrid.AddItem(panel.gui.accessLevelDropDown, 1, 2, 1, 1, 0, 0, false)
	panel.gui.lifecycleDropDown.SetBorder(true).SetTitle("Lifecycle")
	panel.gui.mainGrid.AddItem(panel.gui.lifecycleDropDown, 1, 3, 1, 1, 0, 0, false)
	panel.gui.sortByDropDown.SetBorder(true).SetTitle("Sort By")
	panel.gui.mainGrid.AddItem(panel.gui.sortByDropDown, 1, 4, 1, 1, 0, 0, false)
	panel.gui.sortOrderDropDown.SetBorder(true).SetTitle("Sort Order")
	panel.gui.mainGrid.AddItem(panel.gui.sortOrderDropDown, 1, 5, 1, 1, 0, 0, false)
	panel.gui.limitInput.SetBorder(true).SetTitle("Limit")
	panel.gui.mainGrid.AddItem(panel.gui.limitInput, 1, 6, 1, 1, 0, 0, false)
	panel.gui.mainGrid.AddItem(WrapButton(panel.gui.refreshButton), 1, 7, 1, 1, 0, 0, false)
	panel.gui.mainGrid.AddItem(WrapButton(panel.gui.nextPageButton), 1, 8, 1, 1, 0, 0, false)

	panel.gui.mainTable.SetBorder(true).SetTitle("Compartments Table")
	panel.gui.mainGrid.AddItem(panel.gui.mainTable, 2, 0, 1, 10, 0, 0, false)

	panel.fillInitData()
	panel.makeKeyBindings()
}

func (panel *CompartmentPanel) fillInitData() {
	txt := make([]string, 0)
	for key := range panel.accessLevel {
		txt = append(txt, key)
	}
	sort.Strings(txt)
	panel.gui.accessLevelDropDown.SetOptions(txt, nil)
	panel.gui.accessLevelDropDown.SetCurrentOption(0)

	txt = make([]string, 0)
	for key := range panel.lifecycleState {
		txt = append(txt, key)
	}
	sort.Strings(txt)
	panel.gui.lifecycleDropDown.SetOptions(txt, nil)
	panel.gui.lifecycleDropDown.SetCurrentOption(0)

	txt = make([]string, 0)
	for key := range panel.sortBy {
		txt = append(txt, key)
	}
	sort.Strings(txt)
	panel.gui.sortByDropDown.SetOptions(txt, nil)
	panel.gui.sortByDropDown.SetCurrentOption(0)

	txt = make([]string, 0)
	for key := range panel.sortOrder {
		txt = append(txt, key)
	}
	sort.Strings(txt)
	panel.gui.sortOrderDropDown.SetOptions(txt, nil)
	panel.gui.sortOrderDropDown.SetCurrentOption(0)

	panel.gui.limitInput.SetText("25")
	panel.gui.limitInput.SetAcceptanceFunc(func(textToCheck string, lastChar rune) bool {
		val, err := strconv.Atoi(textToCheck)
		if err != nil {
			return false
		}
		if val > 100 || val < 1 {
			return false
		}
		return true
	})
}

func (panel *CompartmentPanel) makeKeyBindings() {
	panel.gui.limitInput.SetDoneFunc(panel.guiController.BindDefaultDoneFunc(panel.gui.refreshButton, panel.gui.sortOrderDropDown, nil))
	panel.gui.accessLevelDropDown.SetDoneFunc(panel.guiController.BindDefaultDoneFunc(panel.gui.lifecycleDropDown, panel.gui.previousPageButton, nil))
	panel.gui.sortByDropDown.SetDoneFunc(panel.guiController.BindDefaultDoneFunc(panel.gui.sortOrderDropDown, panel.gui.lifecycleDropDown, nil))
	panel.gui.sortOrderDropDown.SetDoneFunc(panel.guiController.BindDefaultDoneFunc(panel.gui.limitInput, panel.gui.sortByDropDown, nil))
	panel.gui.lifecycleDropDown.SetDoneFunc(panel.guiController.BindDefaultDoneFunc(panel.gui.sortByDropDown, panel.gui.accessLevelDropDown, nil))
	panel.gui.refreshButton.SetExitFunc(panel.guiController.BindDefaultDoneFunc(panel.gui.nextPageButton, panel.gui.limitInput, nil))
	panel.gui.nextPageButton.SetExitFunc(panel.guiController.BindDefaultDoneFunc(panel.gui.previousPageButton, panel.gui.refreshButton, nil))
	panel.gui.previousPageButton.SetExitFunc(panel.guiController.BindDefaultDoneFunc(panel.gui.accessLevelDropDown, panel.gui.nextPageButton, nil))

	panel.gui.previousPageButton.SetSelectedFunc(func() {
		panel.guiController.SetLoading()
		changed := false
		go func() {
			defer func() {
				panel.guiController.RemoveLoading()
				if changed {
					panel.guiController.SetFocus(panel.gui.mainTable)
				} else {
					panel.guiController.SetFocus(panel.gui.previousPageButton)
				}
				panel.guiController.RefreshGUI()
			}()
			if panel.currentPageIdx < 0 {
				return
			} else if panel.currentPageIdx > 0 {
				panel.currentPageIdx -= 1
				changed = true
				panel.refreshTable()
			}
		}()
	})
	panel.gui.nextPageButton.SetSelectedFunc(func() {
		panel.guiController.SetLoading()
		go func() {
			changed := false
			defer func() {
				panel.guiController.RemoveLoading()
				if changed {
					panel.guiController.SetFocus(panel.gui.mainTable)
				} else {
					panel.guiController.SetFocus(panel.gui.nextPageButton)
				}
				panel.guiController.RefreshGUI()
			}()

			if panel.currentPageIdx < 0 {
				return
			}
			// if page exists and was downloaded
			if panel.currentPageIdx+1 < len(panel.compartmentsPages) {
				panel.currentPageIdx += 1
				changed = true
				panel.refreshTable()
			} else
			// if page exists but not downloaded
			if *(panel.compartmentsPages[panel.currentPageIdx].nextPage) != "" {
				compartments, nextPage, err := panel.ociController.ListCompartments(
					panel.compartmentId,
					panel.getCurrnetLimit(),
					panel.getCurrentAccessLevel(),
					panel.getCurrentSortBy(),
					panel.getCurrentSortOrder(),
					panel.getCurrentLifecycleState(),
					*(panel.compartmentsPages[panel.currentPageIdx].nextPage),
				)
				if err != nil {
					log.Print(err.Error())
					return
				}
				p := ""
				panel.compartmentsPages = append(panel.compartmentsPages, compartmentsPage{
					page:         &p,
					compartments: &compartments,
					nextPage:     &nextPage,
				})
				panel.currentPageIdx = panel.currentPageIdx + 1
				changed = true
				panel.refreshTable()
			}
		}()

	})
	panel.gui.refreshButton.SetSelectedFunc(func() {
		panel.guiController.SetLoading()

		go func() {
			defer func() {
				panel.guiController.RemoveLoading()
				panel.guiController.SetFocus(panel.gui.mainTable)
				panel.guiController.RefreshGUI()
			}()
			// if data already exists refresh
			if panel.currentPageIdx > -1 {
				panel.currentPageIdx = -1
				panel.compartmentsPages = make([]compartmentsPage, 0)
				panel.gui.mainTable.Clear()
			}
			compartments, nextPage, err := panel.ociController.ListCompartments(
				panel.compartmentId,
				panel.getCurrnetLimit(),
				panel.getCurrentAccessLevel(),
				panel.getCurrentSortBy(),
				panel.getCurrentSortOrder(),
				panel.getCurrentLifecycleState(),
				"",
			)
			if err != nil {
				log.Fatal(err.Error())
				return
			}
			p := ""
			panel.compartmentsPages = append(panel.compartmentsPages, compartmentsPage{
				page:         &p,
				compartments: &compartments,
				nextPage:     &nextPage,
			})
			panel.currentPageIdx = 0

			panel.refreshTable()
		}()

	})
}

func (panel *CompartmentPanel) refreshTable() {
	panel.gui.mainTable.Clear()
	panel.gui.mainTable.SetSelectable(true, false).SetBorders(false)
	comps := panel.compartmentsPages[panel.currentPageIdx].compartments

	// header
	panel.gui.mainTable.SetCell(0, 0, tview.NewTableCell("NAME").SetAlign(tview.AlignCenter).SetSelectable(false))
	panel.gui.mainTable.SetCell(0, 1, tview.NewTableCell("CREATION TIME").SetAlign(tview.AlignCenter).SetSelectable(false))
	panel.gui.mainTable.SetCell(0, 2, tview.NewTableCell("LIFECYCLE STATE").SetAlign(tview.AlignCenter).SetSelectable(false))
	panel.gui.mainTable.SetCell(0, 3, tview.NewTableCell("OCID").SetAlign(tview.AlignCenter).SetSelectable(false))
	panel.gui.mainTable.SetCell(0, 4, tview.NewTableCell("ACCESS").SetAlign(tview.AlignCenter).SetSelectable(false))

	for row, val := range *comps {
		row += 1
		var cellcolor tcell.Color
		if row%2 == 0 {
			cellcolor = tcell.ColorWhite
		} else {
			cellcolor = tcell.ColorDarkGray
		}
		panel.gui.mainTable.SetCell(row, 0, tview.NewTableCell(*val.Name).SetAlign(tview.AlignLeft).SetTextColor(cellcolor))
		panel.gui.mainTable.SetCell(row, 1, tview.NewTableCell(val.TimeCreated.UTC().String()).SetAlign(tview.AlignCenter).SetTextColor(cellcolor))
		stateS, stateC := panel.lifecycleToString(val.LifecycleState)
		panel.gui.mainTable.SetCell(row, 2, tview.NewTableCell(stateS).SetAlign(tview.AlignCenter).SetTextColor(stateC))
		panel.gui.mainTable.SetCell(row, 3, tview.NewTableCell(*val.Id).SetAlign(tview.AlignLeft).SetTextColor(cellcolor))

		accessS, accessC := panel.accessibleToString(val.IsAccessible)
		panel.gui.mainTable.SetCell(row, 4, tview.NewTableCell(accessS).SetAlign(tview.AlignCenter).SetTextColor(accessC))
	}
	// focus on refresh button if esc was pressed
	panel.gui.mainTable.SetDoneFunc(func(key tcell.Key) {
		if tcell.KeyEscape == key {
			panel.guiController.SetFocus(panel.gui.refreshButton)
		}
	})
	// open comparment detail window
	panel.gui.mainTable.SetSelectedFunc(func(row, column int) {
		panelName := "CompartmentDetailPanel"
		comp := *(panel.compartmentsPages[panel.currentPageIdx].compartments)
		c := comp[row-1]
		detail := NewCompartmentDetailPanel(&c)
		panel.guiController.SetFocus(detail.freeTagTable)
		detail.freeTagTable.SetDoneFunc(func(key tcell.Key) {
			if tcell.KeyTab == key {
				panel.guiController.SetFocus(detail.definedTagTable)
			}
			if tcell.KeyEscape == key {
				panel.guiController.RemovePage(panelName, n_main)
				panel.guiController.SetFocus(panel.gui.mainTable)
			}
		})
		detail.definedTagTable.SetDoneFunc(func(key tcell.Key) {
			if tcell.KeyTab == key {
				panel.guiController.SetFocus(detail.freeTagTable)
			}
			if tcell.KeyEscape == key {
				panel.guiController.RemovePage(panelName, n_main)
				panel.guiController.SetFocus(panel.gui.mainTable)
			}
		})
		panel.guiController.AddPage(panelName, detail.GetGUI(), true)
	})
}

func (panel *CompartmentPanel) accessibleToString(access *bool) (string, tcell.Color) {
	if access == nil {
		return "UNKNOWN", tcell.ColorYellow
	}
	if *access {
		return "ACCESSIBLE", tcell.ColorGreen
	} else {
		return "NOT ACCESSIBLE", tcell.ColorRed
	}
}

func (panel *CompartmentPanel) lifecycleToString(lc identity.CompartmentLifecycleStateEnum) (string, tcell.Color) {
	switch lc {
	case identity.CompartmentLifecycleStateActive:
		return string(lc), tcell.ColorGreen
	case identity.CompartmentLifecycleStateInactive:
		return string(lc), tcell.ColorGray
	case identity.CompartmentLifecycleStateCreating:
		return string(lc), tcell.ColorLightGreen
	case identity.CompartmentLifecycleStateDeleting:
		return string(lc), tcell.ColorRed
	case identity.CompartmentLifecycleStateDeleted:
		return string(lc), tcell.ColorDarkRed
	default:
		return "", tcell.ColorWhite
	}
}

func (panel *CompartmentPanel) getCurrentLifecycleState() identity.CompartmentLifecycleStateEnum {
	_, val := panel.gui.lifecycleDropDown.GetCurrentOption()
	return panel.lifecycleState[val]
}

func (panel *CompartmentPanel) getCurrentSortOrder() identity.ListCompartmentsSortOrderEnum {
	_, val := panel.gui.sortOrderDropDown.GetCurrentOption()
	return panel.sortOrder[val]
}

func (panel *CompartmentPanel) getCurrentSortBy() identity.ListCompartmentsSortByEnum {
	_, val := panel.gui.sortByDropDown.GetCurrentOption()
	return panel.sortBy[val]
}

func (panel *CompartmentPanel) getCurrentAccessLevel() identity.ListCompartmentsAccessLevelEnum {
	_, val := panel.gui.accessLevelDropDown.GetCurrentOption()
	return panel.accessLevel[val]
}

func (panel *CompartmentPanel) getCurrnetLimit() int {
	currentLimit, err := strconv.Atoi(panel.gui.limitInput.GetText())
	if err != nil {
		currentLimit = 25
	}
	return currentLimit
}

func (panel *CompartmentPanel) GetPanelName() string {
	return "compartments"
}

func (panel *CompartmentPanel) Show(pages *tview.Pages) {
	if !pages.HasPage(panel.GetPanelName()) {
		pages.AddAndSwitchToPage(panel.GetPanelName(), panel.gui.mainGrid, true)
		panel.guiController.GetSetFocusFunc(panel.gui.refreshButton)()
	}
}

func (panel *CompartmentPanel) Remove(pages *tview.Pages) {
	if pages.HasPage(panel.GetPanelName()) {
		pages.RemovePage(panel.GetPanelName())
	}
}

func (panel *CompartmentPanel) GetInfo() string {
	return "[red]Enter:[white] Details [red]Esc:[white] Exit"
}

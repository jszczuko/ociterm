package gui

import (
	"log"
	"sort"
	"strconv"
	"sync"

	"github.com/gdamore/tcell/v2"
	oci "github.com/jszczuko/ociterm/pkg/oci"
	"github.com/oracle/oci-go-sdk/v52/core"
	"github.com/rivo/tview"
)

type instancesPage struct {
	page      *string
	instances *[]core.Instance
	nextPage  *string
}

type instancesGUI struct {
	mainGrid           *tview.Grid
	limitInput         *tview.InputField
	sortByDropDown     *tview.DropDown
	sortOrderDropDown  *tview.DropDown
	lifecycleDropDown  *tview.DropDown
	refreshButton      *tview.Button
	nextPageButton     *tview.Button
	previousPageButton *tview.Button
	mainTable          *tview.Table
}

type InstancesPanel struct {
	guiController      *GuiController
	ociController      *oci.OCIController
	gui                *instancesGUI
	instancesPages     []instancesPage
	instancesPagesLock sync.RWMutex
	currentPageIdx     int
	tenancyId          string
	compartmentId      string
	limit              int
	sortBy             map[string]core.ListInstancesSortByEnum
	sortOrder          map[string]core.ListInstancesSortOrderEnum
	lifecycleState     map[string]core.InstanceLifecycleStateEnum
}

func NewInstancesPanel(TenancyId string, CompartmentId string, OciController *oci.OCIController, GuiController *GuiController) *InstancesPanel {
	res := InstancesPanel{
		guiController:  GuiController,
		ociController:  OciController,
		compartmentId:  CompartmentId,
		tenancyId:      TenancyId,
		instancesPages: make([]instancesPage, 0),
		limit:          25,
		currentPageIdx: -1,
		sortBy: map[string]core.ListInstancesSortByEnum{
			"NAME":   core.ListInstancesSortByDisplayname,
			"CREATE": core.ListInstancesSortByTimecreated,
		},
		sortOrder: map[string]core.ListInstancesSortOrderEnum{
			"ASC":  core.ListInstancesSortOrderAsc,
			"DESC": core.ListInstancesSortOrderDesc,
		},
		lifecycleState: map[string]core.InstanceLifecycleStateEnum{
			"ALL":          "",
			"CREATING":     core.InstanceLifecycleStateCreatingImage,
			"MOVING":       core.InstanceLifecycleStateMoving,
			"PROVISIONING": core.InstanceLifecycleStateProvisioning,
			"RUNNING":      core.InstanceLifecycleStateRunning,
			"STARTING":     core.InstanceLifecycleStateStarting,
			"STOPPED":      core.InstanceLifecycleStateStopped,
			"STOPPING":     core.InstanceLifecycleStateStopping,
			"TERMINAGED":   core.InstanceLifecycleStateTerminated,
			"TERMINATING":  core.InstanceLifecycleStateTerminating,
		},
		gui: newInstancesGUI(),
	}
	res.createGUI()
	return &res
}

func newInstancesGUI() *instancesGUI {
	res := instancesGUI{
		mainGrid:           tview.NewGrid(),
		limitInput:         tview.NewInputField(),
		sortByDropDown:     tview.NewDropDown(),
		sortOrderDropDown:  tview.NewDropDown(),
		lifecycleDropDown:  tview.NewDropDown(),
		refreshButton:      tview.NewButton("Refresh"),
		nextPageButton:     tview.NewButton("Page >>>"),
		previousPageButton: tview.NewButton("<<< Page"),
		mainTable:          tview.NewTable(),
	}
	return &res
}

func NewInstancesAsGUIPanel(TenancyId string, CompartmentId string, OciController *oci.OCIController, GuiController *GuiController) *GUIPanel {
	var inter interface{}
	var gui GUIPanel
	inter = NewInstancesPanel(TenancyId, CompartmentId, OciController, GuiController)
	gui = inter.(GUIPanel)
	return &gui
}

func (panel *InstancesPanel) refreshInstance(instance *core.Instance) {
	for pageIdx, page := range panel.instancesPages {
		for instIdx, inst := range *page.instances {
			if *(instance.Id) == *(inst.Id) {
				func() {
					panel.instancesPagesLock.Lock()
					defer panel.instancesPagesLock.Unlock()
					(*panel.instancesPages[pageIdx].instances)[instIdx] = *instance
					panel.refreshTable()
				}()
			}
		}
	}
}

func (panel *InstancesPanel) createGUI() {
	panel.gui.mainGrid.SetColumns(0, 20, 20, 20, 20, 20, 20, 20, 0)
	panel.gui.mainGrid.SetRows(0, 3, 30)
	panel.gui.mainGrid.AddItem(WrapButton(panel.gui.previousPageButton), 1, 1, 1, 1, 0, 0, false)
	panel.gui.lifecycleDropDown.SetBorder(true).SetTitle("Lifecycle")
	panel.gui.mainGrid.AddItem(panel.gui.lifecycleDropDown, 1, 2, 1, 1, 0, 0, false)
	panel.gui.sortByDropDown.SetBorder(true).SetTitle("Sort By")
	panel.gui.mainGrid.AddItem(panel.gui.sortByDropDown, 1, 3, 1, 1, 0, 0, false)
	panel.gui.sortOrderDropDown.SetBorder(true).SetTitle("Sort Order")
	panel.gui.mainGrid.AddItem(panel.gui.sortOrderDropDown, 1, 4, 1, 1, 0, 0, false)
	panel.gui.limitInput.SetBorder(true).SetTitle("Limit")
	panel.gui.mainGrid.AddItem(panel.gui.limitInput, 1, 5, 1, 1, 0, 0, false)
	panel.gui.mainGrid.AddItem(WrapButton(panel.gui.refreshButton), 1, 6, 1, 1, 0, 0, false)
	panel.gui.mainGrid.AddItem(WrapButton(panel.gui.nextPageButton), 1, 7, 1, 1, 0, 0, false)

	panel.gui.mainTable.SetBorder(true).SetTitle("Compute Instances Table")
	panel.gui.mainGrid.AddItem(panel.gui.mainTable, 2, 0, 1, 9, 0, 0, false)

	panel.fillInitData()
	panel.makeKeyBindings()
}

func (panel *InstancesPanel) fillInitData() {
	txt := make([]string, 0)
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

func (panel *InstancesPanel) makeKeyBindings() {
	panel.gui.previousPageButton.SetExitFunc(panel.guiController.BindDefaultDoneFunc(panel.gui.lifecycleDropDown, panel.gui.nextPageButton, nil))
	panel.gui.lifecycleDropDown.SetDoneFunc(panel.guiController.BindDefaultDoneFunc(panel.gui.sortByDropDown, panel.gui.previousPageButton, nil))
	panel.gui.sortByDropDown.SetDoneFunc(panel.guiController.BindDefaultDoneFunc(panel.gui.sortOrderDropDown, panel.gui.lifecycleDropDown, nil))
	panel.gui.sortOrderDropDown.SetDoneFunc(panel.guiController.BindDefaultDoneFunc(panel.gui.limitInput, panel.gui.sortByDropDown, nil))
	panel.gui.limitInput.SetDoneFunc(panel.guiController.BindDefaultDoneFunc(panel.gui.refreshButton, panel.gui.sortOrderDropDown, nil))
	panel.gui.refreshButton.SetExitFunc(panel.guiController.BindDefaultDoneFunc(panel.gui.nextPageButton, panel.gui.limitInput, nil))
	panel.gui.nextPageButton.SetExitFunc(panel.guiController.BindDefaultDoneFunc(panel.gui.previousPageButton, panel.gui.refreshButton, nil))

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
			if panel.currentPageIdx+1 < len(panel.instancesPages) {
				panel.currentPageIdx += 1
				changed = true
				panel.refreshTable()
			} else
			// if page exists but not downloaded
			if *(panel.instancesPages[panel.currentPageIdx].nextPage) != "" {
				instances, nextPage, err := panel.ociController.ListInstances(
					panel.compartmentId,
					panel.getCurrnetLimit(),
					panel.getCurrentSortBy(),
					panel.getCurrentSortOrder(),
					panel.getCurrentLifecycleState(),
					*(panel.instancesPages[panel.currentPageIdx].nextPage),
				)
				if err != nil {
					log.Print(err.Error())
					return
				}
				p := ""
				panel.instancesPages = append(panel.instancesPages, instancesPage{
					page:      &p,
					instances: &instances,
					nextPage:  &nextPage,
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
			panel.instancesPagesLock.Lock()
			defer func() {
				panel.instancesPagesLock.Unlock()
				panel.guiController.RemoveLoading()
				panel.guiController.SetFocus(panel.gui.mainTable)
				panel.guiController.RefreshGUI()
			}()
			// if data already exists refresh
			if panel.currentPageIdx > -1 {
				panel.currentPageIdx = -1
				panel.instancesPages = make([]instancesPage, 0)
				panel.gui.mainTable.Clear()
			}
			instances, nextPage, err := panel.ociController.ListInstances(
				panel.compartmentId,
				panel.getCurrnetLimit(),
				panel.getCurrentSortBy(),
				panel.getCurrentSortOrder(),
				panel.getCurrentLifecycleState(),
				"",
			)
			if err != nil {
				log.Print(err.Error())
				return
			}
			p := ""
			panel.instancesPages = append(panel.instancesPages, instancesPage{
				page:      &p,
				instances: &instances,
				nextPage:  &nextPage,
			})
			panel.currentPageIdx = 0

			panel.refreshTable()
		}()

	})
}

func (panel *InstancesPanel) refreshTable() {
	panel.gui.mainTable.Clear()
	panel.gui.mainTable.SetSelectable(true, false).SetBorders(false)
	instances := panel.instancesPages[panel.currentPageIdx].instances

	// header
	panel.gui.mainTable.SetCell(0, 0, tview.NewTableCell("NAME").SetAlign(tview.AlignCenter).SetSelectable(false))
	panel.gui.mainTable.SetCell(0, 1, tview.NewTableCell("CREATION TIME").SetAlign(tview.AlignCenter).SetSelectable(false))
	panel.gui.mainTable.SetCell(0, 2, tview.NewTableCell("LIFECYCLE STATE").SetAlign(tview.AlignCenter).SetSelectable(false))
	panel.gui.mainTable.SetCell(0, 3, tview.NewTableCell("OCID").SetAlign(tview.AlignCenter).SetSelectable(false))
	panel.gui.mainTable.SetCell(0, 4, tview.NewTableCell("FAULT DOMAIN").SetAlign(tview.AlignCenter).SetSelectable(false))
	panel.gui.mainTable.SetCell(0, 5, tview.NewTableCell("AVAILABILITY DOMAIN").SetAlign(tview.AlignCenter).SetSelectable(false))

	for row, val := range *instances {
		row += 1
		var cellcolor tcell.Color
		if row%2 == 0 {
			cellcolor = tcell.ColorWhite
		} else {
			cellcolor = tcell.ColorDarkGray
		}
		panel.gui.mainTable.SetCell(row, 0, tview.NewTableCell(*val.DisplayName).SetAlign(tview.AlignLeft).SetTextColor(cellcolor))
		panel.gui.mainTable.SetCell(row, 1, tview.NewTableCell(val.TimeCreated.UTC().String()).SetAlign(tview.AlignCenter).SetTextColor(cellcolor))
		stateS, stateC := panel.lifecycleToString(val.LifecycleState)
		panel.gui.mainTable.SetCell(row, 2, tview.NewTableCell(stateS).SetAlign(tview.AlignCenter).SetTextColor(stateC))
		panel.gui.mainTable.SetCell(row, 3, tview.NewTableCell(*val.Id).SetAlign(tview.AlignLeft).SetTextColor(cellcolor))
		panel.gui.mainTable.SetCell(row, 4, tview.NewTableCell(*val.FaultDomain).SetAlign(tview.AlignLeft).SetTextColor(cellcolor))
		panel.gui.mainTable.SetCell(row, 5, tview.NewTableCell(*val.AvailabilityDomain).SetAlign(tview.AlignLeft).SetTextColor(cellcolor))
	}
	// focus on refresh button if esc was pressed
	panel.gui.mainTable.SetDoneFunc(func(key tcell.Key) {
		if tcell.KeyEscape == key {
			panel.guiController.SetFocus(panel.gui.refreshButton)
		}
	})
	// Instance Action Panel
	panel.gui.mainTable.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		key := event.Key()
		if tcell.KeyRune == key && event.Rune() == 'a' {
			row, _ := panel.gui.mainTable.GetSelection()
			instances := *(panel.instancesPages[panel.currentPageIdx].instances)
			instance := instances[row-1]
			detail := NewInstanceActionPanel(&instance)
			panel.guiController.SetFocus(detail.actionSelect)
			detail.actionSelect.SetDoneFunc(func(key tcell.Key) {
				if tcell.KeyTab == key {
					panel.guiController.SetFocus(detail.executeButton)
				}
				if tcell.KeyEscape == key {
					panel.guiController.RemovePage(detail.GetPanelName(), n_main)
					panel.guiController.SetFocus(panel.gui.mainTable)
				}
			})
			detail.executeButton.SetExitFunc(func(key tcell.Key) {
				if tcell.KeyTab == key {
					panel.guiController.SetFocus(detail.actionSelect)
				}
				if tcell.KeyEscape == key {
					panel.guiController.RemovePage(detail.GetPanelName(), n_main)
					panel.guiController.SetFocus(panel.gui.mainTable)
				}
			})
			// Modal window to confirm instance action
			modalName := "ModalInstanceActionPanel"
			detail.executeButton.SetSelectedFunc(func() {

				modal := tview.NewModal().
					SetText("Do you want to execute action?").
					AddButtons([]string{"Execute", "Cancel"}).
					SetDoneFunc(func(buttonIndex int, buttonLabel string) {
						defer func() {
							panel.guiController.RemovePage(modalName, detail.GetPanelName())
							panel.guiController.RemovePage(detail.GetPanelName(), n_main)
							panel.guiController.SetFocus(panel.gui.mainTable)
						}()

						if buttonLabel == "Execute" {
							func() {
								panel.guiController.SetLoading()
								defer panel.guiController.RemoveLoading()
								ocid := detail.GetInstanceOCID()
								action := detail.GetSelectedAction()
								newInstance, err := panel.ociController.ExecuteInstanceAction(&ocid, action)
								if err != nil {
									log.Print("ERROR " + err.Error())
									return
								}
								panel.refreshInstance(newInstance)
							}()
						}
					})
				panel.guiController.AddPage(modalName, modal, false)
			})
			panel.guiController.AddPage(detail.GetPanelName(), detail.GetGUI(), true)
		}
		// r for refresh
		if tcell.KeyRune == key && event.Rune() == 'r' {
			row, _ := panel.gui.mainTable.GetSelection()
			instances := *(panel.instancesPages[panel.currentPageIdx].instances)
			instance := instances[row-1]
			go panel.RefreshOciIntance(*instance.Id)
		}

		// m for monitoring
		if tcell.KeyRune == key && event.Rune() == 'm' {
			// TODO
		}
		return event
	})
	// open instace detail window
	panel.gui.mainTable.SetSelectedFunc(func(row, column int) {
		panelName := "InstanceDetailPanel"
		instances := *(panel.instancesPages[panel.currentPageIdx].instances)
		instance := instances[row-1]
		detail := NewInstanceDetailPanel(&instance)
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

func (panel *InstancesPanel) lifecycleToString(li core.InstanceLifecycleStateEnum) (string, tcell.Color) {
	switch li {
	case core.InstanceLifecycleStateRunning:
		return string(li), tcell.ColorGreen
	case core.InstanceLifecycleStateStarting:
		return string(li), tcell.ColorLightGreen
	case core.InstanceLifecycleStateStopped:
		return string(li), tcell.ColorYellow
	case core.InstanceLifecycleStateStopping:
		return string(li), tcell.ColorLightYellow
	case core.InstanceLifecycleStateTerminated:
		return string(li), tcell.ColorGray
	case core.InstanceLifecycleStateTerminating:
		return string(li), tcell.ColorLightGray
	case core.InstanceLifecycleStateCreatingImage:
		return string(li), tcell.ColorLightSeaGreen
	case core.InstanceLifecycleStateMoving:
		return string(li), tcell.ColorMediumSeaGreen
	case core.InstanceLifecycleStateProvisioning:
		return string(li), tcell.ColorLawnGreen
	default:
		return "", tcell.ColorWhite
	}
}

func (panel *InstancesPanel) getCurrentLifecycleState() core.InstanceLifecycleStateEnum {
	_, val := panel.gui.lifecycleDropDown.GetCurrentOption()
	return panel.lifecycleState[val]
}

func (panel *InstancesPanel) getCurrentSortOrder() core.ListInstancesSortOrderEnum {
	_, val := panel.gui.sortOrderDropDown.GetCurrentOption()
	return panel.sortOrder[val]
}

func (panel *InstancesPanel) getCurrentSortBy() core.ListInstancesSortByEnum {
	_, val := panel.gui.sortByDropDown.GetCurrentOption()
	return panel.sortBy[val]
}

func (panel *InstancesPanel) getCurrnetLimit() int {
	currentLimit, err := strconv.Atoi(panel.gui.limitInput.GetText())
	if err != nil {
		currentLimit = 25
	}
	return currentLimit
}

func (panel *InstancesPanel) GetPanelName() string {
	return "instances"
}

func (panel *InstancesPanel) Show(pages *tview.Pages) {
	if !pages.HasPage(panel.GetPanelName()) {
		pages.AddAndSwitchToPage(panel.GetPanelName(), panel.gui.mainGrid, true)
		panel.guiController.GetSetFocusFunc(panel.gui.refreshButton)()
	}
}

func (panel *InstancesPanel) Remove(pages *tview.Pages) {
	if pages.HasPage(panel.GetPanelName()) {
		pages.RemovePage(panel.GetPanelName())
	}
}

func (panel *InstancesPanel) GetInfo() string {
	return "[red]Enter:[white] Details [red]Esc:[white] Exit [green]a:[white] Action [green]r:[white] Refresh [green]m:[white] Monitoring"
}

func (panel *InstancesPanel) RefreshOciIntance(OcidId string) {
	inst, err := panel.ociController.GetInstance(OcidId)
	if err == nil {
		panel.refreshInstance(inst)
	}
}

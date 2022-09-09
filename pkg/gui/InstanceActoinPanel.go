package gui

import (
	"sort"

	"github.com/oracle/oci-go-sdk/v52/core"
	"github.com/rivo/tview"
)

type InstanceActionPanel struct {
	grid          *tview.Grid
	instance      *core.Instance
	executeButton *tview.Button
	actionSelect  *tview.DropDown
	actionMap     map[string]core.InstanceActionActionEnum
}

func (panel *InstanceActionPanel) GetGUI() tview.Primitive {
	return panel.grid
}

func (panel *InstanceActionPanel) GetPanelName() string {
	return "InstanceActionPanel"
}

func (panel *InstanceActionPanel) GetSelectedAction() core.InstanceActionActionEnum {
	_, option := panel.actionSelect.GetCurrentOption()
	return panel.actionMap[option]
}

func (panel *InstanceActionPanel) GetInstanceOCID() string {
	return *panel.instance.Id
}

func NewInstanceActionPanel(instance *core.Instance) *InstanceActionPanel {
	res := InstanceActionPanel{
		grid:     tview.NewGrid(),
		instance: instance,
	}

	ocid := tview.NewInputField().SetLabel("OCID:").SetText(*instance.Id)
	name := tview.NewInputField().SetLabel("Name:").SetText(*instance.DisplayName)
	lifecycle := tview.NewInputField().SetLabel("Lifecycle:").SetText(string(instance.LifecycleState))
	res.actionSelect, res.actionMap = instanceActionToDropDown()
	res.executeButton = tview.NewButton("Execute")

	grid := tview.NewGrid()
	grid.SetColumns(50, 50)
	grid.SetRows(1, 1, 1)

	grid.AddItem(ocid, 0, 0, 1, 2, 0, 0, false)
	grid.AddItem(name, 1, 0, 1, 1, 0, 0, false)
	grid.AddItem(lifecycle, 1, 1, 1, 1, 0, 0, false)
	grid.AddItem(res.actionSelect, 2, 0, 1, 1, 0, 0, false)
	grid.AddItem(res.executeButton, 2, 1, 1, 1, 0, 0, false)

	grid.SetBorder(true).SetTitle("Instance Action")

	res.grid.SetColumns(0, 100, 0)
	res.grid.SetRows(0, 5, 0)
	res.grid.AddItem(grid, 1, 1, 1, 1, 0, 0, false)

	return &res
}

func instanceActionToDropDown() (*tview.DropDown, map[string]core.InstanceActionActionEnum) {
	res := tview.NewDropDown().SetLabel("Select Action:")
	actionsStr := make([]string, 0)
	actionMap := make(map[string]core.InstanceActionActionEnum)
	for _, act := range core.GetInstanceActionActionEnumValues() {
		actionsStr = append(actionsStr, string(act))
		actionMap[string(act)] = act
	}

	sort.Strings(actionsStr)
	res.SetOptions(actionsStr, nil)
	res.SetCurrentOption(0)
	return res, actionMap
}

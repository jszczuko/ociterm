package gui

import (
	"github.com/oracle/oci-go-sdk/v52/core"
	"github.com/rivo/tview"
)

type InstanceDetailPanel struct {
	grid            *tview.Grid
	instance        *core.Instance
	freeTagTable    *tview.Table
	definedTagTable *tview.Table
}

func (panel *InstanceDetailPanel) GetGUI() tview.Primitive {
	return panel.grid
}

func NewInstanceDetailPanel(instance *core.Instance) *InstanceDetailPanel {
	res := InstanceDetailPanel{
		grid:     tview.NewGrid(),
		instance: instance,
	}

	ocid := tview.NewInputField().SetLabel("OCID:").SetText(*instance.Id)
	compId := tview.NewInputField().SetLabel("Parent:").SetText(*instance.CompartmentId)
	ad := tview.NewInputField().SetLabel("AD:").SetText(*instance.AvailabilityDomain)
	region := tview.NewInputField().SetLabel("Region:").SetText(*instance.Region)
	name := tview.NewInputField().SetLabel("Name:").SetText(*instance.DisplayName)
	created := tview.NewInputField().SetLabel("Created:").SetText(instance.TimeCreated.UTC().String())
	fd := tview.NewInputField().SetLabel("FD:").SetText(*instance.FaultDomain)
	lifecycle := tview.NewInputField().SetLabel("Lifecycle:").SetText(string(instance.LifecycleState))

	grid := tview.NewGrid()
	grid.SetColumns(50, 50)
	grid.SetRows(1, 1, 1, 1, 1, 8)

	grid.AddItem(ocid, 0, 0, 1, 2, 0, 0, false)
	grid.AddItem(compId, 1, 0, 1, 2, 0, 0, false)
	grid.AddItem(name, 2, 0, 1, 1, 0, 0, false)
	grid.AddItem(created, 2, 1, 1, 1, 0, 0, false)
	grid.AddItem(ad, 3, 0, 1, 1, 0, 0, false)
	grid.AddItem(fd, 3, 1, 1, 1, 0, 0, false)
	grid.AddItem(region, 4, 0, 1, 1, 0, 0, false)
	grid.AddItem(lifecycle, 4, 1, 1, 1, 0, 0, false)

	res.freeTagTable = getFreeTagTable(instance.FreeformTags)
	res.definedTagTable = getDefinedTagTable(instance.DefinedTags)

	grid.AddItem(res.freeTagTable, 5, 0, 1, 1, 0, 0, true)
	grid.AddItem(res.definedTagTable, 5, 1, 1, 1, 0, 0, false)

	grid.SetBorder(true).SetTitle("Compartment Details")

	res.grid.SetColumns(0, 100, 0)
	res.grid.SetRows(0, 15, 0)
	res.grid.AddItem(grid, 1, 1, 1, 1, 0, 0, false)

	return &res
}

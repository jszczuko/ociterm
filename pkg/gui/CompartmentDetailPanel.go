package gui

import (
	"github.com/oracle/oci-go-sdk/v52/identity"
	"github.com/rivo/tview"
)

type CompartmentDetailPanel struct {
	grid            *tview.Grid
	compartment     *identity.Compartment
	freeTagTable    *tview.Table
	definedTagTable *tview.Table
}

func (panel *CompartmentDetailPanel) GetGUI() tview.Primitive {
	return panel.grid
}

func NewCompartmentDetailPanel(compartment *identity.Compartment) *CompartmentDetailPanel {
	res := CompartmentDetailPanel{
		grid:        tview.NewGrid(),
		compartment: compartment,
	}

	ocid := tview.NewInputField().SetLabel("OCID:").SetText(*compartment.Id)
	parent := tview.NewInputField().SetLabel("Parent OCID:").SetText(*compartment.CompartmentId)
	name := tview.NewInputField().SetLabel("Name:").SetText(*compartment.Name)
	created := tview.NewInputField().SetLabel("Created:").SetText(compartment.TimeCreated.UTC().String())
	description := tview.NewTextView()
	description.SetBorder(true)
	description.SetTitle("Description:")
	description.SetText(*compartment.Description)
	lifecycle := tview.NewInputField().SetLabel("Lifecycle:").SetText(string(compartment.LifecycleState))
	var accessStr string
	if compartment.IsAccessible == nil {
		accessStr = "UNKNOWN"
	} else if *compartment.IsAccessible {
		accessStr = "ACCESSABLE"
	} else {
		accessStr = "NOT ACCESSIBLE"
	}
	access := tview.NewInputField().SetLabel("Access:").SetText(accessStr)

	grid := tview.NewGrid()
	grid.SetColumns(50, 50)
	grid.SetRows(1, 1, 1, 1, 5, 8)

	grid.AddItem(ocid, 0, 0, 1, 2, 0, 0, false)
	grid.AddItem(parent, 1, 0, 1, 2, 0, 0, false)
	grid.AddItem(name, 2, 0, 1, 2, 0, 0, false)
	grid.AddItem(created, 2, 1, 1, 2, 0, 0, false)
	grid.AddItem(lifecycle, 3, 0, 1, 2, 0, 0, false)
	grid.AddItem(access, 3, 1, 1, 2, 0, 0, false)
	grid.AddItem(description, 4, 0, 1, 2, 0, 0, false)

	res.freeTagTable = getFreeTagTable(compartment.FreeformTags)
	res.definedTagTable = getDefinedTagTable(compartment.DefinedTags)

	grid.AddItem(res.freeTagTable, 5, 0, 1, 1, 0, 0, true)
	grid.AddItem(res.definedTagTable, 5, 1, 1, 2, 0, 0, false)

	grid.SetBorder(true).SetTitle("Compartment Details")

	res.grid.SetColumns(0, 100, 0)
	res.grid.SetRows(0, 19, 0)
	res.grid.AddItem(grid, 1, 1, 1, 1, 0, 0, false)

	return &res
}

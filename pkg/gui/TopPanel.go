package gui

import (
	"sync"

	"github.com/oracle/oci-go-sdk/v52/identity"
	"github.com/rivo/tview"
)

type guiTopPanel struct {
	profileInput         *tview.InputField
	regionsDropDown      *tview.DropDown
	compartmentsDropDown *tview.DropDown
	resourcesDropDown    *tview.DropDown
	refreshButton        *tview.Button

	regions        *[]identity.Region
	regionsMu      sync.Mutex
	compartments   *[]identity.Compartment
	compartmentsMu sync.Mutex

	defaultRegion string

	guiPrimitve *tview.Grid

	toBeRefreshed bool
}

func (panel *guiTopPanel) SetDefaultRegion(region string) {
	panel.defaultRegion = region
}

func (panel *guiTopPanel) GetProfileInput() *tview.InputField {
	return panel.profileInput
}

func (panel *guiTopPanel) GetRegionDropDown() *tview.DropDown {
	return panel.regionsDropDown
}

func (panel *guiTopPanel) GetCompartmentsDropDown() *tview.DropDown {
	return panel.compartmentsDropDown
}

func (panel *guiTopPanel) GetRefreshButton() *tview.Button {
	return panel.refreshButton
}

func (panel *guiTopPanel) GetResourcesDropDown() *tview.DropDown {
	return panel.resourcesDropDown
}

func (panel *guiTopPanel) UpdateRegions(regions *[]identity.Region) {
	panel.regionsMu.Lock()
	defer panel.regionsMu.Unlock()
	panel.regions = regions
	panel.toBeRefreshed = true
}

func (panel *guiTopPanel) UpdateCompartments(compartments *[]identity.Compartment) {
	panel.compartmentsMu.Lock()
	defer panel.compartmentsMu.Unlock()
	panel.compartments = compartments
	panel.toBeRefreshed = true
}

func (panel *guiTopPanel) IsToBeRefreshed() bool {
	return panel.toBeRefreshed
}

func (panel *guiTopPanel) UpdateGUI() {
	if panel.toBeRefreshed {
		panel.updateRegionsGUI()
		panel.updateCompartmentsGUI()
		panel.updateResourcesGUI()
		panel.toBeRefreshed = false
	}
}

func (panel *guiTopPanel) GetGuiPrimitive() *tview.Grid {
	return panel.guiPrimitve
}

func (panel *guiTopPanel) updateResourcesGUI() {
	panel.resourcesDropDown.SetOptions([]string{"compartments", "instances"}, nil)
}

func (panel *guiTopPanel) updateRegionsGUI() {
	if panel.regions != nil {
		selIdx := 0
		panel.compartmentsMu.Lock()
		defer panel.compartmentsMu.Unlock()
		var txt []string
		for idx, reg := range *panel.regions {
			txt = append(txt, *reg.Name)
			if panel.defaultRegion == *reg.Name || panel.defaultRegion == *reg.Key {
				selIdx = idx
			}
		}
		panel.regionsDropDown.SetOptions(txt, nil)
		panel.regionsDropDown.SetCurrentOption(selIdx)
	}
}

func (panel *guiTopPanel) updateCompartmentsGUI() {
	if panel.compartments != nil {
		panel.compartmentsMu.Lock()
		defer panel.compartmentsMu.Unlock()
		var txt []string
		txt = append(txt, "")
		for _, com := range *panel.compartments {
			txt = append(txt, *com.Name)
		}
		panel.compartmentsDropDown.SetOptions(txt, nil)
	}
}

func (panel *guiTopPanel) init() {
	panel.regions = nil
	panel.compartments = nil
	panel.toBeRefreshed = false
	panel.guiPrimitve = nil
	panel.defaultRegion = ""
}

func (panel *guiTopPanel) createLayout() {
	panel.profileInput = tview.NewInputField()
	panel.profileInput.SetBorder(true).SetTitle("Profile")
	panel.profileInput.SetPlaceholder("[DEFAULT]")
	panel.regionsDropDown = tview.NewDropDown()
	panel.regionsDropDown.SetBorder(true).SetTitle("Regions")
	panel.compartmentsDropDown = tview.NewDropDown()
	panel.compartmentsDropDown.SetBorder(true).SetTitle("Compartments")
	panel.resourcesDropDown = tview.NewDropDown()
	panel.resourcesDropDown.SetBorder(true).SetTitle("Resource")
	panel.refreshButton = tview.NewButton("Refresh")

	refresGrid := tview.NewGrid().
		SetColumns(0, 10, 0).
		SetRows(0, 1, 0).
		AddItem(panel.refreshButton, 1, 1, 1, 1, 0, 0, false)

	mainGrid := tview.NewGrid().
		SetColumns(20, 20, 20, 20, 20, 0).
		SetRows(0, 3, 0).
		AddItem(panel.profileInput, 1, 0, 1, 1, 0, 0, true).
		AddItem(panel.regionsDropDown, 1, 1, 1, 1, 0, 0, true).
		AddItem(panel.compartmentsDropDown, 1, 2, 1, 1, 0, 0, true).
		AddItem(panel.resourcesDropDown, 1, 3, 1, 1, 0, 0, true).
		AddItem(refresGrid, 1, 4, 1, 1, 0, 0, true)

	panel.guiPrimitve = mainGrid
}

func (panel *guiTopPanel) GetSelectedCompartment() *identity.Compartment {
	idx, _ := panel.compartmentsDropDown.GetCurrentOption()
	if idx > 0 && idx <= len(*panel.compartments) {
		return &(*panel.compartments)[idx-1] // As the first one is empty
	}
	return nil
}

func (panel *guiTopPanel) GetSelectedCompartmentId() string {
	cmp := panel.GetSelectedCompartment()
	if cmp == nil {
		return ""
	}
	return *cmp.Id
}

func (panel *guiTopPanel) GetSelectedRegion() *identity.Region {
	idx, _ := panel.regionsDropDown.GetCurrentOption()
	if idx > -1 && idx < len(*panel.regions) {
		return &(*panel.regions)[idx]
	}
	return nil
}

func (panel *guiTopPanel) GetSelectedRegionName() string {
	reg := panel.GetSelectedRegion()
	if reg == nil {
		return ""
	} else {
		return *reg.Name
	}
}

func (panel *guiTopPanel) GetSelectedResource() string {
	_, txt := panel.resourcesDropDown.GetCurrentOption()
	return txt
}

func (panel *guiTopPanel) GetSelectedProfile() string {
	return panel.profileInput.GetText()
}

func NewTopPanel() *guiTopPanel {
	panel := &guiTopPanel{}
	panel.init()
	panel.createLayout()
	return panel
}

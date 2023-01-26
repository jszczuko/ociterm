package main

import (
	"log"

	"github.com/gdamore/tcell/v2"
	"github.com/jszczuko/ociterm/pkg/gui"
	controller "github.com/jszczuko/ociterm/pkg/oci"
	"github.com/oracle/oci-go-sdk/v52/common"
	"github.com/rivo/tview"
)

type BasicConfiguration struct {
	TenancyId     string
	CompartmentId string
	Region        string
	Profile       string
}

type OciTerm struct {
	app           *tview.Application
	ociController *controller.OCIController
	guiController *gui.GuiController

	errorTextArea *tview.TextView
	mainPages     *tview.Pages
	mainView      *tview.Flex
	currentPanel  *gui.GUIPanel
}

func NewOciTerm() *OciTerm {
	res := &OciTerm{}
	res.init()
	return res
}

func (ociterm *OciTerm) GetBasicConfiguration() (conf *BasicConfiguration, err error) {
	configProvider := ociterm.ociController.GetConfigurationProvider()
	tenancyId, err := configProvider.TenancyOCID()
	if err != nil {
		return nil, err
	}
	compartmentId, err := ociterm.getCompartmentOCID(configProvider)
	if err != nil {
		return nil, err
	}
	regionName, err := ociterm.getRegionName(configProvider)
	if err != nil {
		return nil, err
	}
	profile := ociterm.guiController.GetGUITopPanel().GetProfileInput().GetText()

	return &BasicConfiguration{
		TenancyId:     tenancyId,
		CompartmentId: compartmentId,
		Region:        regionName,
		Profile:       profile,
	}, nil
}

func (ociterm *OciTerm) getRegionName(confProvider common.ConfigurationProvider) (string, error) {
	reg := ociterm.guiController.GetGUITopPanel().GetSelectedRegion()
	if reg != nil {
		return *reg.Name, nil
	}
	confReg, err := confProvider.Region()
	if err != nil {
		return "", err
	}
	return confReg, nil
}

func (ociterm *OciTerm) getCompartmentOCID(confProvider common.ConfigurationProvider) (string, error) {
	comp := ociterm.guiController.GetGUITopPanel().GetSelectedCompartment()
	if comp != nil {
		return *comp.Id, nil
	}
	tenancyId, err := confProvider.TenancyOCID()
	if err != nil {
		return "", err
	}
	return tenancyId, nil
}

func (ociterm *OciTerm) init() {
	ociterm.app = tview.NewApplication()
	ociterm.ociController = controller.NewOCIControllerDefault()
	ociterm.guiController = gui.NewGuiController(ociterm.app)
	ociterm.errorTextArea = tview.NewTextView()
	ociterm.errorTextArea.SetDynamicColors(true).SetBorder(true).SetTitle("INFO")
	ociterm.mainPages = tview.NewPages()
	ociterm.mainPages.SetBorder(true)
	ociterm.mainView = tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(
			ociterm.guiController.GetGUITopPanel().GetGuiPrimitive(), 0, 1, true).
		AddItem(ociterm.mainPages, 0, 8, false).
		AddItem(ociterm.errorTextArea, 0, 1, false)
	ociterm.guiController.GetGUIPages().AddAndSwitchToPage("main", ociterm.mainView, true)
	ociterm.currentPanel = nil

	ociterm.guiController.GetGUITopPanel().GetProfileInput().SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyTAB {
			ociterm.app.SetFocus(ociterm.guiController.GetGUITopPanel().GetRegionDropDown())
		} else if key == tcell.KeyEnter {

			ociterm.guiController.SetLoading()
			go func() {
				defer func() {
					ociterm.guiController.RemoveLoading()
					ociterm.app.QueueUpdateDraw(ociterm.guiController.GetGUITopPanel().UpdateGUI)
				}()

				err := ociterm.ociController.ReloadConfig("", ociterm.guiController.GetGUITopPanel().GetProfileInput().GetText())
				if err != nil {
					ociterm.guiController.LogError(err.Error(), true)
					return
				}

				regs, err := ociterm.ociController.ListRegions()
				if err != nil {
					ociterm.guiController.LogError(err.Error(), true)
					return
				} else {
					ociterm.guiController.GetGUITopPanel().UpdateRegions(&regs)
				}

				comps, err := ociterm.ociController.ListAllCompartments()
				if err != nil {
					ociterm.guiController.LogError(err.Error(), true)
					return
				} else {
					ociterm.guiController.GetGUITopPanel().UpdateCompartments(&comps)
				}

				ociterm.app.SetFocus(ociterm.guiController.GetGUITopPanel().GetProfileInput())
				conf, err := ociterm.GetBasicConfiguration()
				if err == nil {
					ociterm.guiController.GetGUITopPanel().SetDefaultRegion(conf.Region)
				}
			}()
		}
	})

	ociterm.guiController.GetGUITopPanel().GetRegionDropDown().SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyTAB {
			ociterm.app.SetFocus(ociterm.guiController.GetGUITopPanel().GetCompartmentsDropDown())
		}
	})

	ociterm.guiController.GetGUITopPanel().GetCompartmentsDropDown().SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyTAB {
			ociterm.app.SetFocus(ociterm.guiController.GetGUITopPanel().GetResourcesDropDown())
		}
	})

	ociterm.guiController.GetGUITopPanel().GetResourcesDropDown().SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyTAB {
			ociterm.app.SetFocus(ociterm.guiController.GetGUITopPanel().GetRefreshButton())
		}
	})

	ociterm.guiController.GetGUITopPanel().GetRefreshButton().SetExitFunc(func(key tcell.Key) {
		if tcell.KeyTAB == key {
			ociterm.app.SetFocus(ociterm.guiController.GetGUITopPanel().GetProfileInput())
		}
	})

	ociterm.guiController.GetGUITopPanel().GetRefreshButton().SetSelectedFunc(func() {
		// if no resource was selected
		idx, res := ociterm.guiController.GetGUITopPanel().GetResourcesDropDown().GetCurrentOption()
		if idx == -1 || res == "" {
			return
		}
		// if current panel is the one selected
		if ociterm.currentPanel != nil {
			(*ociterm.currentPanel).Remove(ociterm.mainPages)
		}
		conf, err := ociterm.GetBasicConfiguration()
		if err != nil {
			ociterm.guiController.LogError(err.Error(), true)
		}
		// switching region
		selReg := ociterm.guiController.GetGUITopPanel().GetSelectedRegionName()
		log.Printf("region: %s", selReg)
		if selReg != "" {
			ociterm.ociController.ChangeRegion(selReg)
		}
		switch res {
		case "compartments":
			ociterm.currentPanel = gui.NewCompartmentAsGUIPanel(conf.TenancyId, conf.CompartmentId, ociterm.ociController, ociterm.guiController)
			(*ociterm.currentPanel).Show(ociterm.mainPages)
			ociterm.errorTextArea.SetText((*ociterm.currentPanel).GetInfo())
		case "instances":
			// compartment has to be selected
			if ociterm.guiController.GetGUITopPanel().GetSelectedCompartmentId() != "" {
				ociterm.currentPanel = gui.NewInstancesAsGUIPanel(conf.TenancyId, (*ociterm.guiController.GetGUITopPanel()).GetSelectedCompartmentId(), ociterm.ociController, ociterm.guiController)
				(*ociterm.currentPanel).Show(ociterm.mainPages)
				ociterm.errorTextArea.SetText((*ociterm.currentPanel).GetInfo())
			} else {
				ociterm.guiController.LogError("compartment has to be selected", true)
			}
		}
	})
}

func (ociterm *OciTerm) Run() {
	if err := ociterm.app.SetRoot(ociterm.guiController.GetGUIPages(), true).EnableMouse(true).Run(); err != nil {
		panic(err)
	}
	defer ociterm.ociController.CloseContext()
}

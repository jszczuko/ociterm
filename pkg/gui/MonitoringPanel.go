package gui

import (
	"fmt"
	"time"

	oci "github.com/jszczuko/ociterm/pkg/oci"
	"github.com/jszczuko/plot4tview/pkg/gui"
	"github.com/oracle/oci-go-sdk/v52/core"
	"github.com/rivo/tview"
)

type InstanceMonitoringPanel struct {
	guiController *GuiController
	ociController *oci.OCIController
	gui           *instanceMonitoringGUI
	data          *instanceMonitoringData
}

type instanceMonitoringData struct {
	instance   *core.Instance
	memoryData *[][]float64
	cpuData    *[][]float64
}

type instanceMonitoringGUI struct {
	memoryBar *gui.BarPlot
	cpuBar    *gui.BarPlot
	mainGrid  *tview.Grid
}

func NewInstanceMonitoringPanel(GuiController *GuiController, OciController *oci.OCIController, Instance *core.Instance) *InstanceMonitoringPanel {
	res := InstanceMonitoringPanel{
		guiController: GuiController,
		ociController: OciController,
		gui:           newInstanceMonitoringGUI(),
		data:          newInstanceMonitoringData(Instance),
	}
	return &res
}

func newInstanceMonitoringData(inst *core.Instance) *instanceMonitoringData {
	res := instanceMonitoringData{
		instance:   inst,
		memoryData: nil,
		cpuData:    nil,
	}
	return &res
}

func newInstanceMonitoringGUI() *instanceMonitoringGUI {
	res := instanceMonitoringGUI{
		memoryBar: gui.NewBarPlot(),
		cpuBar:    gui.NewBarPlot(),
		mainGrid:  tview.NewGrid(),
	}
	res.memoryBar.SetAxis2String(func(value float64) string {
		return time.Unix(int64(value), 0).Format("2006-01-02 15:04:05")
	}, func(value float64) string {
		return fmt.Sprintf("%f", value)
	})
	res.cpuBar.SetAxis2String(func(value float64) string {
		return time.Unix(int64(value), 0).Format("2006-01-02 15:04:05")
	}, func(value float64) string {
		return fmt.Sprintf("%f", value)
	})
	return &res
}

func (panel *InstanceMonitoringPanel) createGUI() {
	panel.gui.mainGrid.SetColumns(0, 200, 0)
	panel.gui.mainGrid.SetRows(0, 140, 140, 0)

	panel.gui.mainGrid.AddItem(panel.gui.cpuBar, 2, 2, 1, 1, 130, 190, false)
	panel.gui.mainGrid.AddItem(panel.gui.memoryBar, 3, 2, 1, 1, 130, 190, false)

	panel.gui.mainGrid.SetBorder(true).SetTitle("Instance cpu/memory max over 10 min from last 24 h")
}

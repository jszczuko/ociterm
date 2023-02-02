package gui

import (
	// oci "github.com/jszczuko/ociterm/pkg/oci"
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
	res.createGUI()
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

func (panel *InstanceMonitoringPanel) GetGUI() tview.Primitive {
	return panel.gui.mainGrid
}

func (panel *InstanceMonitoringPanel) GetPanelName() string {
	return "InstanceMonitoringPanel"
}

func (panel *InstanceMonitoringPanel) createGUI() {
	grid := tview.NewGrid()
	grid.SetColumns(200)
	grid.SetRows(20, 20)

	panel.gui.mainGrid.SetColumns(0, 200, 0)
	panel.gui.mainGrid.SetRows(0, 40, 0)

	panel.gui.cpuBar.SetBorder(true).SetTitle("CPU")
	panel.gui.memoryBar.SetBorder(true).SetTitle("Memory")

	panel.gui.cpuBar.SetXAxisText("Time", 0)
	panel.gui.cpuBar.SetYAxisText("% of CPU", 1)

	panel.gui.memoryBar.SetXAxisText("Time", 0)
	panel.gui.memoryBar.SetYAxisText("% of Memory", 1)

	grid.AddItem(panel.gui.cpuBar, 0, 0, 1, 1, 0, 0, false)
	grid.AddItem(panel.gui.memoryBar, 1, 0, 1, 1, 0, 0, false)

	grid.SetBorder(true).SetTitle("Instance cpu/memory max over 10 min from last 24 h")

	panel.gui.mainGrid.AddItem(grid, 1, 1, 1, 1, 0, 0, false)

}

package gui

import (
	"fmt"
	"sync"
	"time"

	"github.com/gdamore/tcell/v2"
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
	instance      *core.Instance
	compartmentId string
	memoryData    *[][]float64
	cpuData       *[][]float64
}

type instanceMonitoringGUI struct {
	memoryBar  *gui.BarPlot
	cpuBar     *gui.BarPlot
	mainGrid   *tview.Grid
	exitButton *tview.Button
}

func NewInstanceMonitoringPanel(GuiController *GuiController, OciController *oci.OCIController, Instance *core.Instance, CompartmentId string) *InstanceMonitoringPanel {
	res := InstanceMonitoringPanel{
		guiController: GuiController,
		ociController: OciController,
		gui:           newInstanceMonitoringGUI(),
		data:          newInstanceMonitoringData(Instance, CompartmentId),
	}
	res.createGUI()
	return &res
}

func newInstanceMonitoringData(inst *core.Instance, compId string) *instanceMonitoringData {
	res := instanceMonitoringData{
		instance:      inst,
		compartmentId: compId,
		memoryData:    nil,
		cpuData:       nil,
	}
	return &res
}

func newInstanceMonitoringGUI() *instanceMonitoringGUI {
	res := instanceMonitoringGUI{
		memoryBar:  gui.NewBarPlot(),
		cpuBar:     gui.NewBarPlot(),
		mainGrid:   tview.NewGrid(),
		exitButton: tview.NewButton("Close"),
	}
	res.memoryBar.SetAxis2String(func(value float64) string {
		return time.Unix(int64(value), 0).Format("15:04")
	}, func(value float64) string {
		return fmt.Sprintf("%.2f", value)
	})
	res.cpuBar.SetAxis2String(func(value float64) string {
		return time.Unix(int64(value), 0).Format("15:04")
	}, func(value float64) string {
		return fmt.Sprintf("%.2f", value)
	})

	pointStyle := func(point []float64) tcell.Style {
		if point[1] < 50 {
			return tcell.StyleDefault.Background(tcell.ColorGreen)
		} else if point[1] < 60 {
			return tcell.StyleDefault.Background(tcell.ColorDarkGreen)
		} else if point[1] < 70 {
			return tcell.StyleDefault.Background(tcell.ColorGreenYellow)
		} else if point[1] < 80 {
			return tcell.StyleDefault.Background(tcell.ColorLightYellow)
		} else if point[1] < 90 {
			return tcell.StyleDefault.Background(tcell.ColorYellow)
		} else {
			return tcell.StyleDefault.Background(tcell.ColorRed)
		}
	}

	res.memoryBar.SetStyleForPointFunc(pointStyle)
	res.cpuBar.SetStyleForPointFunc(pointStyle)

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
	grid.SetColumns(82, 8, 82)
	grid.SetRows(20, 20, 1)

	panel.gui.mainGrid.SetColumns(0, 172, 0)
	panel.gui.mainGrid.SetRows(0, 41, 0)

	panel.gui.cpuBar.SetBorder(true).SetTitle(" CPU ")
	panel.gui.memoryBar.SetBorder(true).SetTitle(" Memory ")

	panel.gui.cpuBar.SetXAxisText("Time", 0)
	panel.gui.cpuBar.SetYAxisText("% of CPU", 1)

	panel.gui.memoryBar.SetXAxisText("Time", 0)
	panel.gui.memoryBar.SetYAxisText("% of Memory", 1)

	grid.AddItem(panel.gui.cpuBar, 0, 0, 1, 3, 0, 0, false)
	grid.AddItem(panel.gui.memoryBar, 1, 0, 1, 3, 0, 0, false)
	grid.AddItem(panel.gui.exitButton, 2, 1, 1, 1, 0, 0, true)
	// placeholders
	grid.AddItem(tview.NewTextView().SetBorder(false), 2, 0, 1, 1, 0, 0, false)
	grid.AddItem(tview.NewTextView().SetBorder(false), 2, 2, 1, 1, 0, 0, false)

	grid.SetBorder(true).SetTitle("Instance cpu/memory max over 10 min from last 24 h")

	panel.gui.mainGrid.AddItem(grid, 1, 1, 1, 1, 0, 0, false)
}

func (panel *InstanceMonitoringPanel) LoadData() {

	var wg sync.WaitGroup

	wg.Add(2)
	go func() {
		defer wg.Done()
		data, err := panel.ociController.CpuUtilization10mLast24hMax(panel.data.compartmentId, *panel.data.instance.Id)
		if err != nil {
			panel.gui.cpuBar.SetNoDataText("No data loaded.")
		} else {
			panel.gui.cpuBar.SetData(Float64MapToArray(data))
		}
	}()
	go func() {
		defer wg.Done()
		data, err := panel.ociController.MemoryUtilization10mLast24hMax(panel.data.compartmentId, *panel.data.instance.Id)
		if err != nil {
			panel.gui.memoryBar.SetNoDataText("No data loaded.")
		} else {
			panel.gui.memoryBar.SetData(Float64MapToArray(data))
		}
	}()
	wg.Wait()
}

func Float64MapToArray(floatMap map[float64]float64) [][]float64 {
	res := make([][]float64, len(floatMap))
	i := 0
	for k, v := range floatMap {
		point := make([]float64, 2)
		point[0] = k
		point[1] = v
		res[i] = point
		i++
	}
	return res
}

package gui

import (
	"github.com/jszczuko/plot4tview/pkg/gui"
	"github.com/oracle/oci-go-sdk/v52/core"
)

type InstanceMonitoringPanel struct {
	instance  *core.Instance
	memoryBar *gui.BarPlot
	cpuBar    *gui.BarPlot
}

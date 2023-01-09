package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jszczuko/plot4tview/pkg/gui"
	"github.com/rivo/tview"
)

func main() {
	file, err := os.OpenFile("logs.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}
	log.SetOutput(file)

	app := tview.NewApplication()

	data := make([][]float64, 0)

	data_file, err := os.Open("test_data.txt")
	if err != nil {
		log.Fatal(err)
	}

	index := 0

	for {
		var flt float64
		n, err := fmt.Fscanln(data_file, &flt)
		if n == 0 || err != nil {
			break
		}
		point := make([]float64, 2)
		point[0] = float64(index)
		point[1] = flt
		data = append(data, point)
		index++
	}
	p := gui.NewPlot()
	bp := gui.NewBarPlot()
	dp := gui.NewDotPlot()

	p.SetXAxisText("Time", 0)
	p.SetYAxisText("% of memory", 1)

	bp.SetXAxisText("Time", 0)
	bp.SetYAxisText("% of CPU", 1)

	dp.SetXAxisText("Time", 0)
	dp.SetYAxisText("% of CPU", 1)

	maxEnd := len(data) - 1
	start := 0
	stop := 190

	go func() {
		time.Sleep(3 * time.Second)
		p.SetData(data[start:stop])
		bp.SetData(data[start:stop])
		dp.SetData(data[start:stop])
		time.Sleep(2 * time.Second)
		for {
			time.Sleep(500 * time.Millisecond)
			if stop < maxEnd {
				start++
				stop++
			} else {
				return
			}
			app.QueueUpdateDraw(func() {
				p.SetData(data[start:stop])
				bp.SetData(data[start:stop])
				dp.SetData(data[start:stop])
			})
		}
	}()

	flex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(bp.SetPlotTite(" Bar Plot ").SetPlotBorder(true), 0, 1, false).
		AddItem(dp.SetPlotTite(" Point Plot ").SetPlotBorder(true), 0, 1, false).
		AddItem(p.SetPlotTite(" Plot ").SetPlotBorder(true), 0, 1, true)

	if err := app.SetRoot(flex,
		true).EnableMouse(true).Run(); err != nil {
		panic(err)
	}
}

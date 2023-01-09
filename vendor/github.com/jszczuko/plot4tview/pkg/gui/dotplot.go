package gui

import (
	"fmt"
	"math"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type DotPlot struct {
	*tview.Box
	focus         bool
	data          [][]float64
	extremesX     extremes
	extremesY     extremes
	style4Point   func(point []float64) tcell.Style
	xaxis2String  func(value float64) string
	yaxis2String  func(value float64) string
	xaxisText     string
	xaxisAligment int
	yaxisText     string
	yaxisAligment int
	noDataText    string
}

func NewDotPlot() *DotPlot {
	return &DotPlot{
		Box:       tview.NewBox(),
		focus:     false,
		data:      [][]float64{},
		extremesX: extremes{minVal: 0, maxVal: 0},
		extremesY: extremes{minVal: 0, maxVal: 0},
		style4Point: func(point []float64) tcell.Style {
			return tcell.StyleDefault
		},
		xaxis2String: func(value float64) string {
			return fmt.Sprintf("%-6.2f", value)
		},
		yaxis2String: func(value float64) string {
			return fmt.Sprintf("%-6.2f", value)
		},
		xaxisText:  "",
		yaxisText:  "",
		noDataText: "No Data...",
	}
}

func (plot *DotPlot) SetNoDataText(text string) {
	plot.noDataText = text
}

func (plot *DotPlot) SetXAxisText(text string, aligment int) {
	plot.xaxisText = text
	plot.xaxisAligment = aligment
}

func (plot *DotPlot) SetYAxisText(text string, aligment int) {
	plot.yaxisText = text
	plot.yaxisAligment = aligment
}

func (plot *DotPlot) SetData(Data [][]float64) error {
	if len(Data) < 2 {
		return fmt.Errorf("minimal number of points in data is 2, got %d", len(Data))
	}
	var maxValueX float64 = math.SmallestNonzeroFloat64
	var maxValueY float64 = math.SmallestNonzeroFloat64
	var minValueX float64 = math.MaxFloat64
	var minValueY float64 = math.MaxFloat64
	for i, point := range Data {
		if len(point) != 2 {
			return fmt.Errorf("expected array of lenght 2, got %d. At possition %d", len(point), i)
		}
		if point[0] > maxValueX {
			maxValueX = point[0]
		}
		if point[0] < minValueX {
			minValueX = point[0]
		}
		if point[1] > maxValueY {
			maxValueY = point[1]
		}
		if point[1] < minValueY {
			minValueY = point[1]
		}
	}
	plot.extremesX = extremes{maxVal: maxValueX, minVal: minValueX}
	plot.extremesY = extremes{maxVal: maxValueY, minVal: minValueY}
	plot.data = Data
	return nil
}

func (plot *DotPlot) SetAxis2String(Xaxis2String func(value float64) string, Yaxis2String func(value float64) string) {
	plot.xaxis2String = Xaxis2String
	plot.yaxis2String = Yaxis2String
}

func (plot *DotPlot) SetStyleForPointFunc(style4point func(point []float64) tcell.Style) {
	plot.style4Point = style4point
}

// implements tview.Primitive

func (plot *DotPlot) Draw(screen tcell.Screen) {
	plot.Box.DrawForSubclass(screen, plot)
	x, y, width, height := plot.GetInnerRect()

	width = width - 1 // IDK why yet

	// y axis description
	yadX := x
	yadY := y
	yadWidth := width
	yadHeight := 1
	plot.drawYaxisDescription(screen, yadX, yadY, yadWidth, yadHeight)

	// x axis description
	xadX := x
	xadY := y + height - 1
	xadWidth := width
	xadHeight := 1
	plot.drawXaxisDescription(screen, xadX, xadY, xadWidth, xadHeight)

	// y axis lines
	// xaxisOffset := int(math.Max(float64(len(plot.yaxis2String(plot.extremesY.minVal))), float64(len(plot.yaxis2String(plot.extremesY.maxVal))))) + 3
	xaxisOffset := int(math.Round(math.Max(float64(len(plot.yaxis2String(plot.extremesY.minVal))), float64(len(plot.yaxis2String(plot.extremesY.maxVal)))))) + 3
	yalX := x
	yalY := y + 1
	yalWidth := xaxisOffset
	yalHeight := height - 4
	py, ph := plot.drawYaxisLines(screen, yalX, yalY, yalWidth, yalHeight)

	// xTextOffset := int(math.Max(float64(len(plot.xaxis2String(plot.extremesX.minVal))), float64(len(plot.xaxis2String(plot.extremesX.maxVal)))))
	xTextOffset := int(math.Round(math.Max(float64(len(plot.xaxis2String(plot.extremesX.minVal))), float64(len(plot.xaxis2String(plot.extremesX.maxVal))))))

	// x axis lines
	xalX := x + xaxisOffset
	xalY := y + height - 4
	xalWidth := width - xaxisOffset
	xalHeight := 4
	px, pw := plot.drawXaxisLines(screen, xalX, xalY, xalWidth, xalHeight, xTextOffset)

	// points
	pointsX := px
	pointsY := py
	pointsWidth := pw
	pointsHeight := ph
	plot.drawPoints(screen, pointsX, pointsY, pointsWidth, pointsHeight)

	if len(plot.data) == 0 {
		xD, yD, widthD, heightD := plot.GetInnerRect()
		tview.Print(screen, plot.noDataText, xD-len(plot.noDataText)+widthD/2, yD-2+heightD/2, 20, tview.AlignCenter, tcell.ColorRed)
	}

}

func (plot *DotPlot) drawPoints(screen tcell.Screen, x int, y int, width int, height int) {
	var OFFSET rune = '\u2800'
	var POINTS = [4][2]rune{
		{'\u0001', '\u0008'},
		{'\u0002', '\u0010'},
		{'\u0004', '\u0020'},
		{'\u0040', '\u0080'},
	}

	type pointRune struct {
		r     rune
		style tcell.Style
	}
	// map keys are x, y
	pointMap := make(map[int]map[int]pointRune)

	coefficientX := float64(width) / (plot.extremesX.maxVal - plot.extremesX.minVal)
	coefficientY := float64(height) / (plot.extremesY.maxVal - plot.extremesY.minVal)
	for _, point := range plot.data {
		possRawX := (point[0] - plot.extremesX.minVal) * coefficientX
		possRawY := (plot.extremesY.maxVal - point[1]) * coefficientY
		possRawRoundX := math.Round(possRawX)
		possRawRoundY := math.Round(possRawY)
		possX := x + int(possRawRoundX)
		possY := y + int(possRawRoundY)

		pointX := int((possRawX - float64(int(possRawX))) * 2)
		pointY := int((possRawY - float64(int(possRawY))) * 4)

		_, pX := pointMap[possX]
		if !pX {
			pointMap[possX] = make(map[int]pointRune)
		}
		_, pY := pointMap[possX][possY]
		if !pY {
			pointMap[possX][possY] = pointRune{r: OFFSET, style: plot.style4Point(point)}
		}
		// if window is to small it can go less than 0
		if pointX > -1 && pointY > -1 {
			pr := pointMap[possX][possY]
			pr.r = pr.r | POINTS[pointY][pointX]
			pointMap[possX][possY] = pr
		}
	}
	for x, v := range pointMap {
		for y, r := range v {
			screen.SetContent(x, y, r.r, nil, r.style)
		}

	}

}

func (plot *DotPlot) drawXaxisLines(screen tcell.Screen, x int, y int, width int, height int, xOffset int) (xm int, wm int) {

	minX := 0
	maxX := 0

	xLines := int((float64(width-xOffset) / float64(xOffset)))
	// difference between steps
	stepDiff := (plot.extremesX.maxVal - plot.extremesX.minVal) / float64(xLines)

	xIndex := int(xOffset / 2)

	for i := 0; i <= xLines; i++ {
		val := plot.extremesX.minVal + float64(i)*stepDiff
		screen.SetContent(x+xIndex, y, '|', nil, tcell.StyleDefault)

		if i == 0 {
			minX = x + xIndex
		}
		if i == xLines {
			maxX = x + xIndex
		}
		h := 0
		if i%2 == 0 {
			h = 1
		} else {
			h = 2
		}
		tview.Print(screen, plot.xaxis2String(val), x+xIndex-int(xOffset/2), y+h, 9, 1, tcell.ColorRed)
		xIndex += xOffset
	}
	return minX, maxX - minX
}

func (plot *DotPlot) drawYaxisLines(screen tcell.Screen, x int, y int, width int, height int) (ym int, hm int) {
	yLines := int(math.Floor((float64(height) / float64(2))))
	stepDiffY := (plot.extremesY.maxVal - plot.extremesY.minVal) / float64(yLines-1)

	for i := 0; i < yLines; i++ {
		val := plot.extremesY.maxVal - float64(i)*stepDiffY
		screen.SetContent(width-1, y+i*2, '-', nil, tcell.StyleDefault)
		tview.Print(screen, plot.yaxis2String(val), x, y+i*2, 9, 1, tcell.ColorRed)
	}
	return y, (yLines - 1) * 2
}

func (plot *DotPlot) drawYaxisDescription(screen tcell.Screen, x int, y int, width int, height int) {
	tview.Print(screen, plot.yaxisText, x, y, width, plot.xaxisAligment, tcell.ColorRed)
}

func (plot *DotPlot) drawXaxisDescription(screen tcell.Screen, x int, y int, width int, height int) {
	tview.Print(screen, plot.xaxisText, x, y, width, plot.yaxisAligment, tcell.ColorRed)
}

// InputHandler returns a handler which receives key events when it has focus.
// It is called by the Application class.
//
// A value of nil may also be returned, in which case this primitive cannot
// receive focus and will not process any key events.
//
// The handler will receive the key event and a function that allows it to
// set the focus to a different primitive, so that future key events are sent
// to that primitive.
//
// The Application's Draw() function will be called automatically after the
// handler returns.
//
// The Box class provides functionality to intercept keyboard input. If you
// subclass from Box, it is recommended that you wrap your handler using
// Box.WrapInputHandler() so you inherit that functionality.
func (plot *DotPlot) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	// TODO
	return nil
}

// Focus is called by the application when the primitive receives focus.
// Implementers may call delegate() to pass the focus on to another primitive.
func (plot *DotPlot) Focus(delegate func(p tview.Primitive)) {
	plot.focus = true
}

// HasFocus determines if the primitive has focus. This function must return
// true also if one of this primitive's child elements has focus.
func (plot *DotPlot) HasFocus() bool {
	return plot.focus
}

// Blur is called by the application when the primitive loses focus.
func (plot *DotPlot) Blur() {
	plot.focus = false
}

// MouseHandler returns a handler which receives mouse events.
// It is called by the Application class.
//
// A value of nil may also be returned to stop the downward propagation of
// mouse events.
//
// The Box class provides functionality to intercept mouse events. If you
// subclass from Box, it is recommended that you wrap your handler using
// Box.WrapMouseHandler() so you inherit that functionality.
func (plot *DotPlot) MouseHandler() func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(p tview.Primitive)) (consumed bool, capture tview.Primitive) {
	return plot.WrapMouseHandler(func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(p tview.Primitive)) (consumed bool, capture tview.Primitive) {
		x, y := event.Position()
		_, rectY, _, _ := plot.GetInnerRect()
		if !plot.InRect(x, y) {
			return false, nil
		}

		// Process mouse event.
		if y == rectY {
			if action == tview.MouseLeftDown {
				setFocus(plot)
			}
		}
		return
	})
}

func (b *DotPlot) SetPlotTite(title string) *DotPlot {
	b.Box.SetTitle(title)
	return b
}

func (b *DotPlot) SetPlotBorder(border bool) *DotPlot {
	b.Box.SetBorder(border)
	return b
}

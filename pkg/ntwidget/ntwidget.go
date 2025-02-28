package ntwidget

import (
	"image/color"
	"math"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
)

// RangeSlider defines a custom slider with two draggable circular handles
type RangeSlider struct {
	widget.BaseWidget
	Min, Max   float64
	Start, End float64
	changed    bool
	OnChanged  func()

	barBackground  *canvas.Rectangle // Light grey full bar
	barRange       *canvas.Rectangle // Dark grey range bar
	handleDiameter float32           // Diameter of circular handle
	startHandle    *canvas.Circle
	endHandle      *canvas.Circle
	dragging       string // "start" or "end"
}

// Ensure RangeSlider implements fyne.Draggable and desktop.Mouseable
var _ fyne.Draggable = (*RangeSlider)(nil)
var _ desktop.Mouseable = (*RangeSlider)(nil)

// NewRangeSlider creates a new range slider with circular handles
func NewRangeSlider(min, max, start, end float64) *RangeSlider {
	rs := &RangeSlider{
		Min:           min,
		Max:           max,
		Start:         start,
		End:           end,
		barBackground: canvas.NewRectangle(color.Gray{Y: 200}), // Light grey
		barRange:      canvas.NewRectangle(color.Gray{Y: 100}), // Dark grey
		startHandle:   canvas.NewCircle(color.Gray{Y: 100}),    // Dark grey
		endHandle:     canvas.NewCircle(color.Gray{Y: 100}),    // Dark grey
		// startHandle:   canvas.NewCircle(color.NRGBA{R: 200, G: 50, B: 50, A: 255}),  // Red
		// endHandle:     canvas.NewCircle(color.NRGBA{R: 50, G: 200, B: 50, A: 255}),  // Green
	}
	rs.ExtendBaseWidget(rs)
	return rs
}

// update range slider values: Min, Max, Start, End
func (rs *RangeSlider) UpdateValues(min, max, start, end float64) {
	rs.Min = min
	rs.Max = max
	rs.Start = start
	rs.End = end
}

// Layout positions/size the components, ensuring the handles are circular
func (rs *RangeSlider) Layout(size fyne.Size) {
	barHeight := float32(4)
	rs.handleDiameter = 20 // Diameter of circular handle

	// Set full-width background bar (light grey)
	rs.barBackground.Resize(fyne.NewSize(size.Width, barHeight))
	rs.barBackground.Move(fyne.NewPos(0, size.Height/2-barHeight/2))

	// Compute handle positions
	startX := rs.valueToPosition(rs.Start, size.Width)
	endX := rs.valueToPosition(rs.End, size.Width)

	// Set range bar (dark grey) between the handles
	rs.barRange.Resize(fyne.NewSize(endX-startX, barHeight))
	rs.barRange.Move(fyne.NewPos(startX, size.Height/2-barHeight/2))

	// Start handle size/position (circle)
	rs.startHandle.Resize(fyne.NewSize(rs.handleDiameter, rs.handleDiameter))
	rs.startHandle.Move(fyne.NewPos(startX-rs.handleDiameter/2, size.Height/2-rs.handleDiameter/2))

	// End handle size/position (circle)
	rs.endHandle.Resize(fyne.NewSize(rs.handleDiameter, rs.handleDiameter))
	rs.endHandle.Move(fyne.NewPos(endX-rs.handleDiameter/2, size.Height/2-rs.handleDiameter/2))
}

// Implements fyne.Widget interface
func (rs *RangeSlider) CreateRenderer() fyne.WidgetRenderer {
	return &rangeSliderRenderer{slider: rs}
}

// MouseDown: Detect which handle is clicked and start dragging
func (rs *RangeSlider) MouseDown(e *desktop.MouseEvent) {
	if e.Button == desktop.MouseButtonPrimary {
		rs.detectHandle(e.Position.X) // => update the rs.dragging to be "start" handler or "end" handler
	}
}

// MouseUp: Stop dragging when mouse is released
func (rs *RangeSlider) MouseUp(e *desktop.MouseEvent) {
	rs.dragging = "" // => reset the rs.dragging to be ""
}

// Dragged: Move the selected handle based on mouse movement and refresh UI
func (rs *RangeSlider) Dragged(e *fyne.DragEvent) {
	width := rs.Size().Width

	if rs.dragging == "start" {
		newStart := rs.positionToValue(e.Position.X, width)
		if newStart > rs.End {
			newStart = rs.End
		}
		if newStart != rs.Start { //Only update if the value changed
			rs.Start = newStart
			rs.changed = true
		}

	} else if rs.dragging == "end" {
		newEnd := rs.positionToValue(e.Position.X, width)
		if newEnd < rs.Start {
			newEnd = rs.Start
		}
		if newEnd != rs.Start { //Only update if the value changed
			rs.End = newEnd
			rs.changed = true
		}
	}

	if rs.changed { // Prevent unnecessary UI updates
		// update the barRange, startHandler or endHandler based on the rs.Start / rs.End changes
		rs.updateGraphics()
		if rs.OnChanged != nil {
			rs.OnChanged()
			rs.changed = false
		}
	}
}

// DragEnd: Stop dragging
func (rs *RangeSlider) DragEnd() {
	//fmt.Println("DragEnd called")
	rs.dragging = "" // => reset the rs.dragging to be ""
}

// Detects if the user clicked on a handle
func (rs *RangeSlider) detectHandle(mouseX float32) {
	width := rs.Size().Width
	startX := rs.valueToPosition(rs.Start, width)
	endX := rs.valueToPosition(rs.End, width)

	// Click detection with threshold for circular handles
	handleRadius := float64(rs.handleDiameter / 2) // Half of the handle size
	if math.Abs(float64(mouseX-startX)) < handleRadius {
		rs.dragging = "start"
	} else if math.Abs(float64(mouseX-endX)) < handleRadius {
		rs.dragging = "end"
	}
}

// Converts value to position (px)
func (rs *RangeSlider) valueToPosition(value float64, width float32) float32 {
	return float32((value - rs.Min) / (rs.Max - rs.Min) * float64(width))
}

// Converts position (px) to value
func (rs *RangeSlider) positionToValue(pos float32, width float32) float64 {
	val := rs.Min + float64(pos)/float64(width)*(rs.Max-rs.Min)
	if val < rs.Min {
		val = rs.Min
	} else if val > rs.Max {
		val = rs.Max
	}
	return val
}

// Updates the graphical components on screen when dragging
func (rs *RangeSlider) updateGraphics() {
	rs.Layout(rs.Size()) // Update positions by calling the layout method again and redraw bar, handlers
	canvas.Refresh(rs.barRange)
	canvas.Refresh(rs.startHandle)
	canvas.Refresh(rs.endHandle)
}

// Custom Renderer for RangeSlider
type rangeSliderRenderer struct {
	slider *RangeSlider
}

func (r *rangeSliderRenderer) Layout(size fyne.Size) {
	r.slider.Layout(size)
}

// Fix: Set a static MinSize to avoid infinite recursion
func (r *rangeSliderRenderer) MinSize() fyne.Size {
	return fyne.NewSize(200, 40)
}

func (r *rangeSliderRenderer) Refresh() {
	r.slider.updateGraphics()
	canvas.Refresh(r.slider.barBackground)
	canvas.Refresh(r.slider.barRange)
	canvas.Refresh(r.slider.startHandle)
	canvas.Refresh(r.slider.endHandle)
}

func (r *rangeSliderRenderer) BackgroundColor() color.Color {
	return color.Transparent
}

func (r *rangeSliderRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.slider.barBackground, r.slider.barRange, r.slider.startHandle, r.slider.endHandle}
}

func (r *rangeSliderRenderer) Destroy() {}

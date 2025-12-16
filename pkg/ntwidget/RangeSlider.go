package ntwidget

import (
	"image/color"
	"math"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// RangeSlider defines a custom slider with two draggable circular handles.
type RangeSlider struct {
	widget.BaseWidget

	Min, Max   float64
	Start, End float64

	changed   bool
	OnChanged func()

	barBackground  *canvas.Rectangle
	barRange       *canvas.Rectangle
	handleDiameter float32
	startHandle    *canvas.Circle
	endHandle      *canvas.Circle

	dragging string // "start" or "end"
}

var _ fyne.Draggable = (*RangeSlider)(nil)
var _ desktop.Mouseable = (*RangeSlider)(nil)

func NewRangeSlider(min, max, start, end float64) *RangeSlider {
	rs := &RangeSlider{
		Min:   min,
		Max:   max,
		Start: start,
		End:   end,

		// IMPORTANT: non-transparent defaults so it renders immediately
		barBackground: canvas.NewRectangle(color.NRGBA{0, 0, 0, 25}),
		barRange:      canvas.NewRectangle(color.NRGBA{0, 0, 0, 60}),
		startHandle:   canvas.NewCircle(theme.Color(theme.ColorNamePrimary)),
		endHandle:     canvas.NewCircle(theme.Color(theme.ColorNamePrimary)),
	}

	rs.ExtendBaseWidget(rs)

	// Optional: force one initial refresh (safe)
	rs.Refresh()

	return rs
}

func (rs *RangeSlider) UpdateValues(min, max, start, end float64) {
	rs.Min = min
	rs.Max = max
	rs.Start = start
	rs.End = end
	rs.Refresh()
}

func (rs *RangeSlider) CreateRenderer() fyne.WidgetRenderer {
	return &rangeSliderRenderer{slider: rs}
}

func (rs *RangeSlider) Layout(size fyne.Size) {
	barHeight := float32(4)
	rs.handleDiameter = 26

	// background bar
	rs.barBackground.Resize(fyne.NewSize(size.Width, barHeight))
	rs.barBackground.Move(fyne.NewPos(0, size.Height/2-barHeight/2))

	// positions
	startX := rs.valueToPosition(rs.Start, size.Width)
	endX := rs.valueToPosition(rs.End, size.Width)
	if endX < startX {
		endX = startX
	}

	// range bar
	rs.barRange.Resize(fyne.NewSize(endX-startX, barHeight))
	rs.barRange.Move(fyne.NewPos(startX, size.Height/2-barHeight/2))

	// handles
	rs.startHandle.Resize(fyne.NewSize(rs.handleDiameter, rs.handleDiameter))
	rs.startHandle.Move(fyne.NewPos(startX-rs.handleDiameter/2, size.Height/2-rs.handleDiameter/2))

	rs.endHandle.Resize(fyne.NewSize(rs.handleDiameter, rs.handleDiameter))
	rs.endHandle.Move(fyne.NewPos(endX-rs.handleDiameter/2, size.Height/2-rs.handleDiameter/2))
}

func (rs *RangeSlider) MouseDown(e *desktop.MouseEvent) {
	if e.Button == desktop.MouseButtonPrimary {
		rs.detectHandle(e.Position.X)
	}
}

func (rs *RangeSlider) MouseUp(*desktop.MouseEvent) {
	rs.dragging = ""
}

func (rs *RangeSlider) Dragged(e *fyne.DragEvent) {
	width := rs.Size().Width
	if width <= 0 || rs.Max <= rs.Min {
		return
	}

	if rs.dragging == "start" {
		newStart := rs.positionToValue(e.Position.X, width)
		if newStart > rs.End {
			newStart = rs.End
		}
		if newStart != rs.Start {
			rs.Start = newStart
			rs.changed = true
		}
	} else if rs.dragging == "end" {
		newEnd := rs.positionToValue(e.Position.X, width)
		if newEnd < rs.Start {
			newEnd = rs.Start
		}
		if newEnd != rs.End {
			rs.End = newEnd
			rs.changed = true
		}
	}

	if rs.changed {
		rs.updateGraphics()
		if rs.OnChanged != nil {
			rs.OnChanged()
		}
		rs.changed = false
	}
}

func (rs *RangeSlider) DragEnd() {
	rs.dragging = ""
}

func (rs *RangeSlider) detectHandle(mouseX float32) {
	width := rs.Size().Width
	startX := rs.valueToPosition(rs.Start, width)
	endX := rs.valueToPosition(rs.End, width)

	handleRadius := float64(rs.handleDiameter / 2)
	if math.Abs(float64(mouseX-startX)) < handleRadius {
		rs.dragging = "start"
	} else if math.Abs(float64(mouseX-endX)) < handleRadius {
		rs.dragging = "end"
	}
}

func (rs *RangeSlider) valueToPosition(value float64, width float32) float32 {
	if rs.Max <= rs.Min || width <= 0 {
		return 0
	}
	return float32((value - rs.Min) / (rs.Max - rs.Min) * float64(width))
}

func (rs *RangeSlider) positionToValue(pos float32, width float32) float64 {
	if rs.Max <= rs.Min || width <= 0 {
		return rs.Min
	}
	val := rs.Min + float64(pos)/float64(width)*(rs.Max-rs.Min)
	if val < rs.Min {
		val = rs.Min
	} else if val > rs.Max {
		val = rs.Max
	}
	return val
}

func (rs *RangeSlider) updateGraphics() {
	rs.Layout(rs.Size())
	canvas.Refresh(rs.barBackground)
	canvas.Refresh(rs.barRange)
	canvas.Refresh(rs.startHandle)
	canvas.Refresh(rs.endHandle)
}

// -----------------------------------------------------------------------------
// Renderer
// -----------------------------------------------------------------------------

type rangeSliderRenderer struct {
	slider *RangeSlider
}

func (r *rangeSliderRenderer) Refresh() {
	primary := theme.Color(theme.ColorNamePrimary)
	bg := theme.Color(theme.ColorNameBackground)

	// theme-aware bars
	if isDarkBackground(bg) {
		r.slider.barBackground.FillColor = color.NRGBA{255, 255, 255, 35}
		r.slider.barRange.FillColor = color.NRGBA{255, 255, 255, 70}
	} else {
		r.slider.barBackground.FillColor = color.NRGBA{0, 0, 0, 25}
		r.slider.barRange.FillColor = color.NRGBA{0, 0, 0, 60}
	}

	// handles match toggle switch
	r.slider.startHandle.FillColor = primary
	r.slider.endHandle.FillColor = primary

	r.slider.updateGraphics()
}

func (r *rangeSliderRenderer) Layout(size fyne.Size) {
	r.slider.Layout(size)
}

func (r *rangeSliderRenderer) MinSize() fyne.Size {
	return fyne.NewSize(200, 40)
}

func (r *rangeSliderRenderer) BackgroundColor() color.Color {
	return color.Transparent
}

func (r *rangeSliderRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{
		r.slider.barBackground,
		r.slider.barRange,
		r.slider.startHandle,
		r.slider.endHandle,
	}
}

func (r *rangeSliderRenderer) Destroy() {}

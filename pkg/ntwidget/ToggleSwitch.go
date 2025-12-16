// Package ntwidget provides a themed ON/OFF switch widget for Fyne v2.7+.
package ntwidget

import (
	"image/color"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// Switch is a simple ON/OFF toggle control.
type Switch struct {
	widget.BaseWidget

	Checked   bool
	OnChanged func(bool)
	bound     binding.Bool

	hasFocus bool
}

func NewToggleswitch(initial bool, onChanged func(bool)) *Switch {
	s := &Switch{Checked: initial, OnChanged: onChanged}
	s.ExtendBaseWidget(s)
	return s
}

func (s *Switch) Bind(b binding.Bool) {
	s.bound = b
	if v, err := b.Get(); err == nil {
		s.setCheckedInternal(v, true)
	}
	b.AddListener(binding.NewDataListener(func() {
		if v, err := b.Get(); err == nil {
			fyne.Do(func() { s.setCheckedInternal(v, true) })
		}
	}))
}

func (s *Switch) SetChecked(v bool) { s.setCheckedInternal(v, false) }
func (s *Switch) Toggle()           { s.SetChecked(!s.Checked) }

func isDisabled(o fyne.CanvasObject) bool {
	if d, ok := o.(fyne.Disableable); ok {
		return d.Disabled()
	}
	return false
}

func (s *Switch) setCheckedInternal(v bool, fromBinding bool) {
	if s.Checked == v || isDisabled(s) {
		return
	}
	s.Checked = v
	if !fromBinding {
		if s.bound != nil {
			_ = s.bound.Set(v)
		}
		if s.OnChanged != nil {
			s.OnChanged(v)
		}
	}
	s.Refresh()
}

// Interaction
func (s *Switch) Tapped(*fyne.PointEvent) {
	if !isDisabled(s) {
		s.Toggle()
	}
}
func (s *Switch) TappedSecondary(*fyne.PointEvent) {}

func (s *Switch) FocusGained() { s.hasFocus = true; s.Refresh() }
func (s *Switch) FocusLost()   { s.hasFocus = false; s.Refresh() }

func (s *Switch) Focused() bool    { return s.hasFocus }
func (s *Switch) TypedRune(r rune) {}
func (s *Switch) TypedKey(ev *fyne.KeyEvent) {
	if isDisabled(s) {
		return
	}
	if ev.Name == fyne.KeySpace || ev.Name == fyne.KeyReturn {
		s.Toggle()
	}
}

// -----------------------------------------------------------------------------
// Rendering
// -----------------------------------------------------------------------------

type switchRenderer struct {
	s     *Switch
	track *canvas.Rectangle
	knob  *canvas.Circle
	focus *canvas.Rectangle

	objs        []fyne.CanvasObject
	lastSize    fyne.Size
	prevChecked bool
	anim        *fyne.Animation
}

func (s *Switch) CreateRenderer() fyne.WidgetRenderer {
	r := &switchRenderer{
		s:     s,
		track: canvas.NewRectangle(color.Transparent),
		knob:  canvas.NewCircle(color.Transparent),
		focus: canvas.NewRectangle(color.Transparent),
	}
	r.objs = []fyne.CanvasObject{r.track, r.knob, r.focus}
	r.prevChecked = s.Checked
	r.Refresh()
	return r
}

func (r *switchRenderer) Destroy() {}

func (r *switchRenderer) Layout(sz fyne.Size) {
	r.lastSize = sz

	r.track.Resize(sz)
	r.track.Move(fyne.NewPos(0, 0))
	r.track.CornerRadius = sz.Height / 2

	pad := float32(theme.Padding())
	knobDiam := sz.Height - 2*pad
	if knobDiam < 8 {
		knobDiam = sz.Height
	}
	r.knob.Resize(fyne.NewSize(knobDiam, knobDiam))

	if r.anim == nil {
		r.knob.Move(r.knobPosForState(r.s.Checked))
	}

	r.focus.Resize(sz)
	r.focus.Move(fyne.NewPos(0, 0))
}

func (r *switchRenderer) MinSize() fyne.Size {
	h := theme.IconInlineSize()*0.8 + theme.Padding()*2
	w := h * 2
	return fyne.NewSize(w, h)
}

func (r *switchRenderer) Objects() []fyne.CanvasObject { return r.objs }

func (r *switchRenderer) Refresh() {
	primary := theme.Color(theme.ColorNamePrimary)
	bg := theme.Color(theme.ColorNameBackground)
	disabledCol := theme.Color(theme.ColorNameDisabled)
	focusCol := theme.Color(theme.ColorNameFocus)

	// OFF-state tuning (adjust if desired)
	offFillAlpha := uint8(150) // brightness of OFF track in dark mode

	if isDisabled(r.s) {
		r.track.FillColor = disabledCol
		r.track.StrokeWidth = 0
		r.track.StrokeColor = color.Transparent

	} else if r.s.Checked {
		r.track.FillColor = primary
		r.track.StrokeWidth = 0
		r.track.StrokeColor = color.Transparent

	} else {
		base := toNRGBA(bg)

		if isDarkBackground(bg) {
			// VERY visible OFF state in dark mode
			r.track.FillColor = blendNRGBA(
				base,
				color.NRGBA{255, 255, 255, offFillAlpha}, // bright enough without border
			)
			r.track.StrokeWidth = 0
			r.track.StrokeColor = color.Transparent
		} else {
			// subtle OFF in light mode
			r.track.FillColor = blendNRGBA(base, color.NRGBA{0, 0, 0, 32})
			r.track.StrokeWidth = 0
			r.track.StrokeColor = color.Transparent
		}
	}
	r.track.Refresh()

	// Knob
	r.knob.FillColor = bg
	r.knob.StrokeWidth = 0
	r.knob.Refresh()

	// Focus ring
	if r.s.hasFocus && !isDisabled(r.s) {
		r.focus.FillColor = withAlpha(focusCol, 48)
	} else {
		r.focus.FillColor = color.Transparent
	}
	r.focus.Refresh()

	// Animate knob movement
	if r.prevChecked != r.s.Checked && r.lastSize.Width > 0 {
		if r.anim != nil {
			r.anim.Stop()
			r.anim = nil
		}
		start := r.knob.Position()
		end := r.knobPosForState(r.s.Checked)

		r.anim = canvas.NewPositionAnimation(start, end, 40*time.Millisecond, func(p fyne.Position) {
			r.knob.Move(p)
			r.knob.Refresh()
		})
		r.anim.Curve = fyne.AnimationEaseInOut
		r.anim.Start()
	}
	r.prevChecked = r.s.Checked
}

func (r *switchRenderer) knobPosForState(on bool) fyne.Position {
	pad := float32(theme.Padding())
	knobDiam := r.knob.Size().Width
	y := (r.lastSize.Height - knobDiam) / 2
	if on {
		x := r.lastSize.Width - knobDiam - pad
		return fyne.NewPos(x, y)
	}
	return fyne.NewPos(pad, y)
}

// -----------------------------------------------------------------------------
// Helpers
// -----------------------------------------------------------------------------

func withAlpha(c color.Color, a uint8) color.Color {
	r, g, b, _ := c.RGBA()
	return color.NRGBA{uint8(r >> 8), uint8(g >> 8), uint8(b >> 8), a}
}

func blendNRGBA(base color.NRGBA, overlay color.NRGBA) color.NRGBA {
	br, bg, bb, _ := base.RGBA()
	or, og, ob, oa := overlay.RGBA()
	alpha := float32(oa) / 65535.0
	r := float32(br>>8)*(1-alpha) + float32(or>>8)*alpha
	g := float32(bg>>8)*(1-alpha) + float32(og>>8)*alpha
	b := float32(bb>>8)*(1-alpha) + float32(ob>>8)*alpha
	return color.NRGBA{uint8(r), uint8(g), uint8(b), 255}
}

func toNRGBA(c color.Color) color.NRGBA {
	r, g, b, a := c.RGBA()
	return color.NRGBA{uint8(r >> 8), uint8(g >> 8), uint8(b >> 8), uint8(a >> 8)}
}

// compact toggle row: switch + label
func ToggleRow(sw fyne.CanvasObject, lbl *widget.Label, rowW float32) *fyne.Container {
	lbl.Alignment = fyne.TextAlignLeading

	swSize := sw.MinSize()
	lblSize := lbl.MinSize()
	h := maxF(swSize.Height, lblSize.Height)

	left := container.New(
		layout.NewGridWrapLayout(fyne.NewSize(swSize.Width, h)),
		container.NewCenter(sw),
	)

	rightW := rowW - swSize.Width
	if rightW < 0 {
		rightW = 0
	}
	right := container.New(
		layout.NewGridWrapLayout(fyne.NewSize(rightW, h)),
		container.NewBorder(nil, nil, nil, nil, lbl),
	)

	return container.New(layout.NewHBoxLayout(), left, right)
}

func maxF(a, b float32) float32 {
	if a > b {
		return a
	}
	return b
}

package main

import (
	"fmt"
	"image"
	"image/color"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

const scopeWindow = 200

var (
	scopeColorX = color.RGBA{64, 196, 180, 255} // cyan
	scopeColorY = color.RGBA{240, 180, 90, 255} // orange
)

// renderOscilloscope draws a two-channel rolling trace on a dark grid.
// Both channels share one Y-axis so their relative magnitudes are visible.
func renderOscilloscope(s *Sim, xName, yName string, w, h int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	bg := color.RGBA{13, 13, 16, 255}
	for i := 0; i < w*h*4; i += 4 {
		img.Pix[i] = bg.R
		img.Pix[i+1] = bg.G
		img.Pix[i+2] = bg.B
		img.Pix[i+3] = bg.A
	}

	if s == nil || len(s.SpeciesHistory) < 2 || xName == "" || yName == "" {
		return img
	}

	hist := s.SpeciesHistory
	start := 0
	if len(hist) > scopeWindow {
		start = len(hist) - scopeWindow
	}
	recent := hist[start:]
	if len(recent) < 2 {
		return img
	}

	xs := make([]float64, len(recent))
	ys := make([]float64, len(recent))
	for i, sp := range recent {
		xs[i] = s.speciesValue(sp, xName)
		ys[i] = s.speciesValue(sp, yName)
	}

	// Shared Y-range across both channels, anchored at 0 for reference.
	minV, maxV := xs[0], xs[0]
	for _, v := range xs {
		if v < minV {
			minV = v
		}
		if v > maxV {
			maxV = v
		}
	}
	for _, v := range ys {
		if v < minV {
			minV = v
		}
		if v > maxV {
			maxV = v
		}
	}
	if minV > 0 {
		minV = 0
	}
	if maxV <= minV {
		maxV = minV + 1
	}
	// 5% headroom so the peak isn't glued to the border
	maxV += (maxV - minV) * 0.05

	margin := 10
	plotW := w - 2*margin
	plotH := h - 2*margin
	if plotW < 20 || plotH < 20 {
		return img
	}

	toPx := func(i int, v float64) (int, int) {
		px := margin + int(float64(i)/float64(len(recent)-1)*float64(plotW))
		py := h - margin - int((v-minV)/(maxV-minV)*float64(plotH))
		if py < margin {
			py = margin
		}
		if py > h-margin {
			py = h - margin
		}
		return px, py
	}

	// Grid
	gridC := color.RGBA{40, 45, 55, 255}
	for g := 1; g < 4; g++ {
		y := margin + g*plotH/4
		for x := margin; x < w-margin; x++ {
			img.Set(x, y, gridC)
		}
	}
	for g := 1; g < 5; g++ {
		x := margin + g*plotW/5
		for y := margin; y < h-margin; y++ {
			img.Set(x, y, gridC)
		}
	}

	// Border
	borderC := color.RGBA{70, 80, 95, 255}
	for x := margin; x <= w-margin; x++ {
		img.Set(x, margin, borderC)
		img.Set(x, h-margin, borderC)
	}
	for y := margin; y <= h-margin; y++ {
		img.Set(margin, y, borderC)
		img.Set(w-margin, y, borderC)
	}

	// Traces
	drawTrace := func(vals []float64, c color.Color) {
		for i := 1; i < len(vals); i++ {
			x0, y0 := toPx(i-1, vals[i-1])
			x1, y1 := toPx(i, vals[i])
			drawLinePix(img, x0, y0, x1, y1, c)
		}
		// Current-value dot at right edge
		px, py := toPx(len(vals)-1, vals[len(vals)-1])
		for dx := -2; dx <= 2; dx++ {
			for dy := -2; dy <= 2; dy++ {
				if dx*dx+dy*dy <= 4 {
					xx, yy := px+dx, py+dy
					if xx >= 0 && xx < w && yy >= 0 && yy < h {
						img.Set(xx, yy, c)
					}
				}
			}
		}
	}
	drawTrace(xs, scopeColorX)
	drawTrace(ys, scopeColorY)

	return img
}

// ==================== UI ====================

func (u *UI) buildOscilloscopeTab() fyne.CanvasObject {
	opts := u.speciesOptions()
	u.oscX = widget.NewSelect(opts, func(string) {
		if u.oscCv != nil {
			u.oscCv.Refresh()
		}
	})
	u.oscY = widget.NewSelect(opts, func(string) {
		if u.oscCv != nil {
			u.oscCv.Refresh()
		}
	})

	// Prefer the first two protein channels so loading an oscillator
	// preset immediately shows meaningful traces.
	if len(opts) > 4 {
		u.oscX.SetSelected(opts[3])
		u.oscY.SetSelected(opts[4])
	} else if len(opts) > 1 {
		u.oscX.SetSelected(opts[0])
		u.oscY.SetSelected(opts[1])
	}

	u.oscXVal = mono("—")
	u.oscYVal = mono("—")

	swatchX := canvas.NewRectangle(scopeColorX)
	swatchX.SetMinSize(fyne.NewSize(12, 12))
	swatchY := canvas.NewRectangle(scopeColorY)
	swatchY.SetMinSize(fyne.NewSize(12, 12))

	legend := container.NewHBox(
		swatchX, u.oscXVal,
		widget.NewLabel("    "),
		swatchY, u.oscYVal,
	)

	header := container.NewVBox(
		container.NewHBox(
			widget.NewLabel("X:"), u.oscX,
			widget.NewLabel("Y:"), u.oscY,
			widget.NewSeparator(),
			widget.NewLabel(fmt.Sprintf("Window: last %d ticks (shared Y-axis)", scopeWindow)),
		),
		legend,
	)

	u.oscCv = canvas.NewRaster(func(w, h int) image.Image {
		return renderOscilloscope(u.sim, u.oscX.Selected, u.oscY.Selected, w, h)
	})
	u.oscCv.SetMinSize(fyne.NewSize(900, 450))

	return container.NewBorder(header, nil, nil, nil, u.oscCv)
}

// refreshOscilloscope syncs the channel selectors against the current
// species set, updates the live value readouts, and repaints the raster.
func (u *UI) refreshOscilloscope() {
	if u.oscCv == nil || u.oscX == nil {
		return
	}
	opts := u.speciesOptions()
	if !stringSliceEqual(u.oscX.Options, opts) {
		selX, selY := u.oscX.Selected, u.oscY.Selected
		u.oscX.Options = opts
		u.oscY.Options = opts
		u.oscX.Refresh()
		u.oscY.Refresh()
		stillIn := func(s string) bool {
			for _, o := range opts {
				if o == s {
					return true
				}
			}
			return false
		}
		if stillIn(selX) {
			u.oscX.SetSelected(selX)
		} else if len(opts) > 3 {
			u.oscX.SetSelected(opts[3])
		} else if len(opts) > 0 {
			u.oscX.SetSelected(opts[0])
		}
		if stillIn(selY) {
			u.oscY.SetSelected(selY)
		} else if len(opts) > 4 {
			u.oscY.SetSelected(opts[4])
		} else if len(opts) > 1 {
			u.oscY.SetSelected(opts[1])
		}
	}
	if len(u.sim.SpeciesHistory) > 0 && u.oscXVal != nil {
		last := u.sim.SpeciesHistory[len(u.sim.SpeciesHistory)-1]
		vx := u.sim.speciesValue(last, u.oscX.Selected)
		vy := u.sim.speciesValue(last, u.oscY.Selected)
		u.oscXVal.SetText(fmt.Sprintf("%s = %.3f", u.oscX.Selected, vx))
		u.oscYVal.SetText(fmt.Sprintf("%s = %.3f", u.oscY.Selected, vy))
	}
	u.oscCv.Refresh()
}

func stringSliceEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// silence unused-import warning for strings in some toolchains
var _ = strings.TrimSpace

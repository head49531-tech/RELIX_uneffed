package main

import (
	"fmt"
	"image"
	"image/color"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// RobustnessPoint is one classified sample from a 1D sweep.
type RobustnessPoint struct {
	X      float64
	Class  BehaviorClass
	Period float64
	Alive  float64
}

// robustnessParams enumerates the parameters that can be swept.
// Gene params (BasalScale, HillN) modify every gene in the preset;
// sim params override the SimParams struct.
var robustnessParams = []string{
	"TranslationK", "BasalScale", "HillN",
	"MutationRate", "BaseDeathRate", "DivisionATP", "EnvGlucose",
}

// ==================== Param application ====================

func clonePresetForSweep(p *DNAPreset) *DNAPreset {
	np := &DNAPreset{
		Name:         p.Name,
		Category:     p.Category,
		Description:  p.Description,
		DNA:          append([]byte{}, p.DNA...),
		Genes:        cloneGenes(p.Genes),
		Pathways:     append([]PathwayStep{}, p.Pathways...),
		InitMRNA:     map[string]float64{},
		InitProteins: map[string]float64{},
	}
	for k, v := range p.InitMRNA {
		np.InitMRNA[k] = v
	}
	for k, v := range p.InitProteins {
		np.InitProteins[k] = v
	}
	return np
}

// applyRobustnessTo mutates a preset and/or its params according to one
// swept parameter. Gene params act on every gene in the draft; sim params
// override the corresponding field on params.
func applyRobustnessTo(p *DNAPreset, param string, val float64, params SimParams) (*DNAPreset, SimParams) {
	switch param {
	case "BasalScale":
		for _, g := range p.Genes {
			g.BasalTranscription *= val
		}
	case "HillN":
		for _, g := range p.Genes {
			g.HillN = val
		}
	case "TranslationK":
		params.TranslationK = val
	case "MutationRate":
		params.MutationRate = val
	case "BaseDeathRate":
		params.BaseDeathRate = val
	case "DivisionATP":
		params.DivisionATP = val
	case "EnvGlucose":
		params.EnvGlucose = val
	}
	return p, params
}

// robustnessTargetPreset prefers the pinned A preset, then the currently
// selected preset, then nil.
func (u *UI) robustnessTargetPreset() *DNAPreset {
	if u.presetPinA != nil {
		return u.presetPinA
	}
	return u.selectedPreset()
}

// ==================== 1D sweep runner ====================

func (u *UI) runRobustnessSweep(p *DNAPreset, param string, min, max float64, steps int) []RobustnessPoint {
	out := make([]RobustnessPoint, 0, steps)
	baseParams := classificationParams()
	for i := 0; i < steps; i++ {
		val := min + (max-min)*float64(i)/float64(steps-1)
		np := clonePresetForSweep(p)
		np, params := applyRobustnessTo(np, param, val, baseParams)
		rep := ClassifyPresetWithParams(np, params, 400, int64(i+1))
		out = append(out, RobustnessPoint{
			X: val, Class: rep.Class, Period: rep.Period, Alive: rep.Alive,
		})
	}
	return out
}

// ==================== Class colors / legend ====================

func classColor(c BehaviorClass) color.Color {
	switch c {
	case BehOscillating:
		return color.RGBA{90, 210, 130, 255}
	case BehBistable:
		return color.RGBA{80, 140, 220, 255}
	case BehStable:
		return color.RGBA{120, 120, 130, 255}
	case BehChaotic:
		return color.RGBA{240, 190, 70, 255}
	case BehGrowing:
		return color.RGBA{180, 140, 255, 255}
	case BehExtinct:
		return color.RGBA{220, 80, 90, 255}
	}
	return color.RGBA{60, 60, 70, 255}
}

func classLegendRow() fyne.CanvasObject {
	mk := func(c color.Color, label string) fyne.CanvasObject {
		r := canvas.NewRectangle(c)
		r.SetMinSize(fyne.NewSize(12, 12))
		return container.NewHBox(r, widget.NewLabel(label))
	}
	return container.NewHBox(
		mk(classColor(BehOscillating), "osc"),
		widget.NewLabel("   "),
		mk(classColor(BehBistable), "tog"),
		widget.NewLabel("   "),
		mk(classColor(BehStable), "stable"),
		widget.NewLabel("   "),
		mk(classColor(BehChaotic), "chaotic"),
		widget.NewLabel("   "),
		mk(classColor(BehGrowing), "growing"),
		widget.NewLabel("   "),
		mk(classColor(BehExtinct), "extinct"),
	)
}

// ==================== 1D renderer ====================

func renderRobustness1D(pts []RobustnessPoint, w, h int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	bg := color.RGBA{13, 13, 16, 255}
	for i := 0; i < w*h*4; i += 4 {
		img.Pix[i] = bg.R
		img.Pix[i+1] = bg.G
		img.Pix[i+2] = bg.B
		img.Pix[i+3] = bg.A
	}
	if len(pts) < 2 {
		return img
	}
	margin := 20
	plotW := w - 2*margin
	plotH := h - 2*margin
	if plotW < 20 || plotH < 20 {
		return img
	}
	blockW := plotW / len(pts)
	if blockW < 2 {
		blockW = 2
	}
	for i, pt := range pts {
		x0 := margin + i*blockW
		x1 := x0 + blockW - 1
		c := classColor(pt.Class)
		for x := x0; x <= x1 && x < w-margin; x++ {
			for y := margin; y < h-margin; y++ {
				img.Set(x, y, c)
			}
		}
	}
	border := color.RGBA{70, 80, 95, 255}
	for x := margin; x <= w-margin; x++ {
		img.Set(x, margin, border)
		img.Set(x, h-margin, border)
	}
	for y := margin; y <= h-margin; y++ {
		img.Set(margin, y, border)
		img.Set(w-margin, y, border)
	}
	return img
}

// ==================== 1D tab ====================

func (u *UI) buildRobustnessTab() fyne.CanvasObject {
	u.robustParam = widget.NewSelect(robustnessParams, nil)
	u.robustParam.SetSelected("TranslationK")
	u.robustMin = widget.NewEntry()
	u.robustMin.SetText("0.5")
	u.robustMax = widget.NewEntry()
	u.robustMax.SetText("3.0")
	u.robustSteps = widget.NewEntry()
	u.robustSteps.SetText("15")
	u.robustTarget = widget.NewLabel("(pin A or select a preset)")
	u.robustStatus = mono("—")
	u.robustOutput = widget.NewMultiLineEntry()
	u.robustOutput.Wrapping = fyne.TextWrapOff
	u.robustOutput.TextStyle = fyne.TextStyle{Monospace: true}

	u.robustCv = canvas.NewRaster(func(w, h int) image.Image {
		return renderRobustness1D(u.robustPoints, w, h)
	})
	u.robustCv.SetMinSize(fyne.NewSize(900, 140))

	runBtn := widget.NewButton("Run 1D sweep", func() {
		p := u.robustnessTargetPreset()
		if p == nil {
			dialog.ShowInformation("Robustness", "Pin preset A or select one.", u.window)
			return
		}
		param := u.robustParam.Selected
		minV, errM := strconv.ParseFloat(u.robustMin.Text, 64)
		maxV, errX := strconv.ParseFloat(u.robustMax.Text, 64)
		steps, errS := strconv.Atoi(u.robustSteps.Text)
		if errM != nil || errX != nil || errS != nil || steps < 2 || maxV <= minV {
			u.robustStatus.SetText("Bad range/steps.")
			return
		}
		u.robustTarget.SetText(p.Name)
		u.robustStatus.SetText(fmt.Sprintf("Sweeping %d points…", steps))
		go func() {
			pts := u.runRobustnessSweep(p, param, minV, maxV, steps)
			fyne.Do(func() {
				u.robustPoints = pts
				u.refreshRobustOutput(p.Name, param)
			})
		}()
	})

	header := container.NewVBox(
		container.NewHBox(
			widget.NewLabelWithStyle("Robustness Sweep (1D)", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			widget.NewLabel("Target:"), u.robustTarget,
		),
		container.NewHBox(
			widget.NewLabel("Parameter:"), u.robustParam,
			widget.NewLabel("Min:"), u.robustMin,
			widget.NewLabel("Max:"), u.robustMax,
			widget.NewLabel("Steps:"), u.robustSteps,
			runBtn,
		),
		classLegendRow(),
		u.robustStatus,
		widget.NewSeparator(),
	)

	plotBlock := container.NewVBox(u.robustCv, u.robustOutput)
	return container.NewBorder(header, nil, nil, nil, plotBlock)
}

func (u *UI) refreshRobustOutput(presetName, param string) {
	var sb strings.Builder
	fmt.Fprintf(&sb, "Robustness of '%s' vs %s (%d points)\n\n", presetName, param, len(u.robustPoints))
	fmt.Fprintf(&sb, "%-10s %-10s %-8s %-10s\n", param, "Class", "Period", "Alive")
	fmt.Fprintf(&sb, "%s\n", strings.Repeat("-", 44))
	for _, pt := range u.robustPoints {
		periodStr := "—"
		if pt.Period > 0 {
			periodStr = fmt.Sprintf("%.0f", pt.Period)
		}
		fmt.Fprintf(&sb, "%-10.4f %-10s %-8s %-10.1f\n", pt.X, pt.Class, periodStr, pt.Alive)
	}
	// Summary: oscillating range
	firstOsc, lastOsc := -1, -1
	for i, pt := range u.robustPoints {
		if pt.Class == BehOscillating {
			if firstOsc == -1 {
				firstOsc = i
			}
			lastOsc = i
		}
	}
	if firstOsc >= 0 {
		fmt.Fprintf(&sb, "\nOscillates: %s ∈ [%.3f, %.3f]  (%d/%d)\n",
			param, u.robustPoints[firstOsc].X, u.robustPoints[lastOsc].X,
			lastOsc-firstOsc+1, len(u.robustPoints))
	} else {
		sb.WriteString("\nNo oscillating points in range.\n")
	}
	u.robustOutput.SetText(sb.String())
	u.robustCv.Refresh()
	u.robustStatus.SetText(fmt.Sprintf("Done — %d points classified.", len(u.robustPoints)))
}

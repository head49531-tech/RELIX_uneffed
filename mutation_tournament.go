package main

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"math/rand"
	"sort"
	"strings"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// mutationResult is one mutant's fate.
type mutationResult struct {
	ID        int
	Preset    *DNAPreset
	Class     BehaviorClass
	Period    float64
	Alive     float64
	Deviation float64 // max fractional perturbation applied
}

// ==================== Mutation application ====================

func applyRandomMutations(base *DNAPreset, frac float64, seed int64) *DNAPreset {
	np := clonePresetForSweep(base)
	rng := rand.New(rand.NewSource(seed))
	for _, g := range np.Genes {
		if rng.Float64() < 0.5 {
			g.BasalTranscription *= 1 + (rng.Float64()*2-1)*frac
			if g.BasalTranscription < 0.001 {
				g.BasalTranscription = 0.001
			}
		}
		if rng.Float64() < 0.5 {
			g.ActivationK *= 1 + (rng.Float64()*2-1)*frac
			if g.ActivationK < 0.1 {
				g.ActivationK = 0.1
			}
		}
		if rng.Float64() < 0.5 {
			g.RepressionK *= 1 + (rng.Float64()*2-1)*frac
			if g.RepressionK < 0.1 {
				g.RepressionK = 0.1
			}
		}
		if rng.Float64() < 0.5 {
			g.HillN *= 1 + (rng.Float64()*2-1)*frac
			if g.HillN < 0.5 {
				g.HillN = 0.5
			}
			if g.HillN > 8 {
				g.HillN = 8
			}
		}
	}
	np.Name = fmt.Sprintf("%s mut#%d", base.Name, seed)
	np.Builtin = false
	return np
}

func measureDeviation(base, mut *DNAPreset) float64 {
	maxDev := 0.0
	n := len(base.Genes)
	if len(mut.Genes) < n {
		n = len(mut.Genes)
	}
	for i := 0; i < n; i++ {
		bg, mg := base.Genes[i], mut.Genes[i]
		check := func(b, m float64) {
			if b == 0 {
				return
			}
			d := math.Abs(m-b) / b
			if d > maxDev {
				maxDev = d
			}
		}
		check(bg.BasalTranscription, mg.BasalTranscription)
		check(bg.ActivationK, mg.ActivationK)
		check(bg.RepressionK, mg.RepressionK)
		check(bg.HillN, mg.HillN)
	}
	return maxDev
}

// runMutationTournament produces n mutants and classifies each in parallel.
func (u *UI) runMutationTournament(base *DNAPreset, n int, frac float64, seed int64) []mutationResult {
	results := make([]mutationResult, n)
	var wg sync.WaitGroup
	sem := make(chan struct{}, 4)
	for i := 0; i < n; i++ {
		i := i
		wg.Add(1)
		sem <- struct{}{}
		go func() {
			defer wg.Done()
			defer func() { <-sem }()
			mutSeed := seed + int64(i)*997
			mut := applyRandomMutations(base, frac, mutSeed)
			rep := ClassifyPreset(mut, 300, mutSeed+1)
			dev := measureDeviation(base, mut)
			results[i] = mutationResult{
				ID: i, Preset: mut, Class: rep.Class,
				Period: rep.Period, Alive: rep.Alive, Deviation: dev,
			}
		}()
	}
	wg.Wait()
	return results
}

// ==================== Histogram renderer ====================

var classOrder = []BehaviorClass{
	BehOscillating, BehBistable, BehStable, BehChaotic, BehGrowing, BehExtinct,
}

func renderMutationHistogram(results []mutationResult, w, h int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	bg := color.RGBA{13, 13, 16, 255}
	for i := 0; i < w*h*4; i += 4 {
		img.Pix[i] = bg.R
		img.Pix[i+1] = bg.G
		img.Pix[i+2] = bg.B
		img.Pix[i+3] = bg.A
	}
	if len(results) == 0 {
		return img
	}

	counts := map[BehaviorClass]int{}
	for _, r := range results {
		counts[r.Class]++
	}
	maxCount := 1
	for _, c := range counts {
		if c > maxCount {
			maxCount = c
		}
	}

	margin := 20
	plotW := w - 2*margin
	plotH := h - 2*margin
	if plotW < 40 || plotH < 20 {
		return img
	}
	barW := plotW / len(classOrder)

	for i, cls := range classOrder {
		count := counts[cls]
		frac := float64(count) / float64(maxCount)
		barH := int(frac * float64(plotH))
		x0 := margin + i*barW + 4
		x1 := margin + (i+1)*barW - 4
		y0 := h - margin - barH
		y1 := h - margin
		c := classColor(cls)
		for x := x0; x <= x1 && x < w-margin; x++ {
			for y := y0; y < y1; y++ {
				if x >= 0 && x < w && y >= 0 && y < h {
					img.Set(x, y, c)
				}
			}
		}
		// thin border around bar
		border := color.RGBA{200, 200, 200, 60}
		for x := x0; x <= x1 && x < w-margin; x++ {
			if y0 >= 0 && y0 < h {
				img.Set(x, y0, border)
			}
		}
	}
	// baseline
	axis := color.RGBA{70, 80, 95, 255}
	for x := margin; x < w-margin; x++ {
		img.Set(x, h-margin, axis)
	}
	return img
}

// ==================== Tab UI ====================

func (u *UI) buildMutationTournamentTab() fyne.CanvasObject {
	u.mutTarget = widget.NewLabel("(pin A or select a preset)")
	u.mutCount = widget.NewEntry(); u.mutCount.SetText("20")
	u.mutFrac = widget.NewEntry(); u.mutFrac.SetText("0.20")
	u.mutStatus = mono("—")
	u.mutResults = nil
	u.mutSelected = -1

	u.mutCv = canvas.NewRaster(func(w, h int) image.Image {
		return renderMutationHistogram(u.mutResults, w, h)
	})
	u.mutCv.SetMinSize(fyne.NewSize(600, 160))

	// Column headers
	headers := container.NewGridWithColumns(5,
		mono("Mutant"),
		mono("Class"),
		mono("Period"),
		mono("Alive"),
		mono("Dev"),
	)

	var table *widget.Table
	table = widget.NewTable(
		func() (int, int) { return len(u.mutResults), 5 },
		func() fyne.CanvasObject { return widget.NewLabel("placeholder") },
		func(id widget.TableCellID, o fyne.CanvasObject) {
			lbl := o.(*widget.Label)
			if id.Row < 0 || id.Row >= len(u.mutResults) {
				lbl.SetText("")
				return
			}
			r := u.mutResults[id.Row]
			switch id.Col {
			case 0:
				lbl.SetText(fmt.Sprintf("#%d", r.ID))
			case 1:
				lbl.SetText(string(r.Class))
			case 2:
				if r.Period > 0 {
					lbl.SetText(fmt.Sprintf("%.0f", r.Period))
				} else {
					lbl.SetText("—")
				}
			case 3:
				lbl.SetText(fmt.Sprintf("%.1f", r.Alive))
			case 4:
				lbl.SetText(fmt.Sprintf("%.1f%%", r.Deviation*100))
			}
		},
	)
	table.SetColumnWidth(0, 80)
	table.SetColumnWidth(1, 100)
	table.SetColumnWidth(2, 80)
	table.SetColumnWidth(3, 80)
	table.SetColumnWidth(4, 100)
	table.OnSelected = func(id widget.TableCellID) {
		if id.Row < 0 || id.Row >= len(u.mutResults) {
			u.mutSelected = -1
			return
		}
		u.mutSelected = id.Row
	}
	u.mutTable = table

	runBtn := widget.NewButton("Run tournament", func() {
		p := u.robustnessTargetPreset()
		if p == nil {
			dialog.ShowInformation("Mutation tournament", "Pin preset A or select one.", u.window)
			return
		}
		n := 20
		if v := strings.TrimSpace(u.mutCount.Text); v != "" {
			if nv := parseIntOr(v, 20); nv > 0 && nv <= 200 {
				n = nv
			}
		}
		frac := 0.20
		if v := strings.TrimSpace(u.mutFrac.Text); v != "" {
			if fv := parseFloatOr(v, 0.20); fv > 0 && fv <= 1.0 {
				frac = fv
			}
		}
		u.mutTarget.SetText(p.Name)
		u.mutStatus.SetText(fmt.Sprintf("Running %d mutants (±%.0f%%)…", n, frac*100))
		go func() {
			res := u.runMutationTournament(p, n, frac, 1)
			sort.SliceStable(res, func(i, j int) bool {
				if res[i].Class != res[j].Class {
					return res[i].Class < res[j].Class
				}
				return res[i].Deviation < res[j].Deviation
			})
			fyne.Do(func() {
				u.mutResults = res
				u.mutCv.Refresh()
				u.mutTable.Refresh()
				u.mutStatus.SetText(u.tournamentSummary(res, frac))
			})
		}()
	})

	loadPresetBtn := widget.NewButton("Save selected mutant as preset…", func() {
		if u.mutSelected < 0 || u.mutSelected >= len(u.mutResults) {
			dialog.ShowInformation("Save", "Select a row first.", u.window)
			return
		}
		mp := u.mutResults[u.mutSelected].Preset
		if err := saveUserPreset(mp); err != nil {
			dialog.ShowError(err, u.window)
			return
		}
		u.refreshPresetsList()
		dialog.ShowInformation("Saved", mp.Name, u.window)
	})
	pinABtn := widget.NewButton("Set as Pin A", func() {
		if u.mutSelected < 0 || u.mutSelected >= len(u.mutResults) {
			return
		}
		mp := u.mutResults[u.mutSelected].Preset
		u.presetPinA = mp
		if u.presetPinALabel != nil {
			u.presetPinALabel.SetText("A: " + mp.Name)
		}
	})
	pinBBtn := widget.NewButton("Set as Pin B", func() {
		if u.mutSelected < 0 || u.mutSelected >= len(u.mutResults) {
			return
		}
		mp := u.mutResults[u.mutSelected].Preset
		u.presetPinB = mp
		if u.presetPinBLabel != nil {
			u.presetPinBLabel.SetText("B: " + mp.Name)
		}
	})

	header := container.NewVBox(
		container.NewHBox(
			widget.NewLabelWithStyle("Mutation Tournament", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			widget.NewLabel("Target:"), u.mutTarget,
		),
		container.NewHBox(
			widget.NewLabel("Mutants:"), u.mutCount,
			widget.NewLabel("Perturbation ±:"), u.mutFrac,
			runBtn,
		),
		widget.NewSeparator(),
		classLegendRow(),
		u.mutStatus,
		widget.NewSeparator(),
		headers,
	)

	left := container.NewBorder(header, nil, nil, nil, table)

	plotPanel := container.NewVBox(
		widget.NewLabelWithStyle("Class distribution", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		u.mutCv,
		widget.NewSeparator(),
		container.NewHBox(loadPresetBtn, pinABtn, pinBBtn),
	)

	split := container.NewVSplit(left, plotPanel)
	split.SetOffset(0.6)
	return split
}

func (u *UI) tournamentSummary(results []mutationResult, frac float64) string {
	if len(results) == 0 {
		return "No results."
	}
	counts := map[BehaviorClass]int{}
	for _, r := range results {
		counts[r.Class]++
	}
	var sb strings.Builder
	fmt.Fprintf(&sb, "%d mutants (±%.0f%%)  ·  ", len(results), frac*100)
	first := true
	for _, cls := range classOrder {
		if n := counts[cls]; n > 0 {
			if !first {
				sb.WriteString("  ")
			}
			fmt.Fprintf(&sb, "%s=%d", cls, n)
			first = false
		}
	}
	return sb.String()
}

// ==================== Parse helpers ====================

func parseIntOr(s string, def int) int {
	var n int
	if _, err := fmt.Sscanf(s, "%d", &n); err != nil {
		return def
	}
	return n
}

func parseFloatOr(s string, def float64) float64 {
	var f float64
	if _, err := fmt.Sscanf(s, "%f", &f); err != nil {
		return def
	}
	return f
}

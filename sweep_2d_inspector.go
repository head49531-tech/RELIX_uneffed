package main

import (
	"fmt"
	"strconv"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// robustness2DSnapshot captures everything needed to reconstruct and inspect
// a completed 2D sweep. Stored on UI so cell selection can regenerate any
// individual preset on demand.
type robustness2DSnapshot struct {
	base     *DNAPreset
	px, py   string
	xMin     float64
	xMax     float64
	yMin     float64
	yMax     float64
	steps    int
	grid     [][]BehaviorClass
}

// classAbbrev returns a one-character badge for a behavior class.
func classAbbrev(c BehaviorClass) string {
	switch c {
	case BehOscillating:
		return "O"
	case BehBistable:
		return "T"
	case BehStable:
		return "S"
	case BehChaotic:
		return "C"
	case BehGrowing:
		return "G"
	case BehExtinct:
		return "X"
	}
	return "?"
}

// ==================== 2D parallel sweep ====================

func (u *UI) runInteractive2DSweep(p *DNAPreset, px, py string,
	xMin, xMax, yMin, yMax float64, steps int) [][]BehaviorClass {

	grid := make([][]BehaviorClass, steps)
	for i := range grid {
		grid[i] = make([]BehaviorClass, steps)
	}

	type job struct{ i, j int; x, y float64 }
	jobs := make(chan job, steps*steps)
	for i := 0; i < steps; i++ {
		for j := 0; j < steps; j++ {
			x := xMin + (xMax-xMin)*float64(i)/float64(steps-1)
			y := yMin + (yMax-yMin)*float64(j)/float64(steps-1)
			jobs <- job{i, j, x, y}
		}
	}
	close(jobs)

	var wg sync.WaitGroup
	var mu sync.Mutex
	sem := make(chan struct{}, 4)
	for jb := range jobs {
		jb := jb
		wg.Add(1)
		sem <- struct{}{}
		go func() {
			defer wg.Done()
			defer func() { <-sem }()
			np := clonePresetForSweep(p)
			params := classificationParams()
			np, params = applyRobustnessTo(np, px, jb.x, params)
			np, params = applyRobustnessTo(np, py, jb.y, params)
			seed := int64(jb.i*steps + jb.j + 1)
			rep := ClassifyPresetWithParams(np, params, 300, seed)
			mu.Lock()
			grid[jb.i][jb.j] = rep.Class
			mu.Unlock()
		}()
	}
	wg.Wait()
	return grid
}

// modifiedPresetFromCell reconstructs the exact preset that was used to
// classify cell (i,j) of the last 2D sweep.
func (u *UI) modifiedPresetFromCell(i, j int) *DNAPreset {
	snap := u.rmapSnapshot
	if snap == nil || snap.base == nil {
		return nil
	}
	if i < 0 || i >= snap.steps || j < 0 || j >= snap.steps {
		return nil
	}
	x := snap.xMin + (snap.xMax-snap.xMin)*float64(i)/float64(snap.steps-1)
	y := snap.yMin + (snap.yMax-snap.yMin)*float64(j)/float64(snap.steps-1)

	np := clonePresetForSweep(snap.base)
	params := classificationParams()
	np, params = applyRobustnessTo(np, snap.px, x, params)
	np, params = applyRobustnessTo(np, snap.py, y, params)

	np.Name = fmt.Sprintf("%s (cell %d,%d)", snap.base.Name, i, j)
	np.Description = fmt.Sprintf("Derived from %s at %s=%.3f, %s=%.3f",
		snap.base.Name, snap.px, x, snap.py, y)
	np.Builtin = false
	return np
}

// ==================== Interactive tab ====================

func (u *UI) buildInteractiveRobustnessMapTab() fyne.CanvasObject {
	u.rmapParamX = widget.NewSelect(robustnessParams, nil)
	u.rmapParamX.SetSelected("TranslationK")
	u.rmapParamY = widget.NewSelect(robustnessParams, nil)
	u.rmapParamY.SetSelected("HillN")
	u.rmapXMin = widget.NewEntry(); u.rmapXMin.SetText("0.5")
	u.rmapXMax = widget.NewEntry(); u.rmapXMax.SetText("3.0")
	u.rmapYMin = widget.NewEntry(); u.rmapYMin.SetText("1.0")
	u.rmapYMax = widget.NewEntry(); u.rmapYMax.SetText("4.0")
	u.rmapSteps = widget.NewEntry(); u.rmapSteps.SetText("10")
	u.rmapStatus = mono("—")
	u.rmapTarget = widget.NewLabel("(pin A or select a preset)")
	u.rmapSelInfo = mono("Run a 2D sweep, then click a cell to inspect.")
	u.rmapInspector = container.NewStack(widget.NewLabel("Run 2D sweep to populate grid."))

	loadABtn := widget.NewButton("Set cell as Pin A", func() {
		if u.rmapSelectedCell[0] < 0 {
			return
		}
		np := u.modifiedPresetFromCell(u.rmapSelectedCell[0], u.rmapSelectedCell[1])
		if np == nil {
			return
		}
		u.presetPinA = np
		if u.presetPinALabel != nil {
			u.presetPinALabel.SetText("A: " + np.Name)
		}
		dialog.ShowInformation("Pin A", np.Name, u.window)
	})
	loadBBtn := widget.NewButton("Set cell as Pin B", func() {
		if u.rmapSelectedCell[0] < 0 {
			return
		}
		np := u.modifiedPresetFromCell(u.rmapSelectedCell[0], u.rmapSelectedCell[1])
		if np == nil {
			return
		}
		u.presetPinB = np
		if u.presetPinBLabel != nil {
			u.presetPinBLabel.SetText("B: " + np.Name)
		}
		dialog.ShowInformation("Pin B", np.Name, u.window)
	})
	savePresetBtn := widget.NewButton("Save cell as preset…", func() {
		if u.rmapSelectedCell[0] < 0 {
			return
		}
		np := u.modifiedPresetFromCell(u.rmapSelectedCell[0], u.rmapSelectedCell[1])
		if np == nil {
			return
		}
		if err := saveUserPreset(np); err != nil {
			dialog.ShowError(err, u.window)
			return
		}
		u.refreshPresetsList()
		dialog.ShowInformation("Saved", np.Name, u.window)
	})

	runBtn := widget.NewButton("Run 2D sweep", func() {
		p := u.robustnessTargetPreset()
		if p == nil {
			dialog.ShowInformation("Robustness map", "Pin preset A or select one.", u.window)
			return
		}
		px := u.rmapParamX.Selected
		py := u.rmapParamY.Selected
		xMin, e1 := strconv.ParseFloat(u.rmapXMin.Text, 64)
		xMax, e2 := strconv.ParseFloat(u.rmapXMax.Text, 64)
		yMin, e3 := strconv.ParseFloat(u.rmapYMin.Text, 64)
		yMax, e4 := strconv.ParseFloat(u.rmapYMax.Text, 64)
		steps, e5 := strconv.Atoi(u.rmapSteps.Text)
		if e1 != nil || e2 != nil || e3 != nil || e4 != nil || e5 != nil ||
			steps < 3 || xMax <= xMin || yMax <= yMin {
			u.rmapStatus.SetText("Bad ranges.")
			return
		}
		if steps > 15 {
			steps = 15
		}
		u.rmapTarget.SetText(p.Name)
		u.rmapStatus.SetText(fmt.Sprintf("Sweeping %dx%d grid…", steps, steps))
		go func() {
			grid := u.runInteractive2DSweep(p, px, py, xMin, xMax, yMin, yMax, steps)
			fyne.Do(func() {
				u.rmapSnapshot = &robustness2DSnapshot{
					base: p, px: px, py: py,
					xMin: xMin, xMax: xMax, yMin: yMin, yMax: yMax,
					steps: steps, grid: grid,
				}
				u.rmapSelectedCell = [2]int{-1, -1}
				u.rmapSelInfo.SetText("Click a cell to inspect.")
				u.rebuild2DInspector()
				u.rmapStatus.SetText(fmt.Sprintf("Done — %dx%d classified. %s → %s, %s → %s",
					steps, steps, px, fmtRange(xMin, xMax), py, fmtRange(yMin, yMax)))
			})
		}()
	})

	header := container.NewVBox(
		container.NewHBox(
			widget.NewLabelWithStyle("2D Sweep Inspector", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			widget.NewLabel("Target:"), u.rmapTarget,
		),
		container.NewHBox(
			widget.NewLabel("X:"), u.rmapParamX,
			widget.NewLabel("Min:"), u.rmapXMin,
			widget.NewLabel("Max:"), u.rmapXMax,
			widget.NewSeparator(),
			widget.NewLabel("Y:"), u.rmapParamY,
			widget.NewLabel("Min:"), u.rmapYMin,
			widget.NewLabel("Max:"), u.rmapYMax,
			widget.NewLabel("Steps:"), u.rmapSteps,
			runBtn,
		),
		classLegendRow(),
		u.rmapStatus,
		widget.NewSeparator(),
	)

	footer := container.NewVBox(
		widget.NewSeparator(),
		u.rmapSelInfo,
		container.NewHBox(loadABtn, loadBBtn, savePresetBtn),
	)

	scroll := container.NewScroll(u.rmapInspector)
	scroll.SetMinSize(fyne.NewSize(700, 500))
	return container.NewBorder(header, footer, nil, nil, scroll)
}

func fmtRange(min, max float64) string {
	return fmt.Sprintf("[%.2f, %.2f]", min, max)
}

// rebuild2DInspector replaces the grid of cells with buttons matching the
// current snapshot. Each button, when clicked, selects that cell.
func (u *UI) rebuild2DInspector() {
	if u.rmapInspector == nil || u.rmapSnapshot == nil {
		return
	}
	snap := u.rmapSnapshot
	grid := container.NewGridWithColumns(snap.steps)
	for i := 0; i < snap.steps; i++ {
		for j := 0; j < snap.steps; j++ {
			i, j := i, j
			cls := snap.grid[i][j]
			label := classAbbrev(cls)
			btn := widget.NewButton(label, func() {
				u.select2DCell(i, j)
			})
			btn.Importance = widget.LowImportance
			grid.Add(btn)
		}
	}
	u.rmapInspector.Objects = []fyne.CanvasObject{grid}
	u.rmapInspector.Refresh()
}

func (u *UI) select2DCell(i, j int) {
	snap := u.rmapSnapshot
	if snap == nil {
		return
	}
	if i < 0 || i >= snap.steps || j < 0 || j >= snap.steps {
		return
	}
	x := snap.xMin + (snap.xMax-snap.xMin)*float64(i)/float64(snap.steps-1)
	y := snap.yMin + (snap.yMax-snap.yMin)*float64(j)/float64(snap.steps-1)
	cls := snap.grid[i][j]
	u.rmapSelectedCell = [2]int{i, j}
	u.rmapSelInfo.SetText(fmt.Sprintf(
		"Cell (%d,%d)  ·  %s = %.3f  ·  %s = %.3f  ·  class = %s",
		i, j, snap.px, x, snap.py, y, cls,
	))
}

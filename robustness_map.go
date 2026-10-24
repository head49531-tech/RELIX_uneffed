package main

import (
	"fmt"
	"image"
	"image/color"
	"strconv"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// ==================== 2D sweep runner ====================

func (u *UI) runRobustness2D(p *DNAPreset, px, py string,
	xMin, xMax, yMin, yMax float64, steps int) [][]BehaviorClass {

	grid := make([][]BehaviorClass, steps)
	for i := range grid {
		grid[i] = make([]BehaviorClass, steps)
	}

	type job struct {
		i, j int
		x, y float64
	}
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

// ==================== 2D renderer ====================

func renderRobustness2D(grid [][]BehaviorClass, w, h int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	bg := color.RGBA{13, 13, 16, 255}
	for i := 0; i < w*h*4; i += 4 {
		img.Pix[i] = bg.R
		img.Pix[i+1] = bg.G
		img.Pix[i+2] = bg.B
		img.Pix[i+3] = bg.A
	}

	rows := len(grid)
	if rows == 0 {
		return img
	}
	cols := len(grid[0])
	if cols == 0 {
		return img
	}

	margin := 24
	plotW := w - 2*margin
	plotH := h - 2*margin
	if plotW < cols || plotH < rows {
		return img
	}
	cellW := plotW / cols
	cellH := plotH / rows
	if cellW < 1 {
		cellW = 1
	}
	if cellH < 1 {
		cellH = 1
	}

	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			x0 := margin + j*cellW
			y0 := margin + (rows-1-i)*cellH
			c := classColor(grid[i][j])
			for dx := 0; dx < cellW; dx++ {
				for dy := 0; dy < cellH; dy++ {
					px, py := x0+dx, y0+dy
					if px >= 0 && px < w && py >= 0 && py < h {
						img.Set(px, py, c)
					}
				}
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

// ==================== 2D tab ====================

func (u *UI) buildRobustnessMapTabStatic() fyne.CanvasObject {
	u.rmapParamX = widget.NewSelect(robustnessParams, nil)
	u.rmapParamX.SetSelected("TranslationK")
	u.rmapParamY = widget.NewSelect(robustnessParams, nil)
	u.rmapParamY.SetSelected("HillN")
	u.rmapXMin = widget.NewEntry()
	u.rmapXMin.SetText("0.5")
	u.rmapXMax = widget.NewEntry()
	u.rmapXMax.SetText("3.0")
	u.rmapYMin = widget.NewEntry()
	u.rmapYMin.SetText("1.0")
	u.rmapYMax = widget.NewEntry()
	u.rmapYMax.SetText("4.0")
	u.rmapSteps = widget.NewEntry()
	u.rmapSteps.SetText("10")
	u.rmapStatus = mono("—")
	u.rmapTarget = widget.NewLabel("(pin A or select a preset)")

	u.rmapCv = canvas.NewRaster(func(w, h int) image.Image {
		return renderRobustness2D(u.rmapGrid, w, h)
	})
	u.rmapCv.SetMinSize(fyne.NewSize(700, 500))

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
		u.rmapTarget.SetText(p.Name)
		u.rmapStatus.SetText(fmt.Sprintf("Sweeping %dx%d grid…", steps, steps))
		go func() {
			grid := u.runRobustness2D(p, px, py, xMin, xMax, yMin, yMax, steps)
			fyne.Do(func() {
				u.rmapGrid = grid
				u.rmapCv.Refresh()
				u.rmapStatus.SetText(fmt.Sprintf("Done — %dx%d grid classified.", steps, steps))
			})
		}()
	})

	header := container.NewVBox(
		container.NewHBox(
			widget.NewLabelWithStyle("Robustness Map (2D)", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
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
	return container.NewBorder(header, nil, nil, nil, u.rmapCv)
}


// buildRobustnessMapTab delegates to the interactive inspector.
func (u *UI) buildRobustnessMapTab() fyne.CanvasObject {
	return u.buildInteractiveRobustnessMapTab()
}

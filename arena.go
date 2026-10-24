package main

import (
	"fmt"
	"image"
	"math"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// arenaState is an isolated side-by-side sandbox holding exactly two
// cells, one cloned from preset A and one from preset B. All species
// are prefixed A_ / B_ so a single Sim can host both without name
// collisions, and so the existing renderOscilloscope works unchanged.
type arenaState struct {
	sim    *Sim
	cellA  *Cell
	cellB  *Cell
	labelA string
	labelB string
	ticks  int
}

func arenaParams() SimParams {
	p := DefaultParams()
	p.MaxCells = 2
	p.BaseDeathRate = 0
	p.MutationRate = 0
	p.DivisionATP = 1e9 // effectively disable division for clean A/B compare
	return p
}

// prefixCellSpecies renames genes, products, MRNA keys, and Protein keys
// so A and B can share one Sim. Inducers and quorum signals are global
// and intentionally left alone (AHL cross-talk is a feature).
func prefixCellSpecies(c *Cell, prefix string) {
	// Build set of names that will be prefixed so pathway references to
	// proteins get renamed but references to real metabolites like
	// "Glucose" or "ATP" (which live in c.Metabolites, not c.Proteins) are
	// left alone. Without this, Arena silently disables all pathways.
	nameSet := make(map[string]bool, 2*len(c.Genome.Genes))
	for _, g := range c.Genome.Genes {
		nameSet[g.Name] = true
		nameSet[g.Product] = true
	}
	for _, g := range c.Genome.Genes {
		g.Name = prefix + g.Name
		g.Product = prefix + g.Product
		for i := range g.Activators {
			g.Activators[i] = prefix + g.Activators[i]
		}
		for i := range g.Repressors {
			g.Repressors[i] = prefix + g.Repressors[i]
		}
	}
	for i := range c.Genome.Pathways {
		if nameSet[c.Genome.Pathways[i].Enzyme] {
			c.Genome.Pathways[i].Enzyme = prefix + c.Genome.Pathways[i].Enzyme
		}
		if nameSet[c.Genome.Pathways[i].Substrate] {
			c.Genome.Pathways[i].Substrate = prefix + c.Genome.Pathways[i].Substrate
		}
		if nameSet[c.Genome.Pathways[i].Product] {
			c.Genome.Pathways[i].Product = prefix + c.Genome.Pathways[i].Product
		}
	}
	newM := map[string]float64{}
	for k, v := range c.MRNA {
		newM[prefix+k] = v
	}
	c.MRNA = newM
	newP := map[string]float64{}
	for k, v := range c.Proteins {
		newP[prefix+k] = v
	}
	c.Proteins = newP
}

// ==================== Setup / reset ====================

func (u *UI) setupArena() {
	a, b := u.presetPinA, u.presetPinB
	if a == nil || b == nil {
		dialog.ShowInformation("Arena", "Pin preset A and B first (Presets tab).", u.window)
		return
	}
	if len(a.Genes) == 0 || len(b.Genes) == 0 {
		dialog.ShowInformation("Arena", "Both presets need at least one gene.", u.window)
		return
	}

	s := NewSim(arenaParams(), time.Now().UnixNano())
	s.Cells = nil
	s.NextID = 1

	ca := newCellFromPreset(s, a)
	for k, v := range a.InitMRNA {
		ca.MRNA[k] = v
	}
	for k, v := range a.InitProteins {
		ca.Proteins[k] = v
	}
	prefixCellSpecies(ca, "A_")

	cb := newCellFromPreset(s, b)
	for k, v := range b.InitMRNA {
		cb.MRNA[k] = v
	}
	for k, v := range b.InitProteins {
		cb.Proteins[k] = v
	}
	prefixCellSpecies(cb, "B_")

	s.Cells = []*Cell{ca, cb}

	xname := ""
	yname := ""
	if a.Genes[0].Product != "" {
		xname = "A_" + a.Genes[0].Product
	}
	if b.Genes[0].Product != "" {
		yname = "B_" + b.Genes[0].Product
	}

	u.arena = &arenaState{
		sim: s, cellA: ca, cellB: cb,
		labelA: a.Name, labelB: b.Name,
	}
	u.arenaXName = xname
	u.arenaYName = yname

	// Prime the trace so the user sees something immediately
	u.arenaStep(100)
}

// ==================== Running ====================

func (u *UI) arenaStep(n int) {
	if u.arena == nil || u.arenaRunning {
		return
	}
	for i := 0; i < n; i++ {
		u.arena.sim.Step()
		u.arena.ticks++
	}
	u.refreshArena()
}

func (u *UI) arenaRun(ticks int) {
	if u.arena == nil || u.arenaRunning {
		return
	}
	u.arenaRunning = true
	go func() {
		const batch = 10
		for t := 0; t < ticks; t += batch {
			n := batch
			if t+n > ticks {
				n = ticks - t
			}
			fyne.Do(func() {
				for i := 0; i < n; i++ {
					u.arena.sim.Step()
					u.arena.ticks++
				}
				u.refreshArena()
			})
			time.Sleep(time.Millisecond)
		}
		fyne.Do(func() { u.arenaRunning = false })
	}()
}

// ==================== Refresh / stats ====================

func (u *UI) refreshArena() {
	if u.arenaCv == nil {
		return
	}
	if u.arena == nil {
		if u.arenaTicksLabel != nil {
			u.arenaTicksLabel.SetText("ticks: 0")
		}
		if u.arenaValA != nil {
			u.arenaValA.SetText("A: (not loaded)")
		}
		if u.arenaValB != nil {
			u.arenaValB.SetText("B: (not loaded)")
		}
		if u.arenaStats != nil {
			u.arenaStats.SetText("")
		}
		u.arenaCv.Refresh()
		return
	}

	if u.arenaTicksLabel != nil {
		u.arenaTicksLabel.SetText(fmt.Sprintf("ticks: %d", u.arena.ticks))
	}

	var lastX, lastY float64
	if len(u.arena.sim.SpeciesHistory) > 0 {
		last := u.arena.sim.SpeciesHistory[len(u.arena.sim.SpeciesHistory)-1]
		if u.arenaXName != "" {
			lastX = last.Proteins[u.arenaXName]
		}
		if u.arenaYName != "" {
			lastY = last.Proteins[u.arenaYName]
		}
	}
	if u.arenaValA != nil && u.arenaXName != "" {
		u.arenaValA.SetText(fmt.Sprintf("%s = %.3f", u.arenaXName, lastX))
	}
	if u.arenaValB != nil && u.arenaYName != "" {
		u.arenaValB.SetText(fmt.Sprintf("%s = %.3f", u.arenaYName, lastY))
	}
	if u.arenaStats != nil {
		u.arenaStats.SetText(u.arenaStatText())
	}
	u.arenaCv.Refresh()
}

func (u *UI) arenaStatText() string {
	if u.arena == nil {
		return ""
	}
	var sb strings.Builder
	fmt.Fprintf(&sb, "%-24s %-10s %-8s %-8s %-8s %-8s\n",
		"Species", "Mean", "CV", "Min", "Max", "Period")
	fmt.Fprintf(&sb, "%s\n", strings.Repeat("-", 74))
	shown := 0
	for _, name := range []string{u.arenaXName, u.arenaYName} {
		if name == "" {
			continue
		}
		m, cv, mn, mx, period := u.arenaSeriesStats(name, 200)
		periodStr := "—"
		if period > 0 {
			periodStr = fmt.Sprintf("%.0ft", period)
		}
		fmt.Fprintf(&sb, "%-24s %-10.3f %-8.3f %-8.2f %-8.2f %-8s\n",
			truncateName(name, 24), m, cv, mn, mx, periodStr)
		shown++
	}
	if shown == 0 {
		sb.WriteString("(no species)\n")
	}
	return sb.String()
}

func (u *UI) arenaSeriesStats(name string, window int) (mean, cv, mn, mx, period float64) {
	if u.arena == nil || name == "" {
		return
	}
	hist := u.arena.sim.SpeciesHistory
	if len(hist) == 0 {
		return
	}
	start := 0
	if len(hist) > window {
		start = len(hist) - window
	}
	xs := make([]float64, 0, len(hist)-start)
	for _, sp := range hist[start:] {
		xs = append(xs, sp.Proteins[name])
	}
	if len(xs) == 0 {
		return
	}
	mn, mx = xs[0], xs[0]
	sum := 0.0
	for _, v := range xs {
		sum += v
		if v < mn {
			mn = v
		}
		if v > mx {
			mx = v
		}
	}
	mean = sum / float64(len(xs))
	var ss float64
	for _, v := range xs {
		d := v - mean
		ss += d * d
	}
	if mean > 1e-9 && len(xs) > 1 {
		cv = math.Sqrt(ss/float64(len(xs))) / mean
	}
	if len(xs) >= 12 {
		period = autocorrPeriod(xs, 3, len(xs)/2)
	}
	return
}

// ==================== Tab UI ====================

func (u *UI) buildArenaTab() fyne.CanvasObject {
	u.arenaValA = mono("A: (not loaded)")
	u.arenaValB = mono("B: (not loaded)")
	u.arenaTicksLabel = mono("ticks: 0")
	u.arenaStats = mono("")

	swatchA := canvas.NewRectangle(scopeColorX)
	swatchA.SetMinSize(fyne.NewSize(12, 12))
	swatchB := canvas.NewRectangle(scopeColorY)
	swatchB.SetMinSize(fyne.NewSize(12, 12))

	loadBtn := widget.NewButton("Load A vs B", u.setupArena)
	run500 := widget.NewButton("Run 500", func() { u.arenaRun(500) })
	run1000 := widget.NewButton("Run 1000", func() { u.arenaRun(1000) })
	step1 := widget.NewButton("Step 1", func() { u.arenaStep(1) })
	step10 := widget.NewButton("Step 10", func() { u.arenaStep(10) })
	resetBtn := widget.NewButton("Reset", u.setupArena)

	header := container.NewVBox(
		container.NewHBox(
			widget.NewLabelWithStyle("Arena", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			widget.NewLabel("— head-to-head preset sandbox (pin A and B in Presets tab)"),
			loadBtn,
		),
		container.NewHBox(run500, run1000, widget.NewSeparator(), step1, step10, widget.NewSeparator(), resetBtn, widget.NewSeparator(), u.arenaTicksLabel),
		container.NewHBox(swatchA, u.arenaValA, widget.NewLabel("    "), swatchB, u.arenaValB),
		widget.NewSeparator(),
	)

	u.arenaCv = canvas.NewRaster(func(w, h int) image.Image {
		if u.arena == nil {
			blank := image.NewRGBA(image.Rect(0, 0, w, h))
			for i := 0; i < w*h*4; i += 4 {
				blank.Pix[i] = 13
				blank.Pix[i+1] = 13
				blank.Pix[i+2] = 16
				blank.Pix[i+3] = 255
			}
			return blank
		}
		return renderOscilloscope(u.arena.sim, u.arenaXName, u.arenaYName, w, h)
	})
	u.arenaCv.SetMinSize(fyne.NewSize(900, 420))

	statsPanel := container.NewVScroll(u.arenaStats)
	statsPanel.SetMinSize(fyne.NewSize(900, 90))

	return container.NewBorder(header, statsPanel, nil, nil, u.arenaCv)
}

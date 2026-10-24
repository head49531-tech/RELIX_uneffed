package main

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// ==================== Behavior classification ====================

type BehaviorClass string

const (
	BehUnknown     BehaviorClass = "?"
	BehStable      BehaviorClass = "stable"
	BehOscillating BehaviorClass = "osc"
	BehBistable    BehaviorClass = "tog"
	BehChaotic     BehaviorClass = "chaotic"
	BehExtinct     BehaviorClass = "extinct"
	BehGrowing     BehaviorClass = "growing"
)

type SpeciesBehavior struct {
	Name   string
	Mean   float64
	CV     float64
	Fano   float64
	Period float64
}

type BehaviorReport struct {
	Preset  string
	Class   BehaviorClass
	Period  float64
	Alive   float64
	PopCV   float64
	Species []SpeciesBehavior
	Ticks   int
	Err     string
}

func (b *BehaviorReport) ShortLabel() string {
	if b == nil {
		return ""
	}
	switch b.Class {
	case BehOscillating:
		if b.Period > 0 {
			return fmt.Sprintf("osc %dt", int(b.Period))
		}
		return "osc"
	case BehBistable:
		return "tog"
	case BehStable:
		return "stable"
	case BehChaotic:
		return "chaotic"
	case BehExtinct:
		return "extinct"
	case BehGrowing:
		return "growing"
	}
	return "?"
}

// ==================== Headless preset harness ====================

func classificationParams() SimParams {
	p := DefaultParams()
	p.MaxCells = 1
	p.BaseDeathRate = 0
	p.MutationRate = 0
	return p
}

func newCellFromPreset(s *Sim, p *DNAPreset) *Cell {
	c := &Cell{
		ID:          s.NextID,
		Generation:  0,
		Alive:       true,
		Genome:      p.genomeFromPreset(),
		MRNA:        map[string]float64{},
		Proteins:    map[string]float64{},
		Metabolites: map[string]float64{},
		Glucose:     10,
		ATP:         5,
	}
	s.NextID++
	return c
}

func simFromPreset(p *DNAPreset, seed int64, initState bool) *Sim {
	s := NewSim(classificationParams(), seed)
	s.Cells = nil
	s.NextID = 1
	c := newCellFromPreset(s, p)
	if initState {
		for k, v := range p.InitMRNA {
			c.MRNA[k] = v
		}
		for k, v := range p.InitProteins {
			c.Proteins[k] = v
		}
	}
	s.Cells = append(s.Cells, c)
	return s
}

// ==================== Main classifier ====================

func ClassifyPreset(p *DNAPreset, ticks int, seed int64) *BehaviorReport {
	return classifyBody(p, classificationParams(), ticks, seed, true)
}

func ClassifyPresetWithParams(p *DNAPreset, params SimParams, ticks int, seed int64) *BehaviorReport {
	return classifyBody(p, params, ticks, seed, false)
}

func classifyBody(p *DNAPreset, params SimParams, ticks int, seed int64, doBistable bool) *BehaviorReport {
	rep := &BehaviorReport{Preset: p.Name, Ticks: ticks}
	if p == nil {
		rep.Class = BehUnknown
		return rep
	}
	s := NewSim(params, seed)
	s.Cells = nil
	s.NextID = 1
	c := newCellFromPreset(s, p)
	for k, v := range p.InitMRNA {
		c.MRNA[k] = v
	}
	for k, v := range p.InitProteins {
		c.Proteins[k] = v
	}
	s.Cells = append(s.Cells, c)
	hist := map[string][]float64{}
	for t := 0; t < ticks; t++ {
		s.Step()
		if s.AliveCount() == 0 {
			rep.Class = BehExtinct
			rep.Err = fmt.Sprintf("extinct at t=%d", t)
			return rep
		}
		for k, v := range s.GlobalProteinLevels() {
			hist[k] = append(hist[k], v)
		}
	}

	window := ticks / 4
	if window < 20 {
		window = 20
	}
	if len(s.History) < window {
		window = len(s.History)
	}
	if window < 5 {
		rep.Class = BehUnknown
		rep.Err = "run too short"
		return rep
	}

	for name, vals := range hist {
		if len(vals) < window {
			continue
		}
		tail := vals[len(vals)-window:]
		m, cv := meanCV(tail)
		if m < 0.05 {
			continue
		}
		sb := SpeciesBehavior{
			Name: name, Mean: m, CV: cv,
			Fano: variance(tail) / math.Max(m, 1e-9),
		}
		sb.Period = autocorrPeriod(tail, 3, window/2)
		rep.Species = append(rep.Species, sb)
	}
	sort.Slice(rep.Species, func(i, j int) bool {
		return rep.Species[i].Mean > rep.Species[j].Mean
	})

	popTail := make([]float64, 0, window)
	for i := len(s.History) - window; i < len(s.History); i++ {
		popTail = append(popTail, float64(s.History[i].Alive))
	}
	rep.Alive, rep.PopCV = meanCV(popTail)

	for _, sb := range rep.Species {
		if sb.Period > 0 && sb.CV > 0.15 {
			rep.Class = BehOscillating
			rep.Period = sb.Period
			return rep
		}
	}

	if doBistable && detectBistable(p, seed+1) {
		rep.Class = BehBistable
		return rep
	}

	for _, sb := range rep.Species {
		if sb.CV > 0.3 {
			rep.Class = BehChaotic
			return rep
		}
	}

	if len(s.History) >= 10 {
		early := float64(s.History[len(s.History)-window].Alive)
		late := float64(s.History[len(s.History)-1].Alive)
		if late > early*1.5 && late > 3 {
			rep.Class = BehGrowing
			return rep
		}
	}

	rep.Class = BehStable
	return rep
}

// ==================== Bistability detection ====================

func detectBistable(p *DNAPreset, seed int64) bool {
	const N = 20
	const T = 300

	s := NewSim(classificationParams(), seed)
	s.Params.MaxCells = N
	s.Params.BaseDeathRate = 0
	s.Params.MutationRate = 0
	s.Cells = nil
	s.NextID = 1
	for i := 0; i < N; i++ {
		c := newCellFromPreset(s, p)
		for _, g := range c.Genome.Genes {
			c.MRNA[g.Name] = s.Rng.Float64() * 0.8
		}
		s.Cells = append(s.Cells, c)
	}

	for t := 0; t < T; t++ {
		s.Step()
	}

	speciesSet := map[string]bool{}
	alive := 0
	for _, c := range s.Cells {
		if !c.Alive {
			continue
		}
		alive++
		for name := range c.Proteins {
			speciesSet[name] = true
		}
	}
	if alive < 10 || len(speciesSet) < 2 {
		return false
	}
	species := make([]string, 0, len(speciesSet))
	for name := range speciesSet {
		species = append(species, name)
	}

	for i := 0; i < len(species); i++ {
		for j := i + 1; j < len(species); j++ {
			a, b := species[i], species[j]
			countA, countB := 0, 0
			for _, c := range s.Cells {
				if !c.Alive {
					continue
				}
				va, vb := c.Proteins[a], c.Proteins[b]
				if va > 2*vb && va > 0.5 {
					countA++
				}
				if vb > 2*va && vb > 0.5 {
					countB++
				}
			}
			fracA := float64(countA) / float64(alive)
			fracB := float64(countB) / float64(alive)
			if fracA >= 0.2 && fracB >= 0.2 {
				return true
			}
		}
	}
	return false
}

// ==================== Batch runner ====================

func ClassifyAllPresets(presets []*DNAPreset, ticks int) map[string]*BehaviorReport {
	out := map[string]*BehaviorReport{}
	var mu sync.Mutex
	var wg sync.WaitGroup
	sem := make(chan struct{}, 4)
	seed := int64(1)
	for _, p := range presets {
		p := p
		wg.Add(1)
		sem <- struct{}{}
		go func() {
			defer wg.Done()
			defer func() { <-sem }()
			rep := ClassifyPreset(p, ticks, seed)
			mu.Lock()
			out[p.Name] = rep
			mu.Unlock()
		}()
		seed++
	}
	wg.Wait()
	return out
}

// ==================== UI integration ====================

func (u *UI) presetBadge(name string) string {
	if u.behaviorCache == nil {
		return ""
	}
	rep, ok := u.behaviorCache[name]
	if !ok || rep == nil {
		return ""
	}
	return "[" + rep.ShortLabel() + "]"
}

func (u *UI) screenAllPresets() {
	if u.screeningActive {
		return
	}
	presets := append([]*DNAPreset{}, u.presetsCache...)
	if len(presets) == 0 {
		return
	}
	u.screeningActive = true
	dialog.ShowInformation("Screening",
		fmt.Sprintf("Classifying %d presets in background (600 ticks each)…", len(presets)),
		u.window)
	go func() {
		reports := ClassifyAllPresets(presets, 600)
		fyne.Do(func() {
			if u.behaviorCache == nil {
				u.behaviorCache = map[string]*BehaviorReport{}
			}
			for k, v := range reports {
				u.behaviorCache[k] = v
			}
			u.screeningActive = false
			if u.presetList != nil {
				u.presetList.Refresh()
			}
			u.refreshPresetPreview()
			u.showBehaviorSummary(reports)
		})
	}()
}

func (u *UI) showBehaviorSummaryText(reports map[string]*BehaviorReport) {
	type row struct {
		Name, Cat, Class, Period, Alive, PopCV string
	}
	var rows []row
	for _, p := range u.presetsCache {
		rep, ok := reports[p.Name]
		if !ok || rep == nil {
			continue
		}
		r := row{
			Name: p.Name, Cat: p.Category, Class: string(rep.Class),
			Alive: fmt.Sprintf("%.0f", rep.Alive),
			PopCV: fmt.Sprintf("%.2f", rep.PopCV),
		}
		if rep.Period > 0 {
			r.Period = fmt.Sprintf("%.0f", rep.Period)
		}
		rows = append(rows, r)
	}
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].Class != rows[j].Class {
			return rows[i].Class < rows[j].Class
		}
		return rows[i].Name < rows[j].Name
	})
	counts := map[string]int{}
	for _, r := range rows {
		counts[r.Class]++
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "Behavior Screen — %d presets classified\n\n", len(rows))
	order := []string{"osc", "tog", "stable", "growing", "chaotic", "extinct", "?"}
	for _, c := range order {
		if n, ok := counts[c]; ok && n > 0 {
			fmt.Fprintf(&sb, "  %-10s %d\n", c, n)
		}
	}
	for c := range counts {
		found := false
		for _, o := range order {
			if o == c {
				found = true
				break
			}
		}
		if !found {
			fmt.Fprintf(&sb, "  %-10s %d\n", c, counts[c])
		}
	}
	sb.WriteString("\n")
	fmt.Fprintf(&sb, "%-30s %-12s %-9s %-7s %-6s %s\n",
		"Name", "Category", "Class", "Period", "Alive", "PopCV")
	fmt.Fprintf(&sb, "%s\n", strings.Repeat("-", 78))
	for _, r := range rows {
		fmt.Fprintf(&sb, "%-30s %-12s %-9s %-7s %-6s %s\n",
			truncateName(r.Name, 30), r.Cat, r.Class, r.Period, r.Alive, r.PopCV)
	}

	ta := widget.NewMultiLineEntry()
	ta.Wrapping = fyne.TextWrapOff
	ta.TextStyle = fyne.TextStyle{Monospace: true}
	ta.SetText(sb.String())
	d := dialog.NewCustom("Behavior Screen", "Close", container.NewScroll(ta), u.window)
	d.Resize(fyne.NewSize(760, 600))
	d.Show()
}

func truncateName(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n-1] + "…"
}

// ==================== Stats helpers ====================

func meanCV(xs []float64) (float64, float64) {
	if len(xs) == 0 {
		return 0, 0
	}
	mean := 0.0
	for _, x := range xs {
		mean += x
	}
	mean /= float64(len(xs))
	if mean < 1e-9 {
		return 0, 0
	}
	ss := 0.0
	for _, x := range xs {
		d := x - mean
		ss += d * d
	}
	v := ss / float64(len(xs))
	return mean, math.Sqrt(v) / mean
}

func variance(xs []float64) float64 {
	if len(xs) < 2 {
		return 0
	}
	mean := 0.0
	for _, x := range xs {
		mean += x
	}
	mean /= float64(len(xs))
	ss := 0.0
	for _, x := range xs {
		d := x - mean
		ss += d * d
	}
	return ss / float64(len(xs))
}

func autocorrPeriod(xs []float64, minLag, maxLag int) float64 {
	n := len(xs)
	if n < 8 {
		return 0
	}
	if maxLag > n/2 {
		maxLag = n / 2
	}
	if maxLag < minLag+2 {
		return 0
	}
	mean := 0.0
	for _, x := range xs {
		mean += x
	}
	mean /= float64(n)
	dm := make([]float64, n)
	var0 := 0.0
	for i, x := range xs {
		dm[i] = x - mean
		var0 += dm[i] * dm[i]
	}
	if var0 < 1e-9 {
		return 0
	}
	crossed := false
	for lag := minLag; lag <= maxLag; lag++ {
		s := 0.0
		for i := 0; i+lag < n; i++ {
			s += dm[i] * dm[i+lag]
		}
		ac := s / var0
		if !crossed && ac < -0.1 {
			crossed = true
			continue
		}
		if crossed && ac > 0.4 {
			return float64(lag)
		}
	}
	return 0
}


// showBehaviorSummary routes to the interactive browser; the old plain-text
// version is preserved as showBehaviorSummaryText.
func (u *UI) showBehaviorSummary(reports map[string]*BehaviorReport) {
	u.showBehaviorBrowser(reports)
}

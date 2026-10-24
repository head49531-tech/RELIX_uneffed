package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

type sweepResult struct {
	ParamValue float64
	Alive      int
	TotalATP   float64
	MaxGen     int
	Steady     bool
}

func (u *UI) showSweepDialog() {
	paramSel := widget.NewSelect([]string{"division_atp", "mutation_rate", "translation_k", "base_death_rate", "env_glucose"}, nil)
	paramSel.SetSelected("division_atp")
	minIn := widget.NewEntry(); minIn.SetText("2")
	maxIn := widget.NewEntry(); maxIn.SetText("20")
	stepsIn := widget.NewEntry(); stepsIn.SetText("10")
	ticksIn := widget.NewEntry(); ticksIn.SetText("200")

	dialog.ShowForm("Parameter Sweep", "Run", "Cancel", []*widget.FormItem{
		{Text: "Parameter", Widget: paramSel},
		{Text: "Min", Widget: minIn}, {Text: "Max", Widget: maxIn},
		{Text: "Steps", Widget: stepsIn}, {Text: "Ticks per run", Widget: ticksIn},
	}, func(ok bool) {
		if !ok { return }
		minV, _ := strconv.ParseFloat(minIn.Text, 64)
		maxV, _ := strconv.ParseFloat(maxIn.Text, 64)
		steps, _ := strconv.Atoi(stepsIn.Text)
		ticks, _ := strconv.Atoi(ticksIn.Text)
		if steps < 2 || ticks < 10 {
			dialog.ShowError(fmt.Errorf("need steps>=2, ticks>=10"), u.window)
			return
		}
		go u.runSweep(paramSel.Selected, minV, maxV, steps, ticks)
	}, u.window)
}

func (u *UI) runSweep(param string, min, max float64, steps, ticks int) {
	base := u.sim.Params
	results := make([]sweepResult, steps)
	step := (max - min) / float64(steps-1)
	for i := 0; i < steps; i++ {
		val := min + float64(i)*step
		p := base
		switch param {
		case "division_atp": p.DivisionATP = val
		case "mutation_rate": p.MutationRate = val
		case "translation_k": p.TranslationK = val
		case "base_death_rate": p.BaseDeathRate = val
		case "env_glucose": p.EnvGlucose = val
		}
		sim := NewSim(p, time.Now().UnixNano()+int64(i))
		for t := 0; t < ticks; t++ { sim.Step() }
		atp := 0.0
		for _, c := range sim.Cells { if c.Alive { atp += c.ATP } }
		results[i] = sweepResult{ParamValue: val, Alive: sim.AliveCount(), TotalATP: atp, MaxGen: sim.MaxGen(), Steady: sim.Steady}
	}
	u.showSweepResults(param, results)
}

func (u *UI) showSweepResults(param string, results []sweepResult) {
	var sb strings.Builder
	fmt.Fprintf(&sb, "Parameter: %s\n\n", param)
	fmt.Fprintf(&sb, "%-12s %-8s %-12s %-8s %-8s\n", "Value", "Alive", "Total ATP", "MaxGen", "Steady")
	fmt.Fprintf(&sb, "%s\n", strings.Repeat("-", 52))
	for _, r := range results {
		steadyMark := "no"
		if r.Steady { steadyMark = "yes" }
		fmt.Fprintf(&sb, "%-12.4f %-8d %-12.2f %-8d %-8s\n", r.ParamValue, r.Alive, r.TotalATP, r.MaxGen, steadyMark)
	}
	ta := widget.NewMultiLineEntry()
	ta.Wrapping = fyne.TextWrapOff
	ta.TextStyle = fyne.TextStyle{Monospace: true}
	ta.SetText(sb.String())
	d := dialog.NewCustom("Sweep Results", "Close", container.NewScroll(ta), u.window)
	d.Resize(fyne.NewSize(620, 480))
	d.Show()
}

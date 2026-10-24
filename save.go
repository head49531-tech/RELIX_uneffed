package main

import (
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
)

// savedSim mirrors Sim but skips Rng (unexported-internal, not serializable).
type savedSim struct {
	Version        int                `json:"version"`
	Tick           int                `json:"tick"`
	Cells          []*Cell            `json:"cells"`
	NextID         int                `json:"next_id"`
	Params         SimParams          `json:"params"`
	EnvPool        float64            `json:"env_pool"`
	History        []HistoryPoint     `json:"history"`
	SpeciesHistory []SpeciesPoint     `json:"species_history"`
	Steady         bool               `json:"steady"`
	SteadyTicks    int                `json:"steady_ticks"`
	Inducers       map[string]float64 `json:"inducers"`
	AHL            float64            `json:"ahl"`
	Stochastic     bool               `json:"stochastic"`
	Chem           *Chemex            `json:"chem"`
}

func (s *Sim) ToSaved() *savedSim {
	return &savedSim{
		Version:        1,
		Tick:           s.Tick,
		Cells:          s.Cells,
		NextID:         s.NextID,
		Params:         s.Params,
		EnvPool:        s.EnvPool,
		History:        s.History,
		SpeciesHistory: s.SpeciesHistory,
		Steady:         s.Steady,
		SteadyTicks:    s.SteadyTicks,
		Inducers:       s.Inducers,
		AHL:            s.AHL,
		Stochastic:     s.Stochastic,
		Chem:           s.Chem,
	}
}

func (sv *savedSim) ToSim() *Sim {
	s := &Sim{
		Tick:           sv.Tick,
		Cells:          sv.Cells,
		NextID:         sv.NextID,
		Params:         sv.Params,
		EnvPool:        sv.EnvPool,
		History:        sv.History,
		SpeciesHistory: sv.SpeciesHistory,
		Steady:         sv.Steady,
		SteadyTicks:    sv.SteadyTicks,
		Inducers:       sv.Inducers,
		AHL:            sv.AHL,
		Stochastic:     sv.Stochastic,
		Chem:           sv.Chem,
		Rng:            rand.New(rand.NewSource(time.Now().UnixNano())),
	}
	if s.Inducers == nil {
		s.Inducers = map[string]float64{"IPTG": 0, "aTc": 0, "Arabinose": 0}
	}
	if s.Chem == nil {
		s.Chem = NewChemex()
	}
	return s
}

func (u *UI) saveProject() {
	d := dialog.NewFileSave(func(w fyne.URIWriteCloser, err error) {
		if err != nil || w == nil {
			return
		}
		defer w.Close()
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		if err := enc.Encode(u.sim.ToSaved()); err != nil {
			dialog.ShowError(fmt.Errorf("save failed: %w", err), u.window)
		}
	}, u.window)
	d.SetFileName("relix-project.json")
	d.Show()
}

func (u *UI) loadProject() {
	d := dialog.NewFileOpen(func(r fyne.URIReadCloser, err error) {
		if err != nil || r == nil {
			return
		}
		defer r.Close()
		data, err := io.ReadAll(r)
		if err != nil {
			dialog.ShowError(err, u.window)
			return
		}
		var sv savedSim
		if err := json.Unmarshal(data, &sv); err != nil {
			dialog.ShowError(fmt.Errorf("load failed: %w", err), u.window)
			return
		}
		u.sim = sv.ToSim()
		u.selectedCell = -1
		u.sortedCache = nil
		if u.chemSearch != nil {
			u.rebuildChemBox(u.chemSearch.Text)
		}
		u.render()
	}, u.window)
	d.Show()
}

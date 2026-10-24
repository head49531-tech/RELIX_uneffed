package main

import (
	"fmt"
	"sort"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// showBehaviorBrowser is the interactive replacement for the old static
// text summary. Users can sort, filter, click rows, and load/pin directly.
func (u *UI) showBehaviorBrowser(reports map[string]*BehaviorReport) {
	if len(reports) == 0 {
		dialog.ShowInformation("Behavior browser",
			"No reports. Run 'Screen behaviors' first.", u.window)
		return
	}

	type row struct {
		Preset *DNAPreset
		Class  BehaviorClass
		Period float64
		Alive  float64
		PopCV  float64
	}

	base := make([]row, 0, len(reports))
	for _, p := range u.presetsCache {
		rep, ok := reports[p.Name]
		if !ok || rep == nil {
			continue
		}
		base = append(base, row{
			Preset: p, Class: rep.Class,
			Period: rep.Period, Alive: rep.Alive, PopCV: rep.PopCV,
		})
	}

	var filtered = append([]row{}, base...)
	classFilter := "All"
	textFilter := ""
	sortMode := "class"
	selectedIdx := -1

	// Table widget
	var table *widget.Table
	table = widget.NewTable(
		func() (int, int) { return len(filtered), 6 },
		func() fyne.CanvasObject { return widget.NewLabel("placeholder") },
		func(id widget.TableCellID, o fyne.CanvasObject) {
			lbl := o.(*widget.Label)
			if id.Row < 0 || id.Row >= len(filtered) {
				lbl.SetText("")
				return
			}
			r := filtered[id.Row]
			switch id.Col {
			case 0:
				lbl.SetText(r.Preset.Name)
			case 1:
				lbl.SetText(r.Preset.Category)
			case 2:
				lbl.SetText(string(r.Class))
			case 3:
				if r.Period > 0 {
					lbl.SetText(fmt.Sprintf("%.0f", r.Period))
				} else {
					lbl.SetText("—")
				}
			case 4:
				lbl.SetText(fmt.Sprintf("%.1f", r.Alive))
			case 5:
				lbl.SetText(fmt.Sprintf("%.2f", r.PopCV))
			}
		},
	)
	table.SetColumnWidth(0, 200)
	table.SetColumnWidth(1, 110)
	table.SetColumnWidth(2, 80)
	table.SetColumnWidth(3, 70)
	table.SetColumnWidth(4, 70)
	table.SetColumnWidth(5, 70)

	applyFilters := func() {
		q := strings.ToLower(strings.TrimSpace(textFilter))
		filtered = filtered[:0]
		for _, r := range base {
			if classFilter != "All" && string(r.Class) != classFilter {
				continue
			}
			if q != "" && !strings.Contains(strings.ToLower(r.Preset.Name), q) {
				continue
			}
			filtered = append(filtered, r)
		}
		switch sortMode {
		case "class":
			sort.SliceStable(filtered, func(i, j int) bool {
				if filtered[i].Class != filtered[j].Class {
					return filtered[i].Class < filtered[j].Class
				}
				return filtered[i].Preset.Name < filtered[j].Preset.Name
			})
		case "period":
			sort.SliceStable(filtered, func(i, j int) bool {
				pi, pj := filtered[i].Period, filtered[j].Period
				if pi == 0 && pj == 0 {
					return filtered[i].Preset.Name < filtered[j].Preset.Name
				}
				if pi == 0 {
					return false
				}
				if pj == 0 {
					return true
				}
				return pi < pj
			})
		case "alive":
			sort.SliceStable(filtered, func(i, j int) bool {
				return filtered[i].Alive > filtered[j].Alive
			})
		default: // name
			sort.SliceStable(filtered, func(i, j int) bool {
				return filtered[i].Preset.Name < filtered[j].Preset.Name
			})
		}
		table.Refresh()
	}

	table.OnSelected = func(id widget.TableCellID) {
		if id.Row < 0 || id.Row >= len(filtered) {
			selectedIdx = -1
			return
		}
		selectedIdx = id.Row
	}

	// --- Filter / sort controls ---
	classSel := widget.NewSelect(
		[]string{"All", "osc", "tog", "stable", "growing", "chaotic", "extinct"},
		func(s string) { classFilter = s; applyFilters() },
	)
	classSel.SetSelected("All")

	searchIn := widget.NewEntry()
	searchIn.SetPlaceHolder("Filter by name…")
	searchIn.OnChanged = func(s string) { textFilter = s; applyFilters() }

	sortBar := container.NewHBox(
		widget.NewLabel("Sort by:"),
		widget.NewButton("Name", func() { sortMode = "name"; applyFilters() }),
		widget.NewButton("Class", func() { sortMode = "class"; applyFilters() }),
		widget.NewButton("Period", func() { sortMode = "period"; applyFilters() }),
		widget.NewButton("Alive", func() { sortMode = "alive"; applyFilters() }),
	)

	filterBar := container.NewHBox(
		widget.NewLabel("Class:"), classSel,
		widget.NewLabel("Search:"), searchIn,
	)

	// --- Column headers ---
	headers := container.NewGridWithColumns(6,
		mono("Name"),
		mono("Category"),
		mono("Class"),
		mono("Period"),
		mono("Alive"),
		mono("PopCV"),
	)

	// --- Footer actions ---
	loadBtn := widget.NewButton("Load into current cell", func() {
		if selectedIdx < 0 || selectedIdx >= len(filtered) {
			return
		}
		u.loadPresetDirect(filtered[selectedIdx].Preset)
	})
	loadAsBtn := widget.NewButton("Spawn as new cell", func() {
		if selectedIdx < 0 || selectedIdx >= len(filtered) {
			return
		}
		u.spawnCellDirect(filtered[selectedIdx].Preset)
	})
	pinABtn := widget.NewButton("Set as Pin A", func() {
		if selectedIdx < 0 || selectedIdx >= len(filtered) {
			return
		}
		u.presetPinA = filtered[selectedIdx].Preset
		if u.presetPinALabel != nil {
			u.presetPinALabel.SetText("A: " + filtered[selectedIdx].Preset.Name)
		}
	})
	pinBBtn := widget.NewButton("Set as Pin B", func() {
		if selectedIdx < 0 || selectedIdx >= len(filtered) {
			return
		}
		u.presetPinB = filtered[selectedIdx].Preset
		if u.presetPinBLabel != nil {
			u.presetPinBLabel.SetText("B: " + filtered[selectedIdx].Preset.Name)
		}
	})

	footer := container.NewVBox(
		widget.NewSeparator(),
		container.NewHBox(loadBtn, loadAsBtn, pinABtn, pinBBtn),
	)

	top := container.NewVBox(
		widget.NewLabelWithStyle(
			fmt.Sprintf("Behavior Browser — %d classified presets", len(base)),
			fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		filterBar,
		sortBar,
		widget.NewSeparator(),
		headers,
	)

	applyFilters()

	content := container.NewBorder(top, footer, nil, nil, table)
	d := dialog.NewCustom("Behavior Browser", "Close", content, u.window)
	d.Resize(fyne.NewSize(900, 720))
	d.Show()
}

// loadPresetDirect applies a preset to the currently-selected cell without
// going through the presets-tab index. Publicly reusable.
func (u *UI) loadPresetDirect(p *DNAPreset) {
	if p == nil {
		return
	}
	c := u.getSelectedCell()
	if c == nil {
		dialog.ShowInformation("Load preset", "Select a cell first.", u.window)
		return
	}
	c.Genome = p.genomeFromPreset()
	c.MRNA = map[string]float64{}
	c.Proteins = map[string]float64{}
	c.Metabolites = map[string]float64{}
	for k, v := range p.InitMRNA {
		c.MRNA[k] = v
	}
	for k, v := range p.InitProteins {
		c.Proteins[k] = v
	}
	u.renderGenome()
	u.refreshSpecies()
	u.refreshDetail()
}

// spawnCellDirect spawns a new cell from a preset without going through the
// presets-tab index. Publicly reusable.
func (u *UI) spawnCellDirect(p *DNAPreset) {
	if p == nil {
		return
	}
	if len(u.sim.Cells) >= u.sim.Params.MaxCells {
		dialog.ShowInformation("Cannot spawn", "Max cells reached.", u.window)
		return
	}
	c := newCellFromPreset(u.sim, p)
	for k, v := range p.InitMRNA {
		c.MRNA[k] = v
	}
	for k, v := range p.InitProteins {
		c.Proteins[k] = v
	}
	u.sim.Cells = append(u.sim.Cells, c)
	u.selectedCell = c.ID
	u.render()
}

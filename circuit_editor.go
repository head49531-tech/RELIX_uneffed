package main

import (
	"fmt"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// circuitEditorState holds the working draft for one editor session.
// Each open dialog gets its own state; nothing leaks onto UI.
type circuitEditorState struct {
	source   *DNAPreset
	genes    []*Gene
	selected int

	listWidget *widget.List
	preview    *widget.Label

	nameIn    *widget.Entry
	productIn *widget.Entry
	basalIn   *widget.Entry
	actkIn    *widget.Entry
	repkIn    *widget.Entry
	hillIn    *widget.Entry
	actIn     *widget.Entry
	repIn     *widget.Entry
}

// ==================== Helpers ====================

func formRow(label string, w fyne.CanvasObject) fyne.CanvasObject {
	lbl := widget.NewLabel(label)
	return container.NewHBox(lbl, w)
}

func splitCSV(s string) []string {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		t := strings.TrimSpace(p)
		if t != "" {
			out = append(out, t)
		}
	}
	return out
}

func geneSummary(g *Gene) string {
	name := g.Name
	if name == "" {
		name = "?"
	}
	prod := g.Product
	if prod == "" {
		prod = "?"
	}
	act := "act=[-]"
	if len(g.Activators) > 0 {
		act = "act=[" + strings.Join(g.Activators, ",") + "]"
	}
	rep := "rep=[-]"
	if len(g.Repressors) > 0 {
		rep = "rep=[" + strings.Join(g.Repressors, ",") + "]"
	}
	return fmt.Sprintf("%-10s→%-12s b=%.2f n=%.1f %s %s", name, prod, g.BasalTranscription, g.HillN, act, rep)
}

func deepCloneGenes(genes []Gene) []Gene {
	out := make([]Gene, len(genes))
	for i, g := range genes {
		c := g
		c.Activators = append([]string{}, g.Activators...)
		c.Repressors = append([]string{}, g.Repressors...)
		c.Inducers = append([]string{}, g.Inducers...)
		c.InducerModes = append([]string{}, g.InducerModes...)
		c.InducerKs = append([]float64{}, g.InducerKs...)
		out[i] = c
	}
	return out
}

// buildPresetFromDraft turns the editor draft into a savable DNAPreset.
func buildPresetFromDraft(st *circuitEditorState, name, desc string) *DNAPreset {
	spec := presetSpec{Name: name, Category: "User", Description: desc}
	for _, g := range st.genes {
		spec.Genes = append(spec.Genes, *g)
	}
	spec.Genes = deepCloneGenes(spec.Genes)
	p := spec.makePreset()
	p.Builtin = false
	return p
}

// previewDraft classifies the draft circuit for 200 ticks.
func previewDraft(st *circuitEditorState) *BehaviorReport {
	if len(st.genes) == 0 {
		return nil
	}
	p := buildPresetFromDraft(st, "draft", "")
	return ClassifyPreset(p, 200, 1)
}

// ==================== Editor dialog ====================

func (u *UI) openCircuitEditor(p *DNAPreset) {
	if p == nil {
		dialog.ShowInformation("Edit circuit", "Select a preset first.", u.window)
		return
	}

	st := &circuitEditorState{source: p, selected: -1}
	for _, g := range p.Genes {
		c := *g
		c.Activators = append([]string{}, g.Activators...)
		c.Repressors = append([]string{}, g.Repressors...)
		c.Inducers = append([]string{}, g.Inducers...)
		c.InducerModes = append([]string{}, g.InducerModes...)
		c.InducerKs = append([]float64{}, g.InducerKs...)
		st.genes = append(st.genes, &c)
	}

	// --- widgets ---
	st.listWidget = widget.NewList(
		func() int { return len(st.genes) },
		func() fyne.CanvasObject { return mono("template") },
		func(i widget.ListItemID, o fyne.CanvasObject) {
			if i < 0 || i >= len(st.genes) {
				return
			}
			o.(*widget.Label).SetText(geneSummary(st.genes[i]))
		},
	)

	st.nameIn = widget.NewEntry()
	st.productIn = widget.NewEntry()
	st.basalIn = widget.NewEntry()
	st.actkIn = widget.NewEntry()
	st.repkIn = widget.NewEntry()
	st.hillIn = widget.NewEntry()
	st.actIn = widget.NewEntry()
	st.repIn = widget.NewEntry()

	st.preview = mono("(click 'Preview behavior' to classify the draft)")

	// Apply current form values to the selected gene.
	apply := func() {
		if st.selected < 0 || st.selected >= len(st.genes) {
			return
		}
		g := st.genes[st.selected]
		g.Name = strings.TrimSpace(st.nameIn.Text)
		g.Product = strings.TrimSpace(st.productIn.Text)
		if v, err := strconv.ParseFloat(st.basalIn.Text, 64); err == nil {
			g.BasalTranscription = v
		}
		if v, err := strconv.ParseFloat(st.actkIn.Text, 64); err == nil {
			g.ActivationK = v
		}
		if v, err := strconv.ParseFloat(st.repkIn.Text, 64); err == nil {
			g.RepressionK = v
		}
		if v, err := strconv.ParseFloat(st.hillIn.Text, 64); err == nil {
			g.HillN = v
		}
		g.Activators = splitCSV(st.actIn.Text)
		g.Repressors = splitCSV(st.repIn.Text)
		st.listWidget.Refresh()
	}

	st.nameIn.OnChanged = func(string) { apply() }
	st.productIn.OnChanged = func(string) { apply() }
	st.basalIn.OnChanged = func(string) { apply() }
	st.actkIn.OnChanged = func(string) { apply() }
	st.repkIn.OnChanged = func(string) { apply() }
	st.hillIn.OnChanged = func(string) { apply() }
	st.actIn.OnChanged = func(string) { apply() }
	st.repIn.OnChanged = func(string) { apply() }

	loadGene := func() {
		if st.selected < 0 || st.selected >= len(st.genes) {
			st.nameIn.SetText("")
			st.productIn.SetText("")
			st.basalIn.SetText("")
			st.actkIn.SetText("")
			st.repkIn.SetText("")
			st.hillIn.SetText("")
			st.actIn.SetText("")
			st.repIn.SetText("")
			return
		}
		g := st.genes[st.selected]
		st.nameIn.SetText(g.Name)
		st.productIn.SetText(g.Product)
		st.basalIn.SetText(fmt.Sprintf("%g", g.BasalTranscription))
		st.actkIn.SetText(fmt.Sprintf("%g", g.ActivationK))
		st.repkIn.SetText(fmt.Sprintf("%g", g.RepressionK))
		st.hillIn.SetText(fmt.Sprintf("%g", g.HillN))
		st.actIn.SetText(strings.Join(g.Activators, ","))
		st.repIn.SetText(strings.Join(g.Repressors, ","))
	}

	st.listWidget.OnSelected = func(id widget.ListItemID) {
		st.selected = id
		loadGene()
	}

	addBtn := widget.NewButton("+ Add gene", func() {
		ng := &Gene{
			Name:               fmt.Sprintf("gene%d", len(st.genes)+1),
			Product:            fmt.Sprintf("P%d", len(st.genes)+1),
			BasalTranscription: 0.1,
			ActivationK:        15,
			RepressionK:        15,
			HillN:              2,
		}
		st.genes = append(st.genes, ng)
		st.selected = len(st.genes) - 1
		st.listWidget.Refresh()
		st.listWidget.Select(st.selected)
		loadGene()
	})

	cloneBtn := widget.NewButton("Clone gene", func() {
		if st.selected < 0 || st.selected >= len(st.genes) {
			return
		}
		src := st.genes[st.selected]
		c := *src
		c.Name = src.Name + "_copy"
		c.Product = src.Product
		c.Activators = append([]string{}, src.Activators...)
		c.Repressors = append([]string{}, src.Repressors...)
		st.genes = append(st.genes, &c)
		st.selected = len(st.genes) - 1
		st.listWidget.Refresh()
		st.listWidget.Select(st.selected)
		loadGene()
	})

	delBtn := widget.NewButton("Delete gene", func() {
		if st.selected < 0 || st.selected >= len(st.genes) {
			return
		}
		st.genes = append(st.genes[:st.selected], st.genes[st.selected+1:]...)
		if st.selected >= len(st.genes) {
			st.selected = len(st.genes) - 1
		}
		st.listWidget.Refresh()
		if st.selected >= 0 {
			st.listWidget.Select(st.selected)
			loadGene()
		} else {
			loadGene()
		}
	})

	previewBtn := widget.NewButton("Preview behavior", func() {
		rep := previewDraft(st)
		if rep == nil {
			st.preview.SetText("No genes in draft.")
			return
		}
		periodStr := "—"
		if rep.Period > 0 {
			periodStr = fmt.Sprintf("%.0ft", rep.Period)
		}
		st.preview.SetText(fmt.Sprintf(
			"Class: %s\nPeriod: %s\nAlive: %.1f\nPopCV: %.2f\nSpecies tracked: %d\nErr: %s",
			rep.Class, periodStr, rep.Alive, rep.PopCV, len(rep.Species), rep.Err))
	})

	saveBtn := widget.NewButton("Save as new preset…", func() {
		nameIn := widget.NewEntry()
		nameIn.SetText(st.source.Name + " (edited)")
		descIn := widget.NewMultiLineEntry()
		descIn.SetText("Derived from " + st.source.Name)
		dialog.ShowForm("Save edited circuit", "Save", "Cancel",
			[]*widget.FormItem{
				{Text: "Name", Widget: nameIn},
				{Text: "Description", Widget: descIn},
			}, func(ok bool) {
				if !ok {
					return
				}
				name := strings.TrimSpace(nameIn.Text)
				if name == "" {
					dialog.ShowInformation("Save", "Name required.", u.window)
					return
				}
				np := buildPresetFromDraft(st, name, strings.TrimSpace(descIn.Text))
				if err := saveUserPreset(np); err != nil {
					dialog.ShowError(err, u.window)
					return
				}
				u.refreshPresetsList()
				dialog.ShowInformation("Saved", "Preset '"+name+"' created.", u.window)
			}, u.window)
	})

	// --- layout ---
	left := container.NewBorder(
		widget.NewLabelWithStyle("Draft genes", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		container.NewVBox(
			container.NewHBox(addBtn, cloneBtn),
			container.NewHBox(delBtn),
		),
		nil, nil,
		st.listWidget,
	)

	form := container.NewVBox(
		widget.NewLabelWithStyle("Selected gene", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		formRow("Name", st.nameIn),
		formRow("Product", st.productIn),
		formRow("Basal", st.basalIn),
		formRow("Act K", st.actkIn),
		formRow("Rep K", st.repkIn),
		formRow("Hill N", st.hillIn),
		formRow("Activators (csv)", st.actIn),
		formRow("Repressors (csv)", st.repIn),
		widget.NewSeparator(),
		previewBtn,
		st.preview,
	)
	right := container.NewVScroll(form)
	split := container.NewHSplit(left, right)
	split.SetOffset(0.4)

	top := container.NewVBox(
		widget.NewLabelWithStyle("Circuit Editor — "+st.source.Name, fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabel(fmt.Sprintf("%d genes in draft  |  source has %d", len(st.genes), len(st.source.Genes))),
		widget.NewSeparator(),
	)
	bottom := container.NewVBox(
		widget.NewSeparator(),
		container.NewHBox(saveBtn),
	)

	d := dialog.NewCustom("Circuit Editor", "Close",
		container.NewBorder(top, bottom, nil, nil, split), u.window)
	d.Resize(fyne.NewSize(1000, 720))
	d.Show()

	if len(st.genes) > 0 {
		st.selected = 0
		st.listWidget.Select(0)
		loadGene()
	}
}

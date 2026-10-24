package main

import (
	"fmt"
	"sort"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
)

func (u *UI) buildTimecourseCSV() string {
	if len(u.sim.SpeciesHistory) == 0 {
		return "tick,alive,atp,glucose,env_glucose\n"
	}
	mrnaSet := map[string]bool{}
	protSet := map[string]bool{}
	for _, sp := range u.sim.SpeciesHistory {
		for k := range sp.MRNA { mrnaSet[k] = true }
		for k := range sp.Proteins { protSet[k] = true }
	}
	mrnaNames := make([]string, 0, len(mrnaSet))
	for k := range mrnaSet { mrnaNames = append(mrnaNames, k) }
	sort.Strings(mrnaNames)
	protNames := make([]string, 0, len(protSet))
	for k := range protSet { protNames = append(protNames, k) }
	sort.Strings(protNames)

	var sb strings.Builder
	header := []string{"tick", "alive", "atp", "glucose", "env_glucose"}
	for _, n := range mrnaNames { header = append(header, "mRNA_"+n) }
	for _, n := range protNames { header = append(header, "prot_"+n) }
	sb.WriteString(strings.Join(header, ",") + "\n")

	for _, sp := range u.sim.SpeciesHistory {
		row := []string{
			fmt.Sprintf("%d", sp.Tick),
			fmt.Sprintf("%d", sp.Alive),
			fmt.Sprintf("%.4f", sp.TotalATP),
			fmt.Sprintf("%.4f", sp.TotalGlucose),
			fmt.Sprintf("%.4f", sp.EnvGlucose),
		}
		for _, n := range mrnaNames { row = append(row, fmt.Sprintf("%.4f", sp.MRNA[n])) }
		for _, n := range protNames { row = append(row, fmt.Sprintf("%.4f", sp.Proteins[n])) }
		sb.WriteString(strings.Join(row, ",") + "\n")
	}
	return sb.String()
}

func (u *UI) exportCSV() {
	d := dialog.NewFileSave(func(w fyne.URIWriteCloser, err error) {
		if err != nil || w == nil { return }
		defer w.Close()
		fmt.Fprint(w, u.buildTimecourseCSV())
	}, u.window)
	d.SetFileName("relix-timecourse.csv")
	d.Show()
}

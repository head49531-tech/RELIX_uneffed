package main

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"

	"github.com/phpdave11/gofpdf"
)

func (u *UI) buildTXTReport() string {
	var sb strings.Builder
	cells := u.sim.SortedCells()
	alive := u.sim.AliveCount()
	maxGen := u.sim.MaxGen()

	fmt.Fprintf(&sb, "RELIX SIMULATION RESULTS\n========================\n\n")
	fmt.Fprintf(&sb, "Generated: %s\n", time.Now().Format(time.RFC1123))
	fmt.Fprintf(&sb, "Ticks:     %d\n", u.sim.Tick)
	fmt.Fprintf(&sb, "Cells:     %d total, %d alive\n", len(cells), alive)
	fmt.Fprintf(&sb, "Max gen:   %d\n", maxGen)
	fmt.Fprintf(&sb, "Env pool:  %.2f glucose remaining\n\n", u.sim.EnvPool)

	fmt.Fprintf(&sb, "--- PARAMETERS ---\n")
	fmt.Fprintf(&sb, "Mutation rate:      %.4f\n", u.sim.Params.MutationRate)
	fmt.Fprintf(&sb, "Base death rate:    %.4f\n", u.sim.Params.BaseDeathRate)
	fmt.Fprintf(&sb, "Division ATP:       %.2f\n", u.sim.Params.DivisionATP)
	fmt.Fprintf(&sb, "Translation k:      %.2f\n", u.sim.Params.TranslationK)
	fmt.Fprintf(&sb, "mRNA degradation:   %.4f\n", u.sim.Params.MRNADegradation)
	fmt.Fprintf(&sb, "Protein degradation:%.4f\n", u.sim.Params.ProteinDegradation)
	fmt.Fprintf(&sb, "Enzyme kcat:        %.2f\n", u.sim.Params.EnzymeKcat)
	fmt.Fprintf(&sb, "Glucose influx:     %.2f\n\n", u.sim.Params.GlucoseInflux)

	fmt.Fprintf(&sb, "--- GLOBAL SPECIES (sum across alive cells) ---\n")
	fmt.Fprintf(&sb, "mRNA:\n")
	mrnaLevels := u.sim.GlobalMRNALevels()
	keys := make([]string, 0, len(mrnaLevels))
	for k := range mrnaLevels {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		fmt.Fprintf(&sb, "  %-10s %.2f\n", k, mrnaLevels[k])
	}
	fmt.Fprintf(&sb, "Proteins:\n")
	protLevels := u.sim.GlobalProteinLevels()
	keys = keys[:0]
	for k := range protLevels {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		fmt.Fprintf(&sb, "  %-14s %.2f\n", k, protLevels[k])
	}
	fmt.Fprintf(&sb, "\n--- CELLS ---\n\n")
	for _, c := range cells {
		status := "DEAD:" + c.DeathCause
		if c.Alive {
			status = "ALIVE"
		}
		fmt.Fprintf(&sb, "Cell #%d [%s] gen=%d parent=%d age=%d\n", c.ID, status, c.Generation, c.ParentID, c.Age)
		fmt.Fprintf(&sb, "  ATP=%.2f  glucose=%.2f  totalATP=%.1f\n", c.ATP, c.Glucose, c.TotalATPProduced)
		fmt.Fprintf(&sb, "  DNA (%d nt): %s\n", len(c.Genome.DNA), string(c.Genome.DNA))
		fmt.Fprintf(&sb, "  Genes: ")
		for _, g := range c.Genome.Genes {
			fmt.Fprintf(&sb, "%s→%s ", g.Name, g.Product)
		}
		fmt.Fprintf(&sb, "\n")
		fmt.Fprintf(&sb, "  mRNA: ")
		for k, v := range c.MRNA {
			fmt.Fprintf(&sb, "%s=%.2f ", k, v)
		}
		fmt.Fprintf(&sb, "\n  Proteins: ")
		for k, v := range c.Proteins {
			fmt.Fprintf(&sb, "%s=%.2f ", k, v)
		}
		fmt.Fprintf(&sb, "\n\n")
	}
	return sb.String()
}

func (u *UI) exportTXT() {
	d := dialog.NewFileSave(func(w fyne.URIWriteCloser, err error) {
		if err != nil || w == nil {
			return
		}
		defer w.Close()
		fmt.Fprint(w, u.buildTXTReport())
	}, u.window)
	d.SetFileName("relix-report.txt")
	d.Show()
}

func (u *UI) exportPDF() {
	d := dialog.NewFileSave(func(w fyne.URIWriteCloser, err error) {
		if err != nil || w == nil {
			return
		}
		path := w.URI().Path()
		w.Close()
		if err := u.writePDF(path); err != nil {
			dialog.ShowError(err, u.window)
		}
	}, u.window)
	d.SetFileName("relix-report.pdf")
	d.Show()
}

func (u *UI) writePDF(path string) error {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(15, 15, 15)
	pdf.AddPage()

	pdf.SetFont("Helvetica", "B", 20)
	pdf.CellFormat(0, 12, "RELIX Simulation Report", "", 1, "L", false, 0, "")
	pdf.SetFont("Helvetica", "I", 9)
	pdf.SetTextColor(120, 120, 120)
	pdf.CellFormat(0, 6, time.Now().Format(time.RFC1123), "", 1, "L", false, 0, "")
	pdf.SetTextColor(0, 0, 0)
	pdf.Ln(4)

	alive := u.sim.AliveCount()
	maxGen := u.sim.MaxGen()
	pdf.SetFont("Helvetica", "B", 12)
	pdf.CellFormat(0, 8, "Summary", "", 1, "L", false, 0, "")
	pdf.SetFont("Helvetica", "", 10)
	for _, kv := range [][2]string{
		{"Ticks", fmt.Sprintf("%d", u.sim.Tick)},
		{"Total cells", fmt.Sprintf("%d", len(u.sim.Cells))},
		{"Alive cells", fmt.Sprintf("%d", alive)},
		{"Max generation", fmt.Sprintf("%d", maxGen)},
		{"Env glucose", fmt.Sprintf("%.2f", u.sim.EnvPool)},
	} {
		pdf.CellFormat(60, 6, kv[0], "1", 0, "L", false, 0, "")
		pdf.CellFormat(0, 6, kv[1], "1", 1, "L", false, 0, "")
	}

	pdf.Ln(4)
	pdf.SetFont("Helvetica", "B", 12)
	pdf.CellFormat(0, 8, "Global Species", "", 1, "L", false, 0, "")
	pdf.SetFont("Courier", "", 9)
	mrna := u.sim.GlobalMRNALevels()
	mrnaKeys := make([]string, 0, len(mrna))
	for k := range mrna {
		mrnaKeys = append(mrnaKeys, k)
	}
	sort.Strings(mrnaKeys)
	for _, k := range mrnaKeys {
		pdf.CellFormat(0, 5, fmt.Sprintf("mRNA   %-10s %.2f", k, mrna[k]), "", 1, "L", false, 0, "")
	}
	prot := u.sim.GlobalProteinLevels()
	protKeys := make([]string, 0, len(prot))
	for k := range prot {
		protKeys = append(protKeys, k)
	}
	sort.Strings(protKeys)
	for _, k := range protKeys {
		pdf.CellFormat(0, 5, fmt.Sprintf("prot   %-14s %.2f", k, prot[k]), "", 1, "L", false, 0, "")
	}

	pdf.AddPage()
	pdf.SetFont("Helvetica", "B", 12)
	pdf.CellFormat(0, 8, "Cells", "", 1, "L", false, 0, "")
	pdf.SetFont("Helvetica", "B", 9)
	pdf.CellFormat(15, 6, "ID", "1", 0, "L", false, 0, "")
	pdf.CellFormat(20, 6, "Status", "1", 0, "L", false, 0, "")
	pdf.CellFormat(15, 6, "Gen", "1", 0, "L", false, 0, "")
	pdf.CellFormat(20, 6, "ATP", "1", 0, "L", false, 0, "")
	pdf.CellFormat(0, 6, "DNA (first 40 nt)", "1", 1, "L", false, 0, "")
	pdf.SetFont("Courier", "", 8)
	for i, c := range u.sim.SortedCells() {
		if i >= 40 {
			break
		}
		status := "alive"
		if !c.Alive {
			status = "dead"
		}
		dna := string(c.Genome.DNA)
		if len(dna) > 40 {
			dna = dna[:40]
		}
		pdf.CellFormat(15, 5, fmt.Sprintf("%d", c.ID), "1", 0, "L", false, 0, "")
		pdf.CellFormat(20, 5, status, "1", 0, "L", false, 0, "")
		pdf.CellFormat(15, 5, fmt.Sprintf("%d", c.Generation), "1", 0, "L", false, 0, "")
		pdf.CellFormat(20, 5, fmt.Sprintf("%.1f", c.ATP), "1", 0, "L", false, 0, "")
		pdf.CellFormat(0, 5, dna, "1", 1, "L", false, 0, "")
	}
	return pdf.OutputFileAndClose(path)
}

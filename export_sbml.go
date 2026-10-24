package main

import (
	"fmt"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
)

func mmPlus(a, b string) string  { return fmt.Sprintf(`<apply><plus/>%s%s</apply>`, a, b) }
func mmMinus(a, b string) string { return fmt.Sprintf(`<apply><minus/>%s%s</apply>`, a, b) }
func mmTimes(a, b string) string { return fmt.Sprintf(`<apply><times/>%s%s</apply>`, a, b) }
func mmDivide(a, b string) string { return fmt.Sprintf(`<apply><divide/>%s%s</apply>`, a, b) }
func mmPower(a, b string) string  { return fmt.Sprintf(`<apply><power/>%s%s</apply>`, a, b) }
func mmCi(name string) string     { return "<ci>" + name + "</ci>" }
func mmCn(v string) string        { return "<cn>" + v + "</cn>" }
func mmMath(e string) string      { return `<math xmlns="http://www.w3.org/1998/Math/MathML">` + e + `</math>` }

func hillExpr(tf, K, n string) string {
	num := mmPower(mmCi(tf), mmCi(n))
	den := mmPlus(mmPower(mmCi(K), mmCi(n)), mmPower(mmCi(tf), mmCi(n)))
	return mmDivide(num, den)
}

func (u *UI) buildSBML() string {
	c := u.getSelectedCell()
	if c == nil {
		for _, x := range u.sim.Cells {
			if x.Alive { c = x; break }
		}
	}
	if c == nil {
		return "<?xml version=\"1.0\"?>\n<!-- no cell -->\n"
	}
	var sb strings.Builder
	sb.WriteString("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n")
	sb.WriteString("<sbml xmlns=\"http://www.sbml.org/sbml/level3/version1/core\" level=\"3\" version=\"1\">\n")
	sb.WriteString(fmt.Sprintf("  <model id=\"relix_cell_%d\">\n", c.ID))

	sb.WriteString("    <listOfCompartments><compartment id=\"cell\" size=\"1\" constant=\"true\"/></listOfCompartments>\n")

	sb.WriteString("    <listOfSpecies>\n")
	for _, gene := range c.Genome.Genes {
		sb.WriteString(fmt.Sprintf("      <species id=\"%s_mRNA\" compartment=\"cell\" initialConcentration=\"%g\" hasOnlySubstanceUnits=\"false\" boundaryCondition=\"false\" constant=\"false\"/>\n", gene.Name, c.MRNA[gene.Name]))
		sb.WriteString(fmt.Sprintf("      <species id=\"%s_prot\" compartment=\"cell\" initialConcentration=\"%g\" hasOnlySubstanceUnits=\"false\" boundaryCondition=\"false\" constant=\"false\"/>\n", gene.Name, c.Proteins[gene.Product]))
	}
	sb.WriteString(fmt.Sprintf("      <species id=\"Glucose\" compartment=\"cell\" initialConcentration=\"%g\" hasOnlySubstanceUnits=\"false\" boundaryCondition=\"false\" constant=\"false\"/>\n", c.Glucose))
	sb.WriteString(fmt.Sprintf("      <species id=\"ATP\" compartment=\"cell\" initialConcentration=\"%g\" hasOnlySubstanceUnits=\"false\" boundaryCondition=\"false\" constant=\"false\"/>\n", c.ATP))
	sb.WriteString("    </listOfSpecies>\n")

	sb.WriteString("    <listOfParameters>\n")
	for _, kv := range [][2]string{
		{"k_translation", fmt.Sprintf("%g", u.sim.Params.TranslationK)},
		{"k_mrna_deg", fmt.Sprintf("%g", u.sim.Params.MRNADegradation)},
		{"k_prot_deg", fmt.Sprintf("%g", u.sim.Params.ProteinDegradation)},
		{"kcat", fmt.Sprintf("%g", u.sim.Params.EnzymeKcat)},
		{"Km_glucose", "5.0"},
		{"glucose_influx", fmt.Sprintf("%g", u.sim.Params.GlucoseInflux)},
		{"atp_basal", fmt.Sprintf("%g", u.sim.Params.ATPBasal)},
	} {
		sb.WriteString(fmt.Sprintf("      <parameter id=\"%s\" value=\"%s\" constant=\"true\"/>\n", kv[0], kv[1]))
	}
	for _, gene := range c.Genome.Genes {
		sb.WriteString(fmt.Sprintf("      <parameter id=\"basal_%s\" value=\"%g\" constant=\"true\"/>\n", gene.Name, gene.BasalTranscription))
		sb.WriteString(fmt.Sprintf("      <parameter id=\"Kact_%s\" value=\"%g\" constant=\"true\"/>\n", gene.Name, gene.ActivationK))
		sb.WriteString(fmt.Sprintf("      <parameter id=\"Krep_%s\" value=\"%g\" constant=\"true\"/>\n", gene.Name, gene.RepressionK))
		sb.WriteString(fmt.Sprintf("      <parameter id=\"n_%s\" value=\"%g\" constant=\"true\"/>\n", gene.Name, gene.HillN))
		sb.WriteString(fmt.Sprintf("      <parameter id=\"rbs_%s\" value=\"%g\" constant=\"true\"/>\n", gene.Name, c.Genome.RBSStrength(gene)))
	}
	sb.WriteString("    </listOfParameters>\n")

	sb.WriteString("    <listOfRules>\n")
	for _, gene := range c.Genome.Genes {
		if gene.Disabled {
			sb.WriteString(fmt.Sprintf("      <assignmentRule variable=\"occ_%s\">%s</assignmentRule>\n", gene.Name, mmMath(mmCn("0"))))
			continue
		}
		act := mmCn("1")
		for _, tf := range gene.Activators {
			act = mmTimes(act, mmPlus(mmCn("1"), hillExpr(tf+"_prot", "Kact_"+gene.Name, "n_"+gene.Name)))
		}
		rep := mmCn("1")
		for _, tf := range gene.Repressors {
			rep = mmTimes(rep, mmMinus(mmCn("1"), mmTimes(mmCn("0.9"), hillExpr(tf+"_prot", "Krep_"+gene.Name, "n_"+gene.Name))))
		}
		sb.WriteString(fmt.Sprintf("      <assignmentRule variable=\"occ_%s\">%s</assignmentRule>\n", gene.Name, mmMath(mmTimes(act, rep))))
	}
	for _, gene := range c.Genome.Genes {
		prod := mmTimes(mmCi("basal_"+gene.Name), mmCi("occ_"+gene.Name))
		deg := mmTimes(mmCi("k_mrna_deg"), mmCi(gene.Name+"_mRNA"))
		sb.WriteString(fmt.Sprintf("      <rateRule variable=\"%s_mRNA\">%s</rateRule>\n", gene.Name, mmMath(mmMinus(prod, deg))))
		trans := mmTimes(mmTimes(mmCi("k_translation"), mmCi(gene.Name+"_mRNA")), mmCi("rbs_"+gene.Name))
		pdeg := mmTimes(mmCi("k_prot_deg"), mmCi(gene.Name+"_prot"))
		sb.WriteString(fmt.Sprintf("      <rateRule variable=\"%s_prot\">%s</rateRule>\n", gene.Name, mmMath(mmMinus(trans, pdeg))))
	}
	metab := mmTimes(mmTimes(mmCi("kcat"), mmCi("MetA_prot")), mmDivide(mmCi("Glucose"), mmPlus(mmCi("Km_glucose"), mmCi("Glucose"))))
	sb.WriteString(fmt.Sprintf("      <rateRule variable=\"Glucose\">%s</rateRule>\n", mmMath(mmMinus(mmCi("glucose_influx"), metab))))
	sb.WriteString(fmt.Sprintf("      <rateRule variable=\"ATP\">%s</rateRule>\n", mmMath(mmMinus(metab, mmCi("atp_basal")))))
	sb.WriteString("    </listOfRules>\n")
	sb.WriteString("  </model>\n</sbml>\n")
	return sb.String()
}

func (u *UI) exportSBML() {
	d := dialog.NewFileSave(func(w fyne.URIWriteCloser, err error) {
		if err != nil || w == nil { return }
		defer w.Close()
		fmt.Fprint(w, u.buildSBML())
	}, u.window)
	d.SetFileName("relix-model.xml")
	d.Show()
}

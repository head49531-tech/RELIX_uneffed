package main

import (
	"fmt"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
)

// SBOL 2.3.0 constants — SO roles and BioPAX types.
const (
	sbolSOGene     = "http://identifiers.org/so/SO:0000704"
	sbolSOCDS      = "http://identifiers.org/so/SO:0000316"
	sbolSOPlasmid  = "http://identifiers.org/so/SO:0000804"
	sbolDnaRegion  = "http://www.biopax.org/release/biopax-level3.owl#DnaRegion"
	sbolEncoding   = "http://www.chem.qmul.ac.uk/iubmb/misc/naseq.html"
	sbolInline     = "http://sbols.org/v2#inline"
)

func xmlEscape(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, `"`, "&quot;")
	s = strings.ReplaceAll(s, "'", "&apos;")
	return s
}

func sanitizeID(s string) string {
	var b strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' {
			b.WriteRune(r)
		} else {
			b.WriteRune('_')
		}
	}
	if b.Len() == 0 {
		return "unnamed"
	}
	out := b.String()
	if out[0] >= '0' && out[0] <= '9' {
		return "_" + out
	}
	return out
}

func (u *UI) buildSBOL() string {
	c := u.getSelectedCell()
	if c == nil {
		return `<?xml version="1.0" encoding="UTF-8"?>` + "\n" +
			`<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#" xmlns:sbol="http://sbols.org/v2#">` + "\n" +
			`  <!-- RELIX: no cell selected -->` + "\n" +
			`</rdf:RDF>` + "\n"
	}

	circular := u.digestCircular != nil && u.digestCircular.Checked
	topology := "linear"
	if circular {
		topology = "circular"
	}

	plasmidID := fmt.Sprintf("relix_plasmid_cell%d", c.ID)
	seqID := fmt.Sprintf("relix_seq_cell%d", c.ID)

	var sb strings.Builder
	sb.WriteString(`<?xml version="1.0" encoding="UTF-8"?>` + "\n")
	sb.WriteString(`<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"` + "\n")
	sb.WriteString(`         xmlns:sbol="http://sbols.org/v2#"` + "\n")
	sb.WriteString(`         xmlns:dcterms="http://purl.org/dc/terms/">` + "\n")

	// --- Top-level plasmid ComponentDefinition ---
	sb.WriteString(fmt.Sprintf(`  <sbol:ComponentDefinition rdf:about="http://relix.local/%s">`, plasmidID) + "\n")
	sb.WriteString(fmt.Sprintf(`    <sbol:displayId>%s</sbol:displayId>`, plasmidID) + "\n")
	sb.WriteString(fmt.Sprintf(`    <sbol:name>RELIX Cell %d Genome</sbol:name>`, c.ID) + "\n")
	sb.WriteString(fmt.Sprintf(`    <sbol:description>%s genome, %d bp, %d genes</sbol:description>`,
		topology, len(c.Genome.DNA), len(c.Genome.Genes)) + "\n")
	sb.WriteString(fmt.Sprintf(`    <sbol:type rdf:resource="%s"/>`, sbolDnaRegion) + "\n")
	sb.WriteString(fmt.Sprintf(`    <sbol:role rdf:resource="%s"/>`, sbolSOPlasmid) + "\n")

	// SequenceAnnotations: one per gene
	for i, gene := range c.Genome.Genes {
		gid := fmt.Sprintf("g%d_%s", i, sanitizeID(gene.Name))
		sb.WriteString(`    <sbol:sequenceAnnotation>` + "\n")
		sb.WriteString(fmt.Sprintf(`      <sbol:SequenceAnnotation rdf:about="http://relix.local/annot_%s">`, gid) + "\n")
		sb.WriteString(fmt.Sprintf(`        <sbol:displayId>annot_%s</sbol:displayId>`, gid) + "\n")
		sb.WriteString(fmt.Sprintf(`        <sbol:role rdf:resource="%s"/>`, sbolSOGene) + "\n")
		sb.WriteString(`        <sbol:location>` + "\n")
		sb.WriteString(fmt.Sprintf(`          <sbol:Range rdf:about="http://relix.local/range_%s">`, gid) + "\n")
		sb.WriteString(fmt.Sprintf(`            <sbol:start>%d</sbol:start>`, gene.PromoterStart+1) + "\n")
		sb.WriteString(fmt.Sprintf(`            <sbol:end>%d</sbol:end>`, gene.CDSEnd) + "\n")
		sb.WriteString(fmt.Sprintf(`            <sbol:orientation rdf:resource="%s"/>`, sbolInline) + "\n")
		sb.WriteString(`          </sbol:Range>` + "\n")
		sb.WriteString(`        </sbol:location>` + "\n")
		sb.WriteString(`        <sbol:component>` + "\n")
		sb.WriteString(fmt.Sprintf(`          <sbol:Component rdf:about="http://relix.local/comp_%s">`, gid) + "\n")
		sb.WriteString(fmt.Sprintf(`            <sbol:displayId>comp_%s</sbol:displayId>`, gid) + "\n")
		sb.WriteString(fmt.Sprintf(`            <sbol:definition rdf:resource="http://relix.local/cd_%s"/>`, gid) + "\n")
		sb.WriteString(`          </sbol:Component>` + "\n")
		sb.WriteString(`        </sbol:component>` + "\n")
		sb.WriteString(`      </sbol:SequenceAnnotation>` + "\n")
		sb.WriteString(`    </sbol:sequenceAnnotation>` + "\n")
	}

	sb.WriteString(`  </sbol:ComponentDefinition>` + "\n\n")

	// --- Plasmid Sequence ---
	sb.WriteString(fmt.Sprintf(`  <sbol:Sequence rdf:about="http://relix.local/%s">`, seqID) + "\n")
	sb.WriteString(fmt.Sprintf(`    <sbol:displayId>%s</sbol:displayId>`, seqID) + "\n")
	sb.WriteString(fmt.Sprintf(`    <sbol:elements>%s</sbol:elements>`, string(c.Genome.DNA)) + "\n")
	sb.WriteString(fmt.Sprintf(`    <sbol:encoding rdf:resource="%s"/>`, sbolEncoding) + "\n")
	sb.WriteString(`  </sbol:Sequence>` + "\n\n")

	// --- Per-gene ComponentDefinitions + CDS Sequences ---
	for i, gene := range c.Genome.Genes {
		gid := fmt.Sprintf("g%d_%s", i, sanitizeID(gene.Name))
		sb.WriteString(fmt.Sprintf(`  <sbol:ComponentDefinition rdf:about="http://relix.local/cd_%s">`, gid) + "\n")
		sb.WriteString(fmt.Sprintf(`    <sbol:displayId>cd_%s</sbol:displayId>`, gid) + "\n")
		sb.WriteString(fmt.Sprintf(`    <sbol:name>%s</sbol:name>`, xmlEscape(gene.Name)) + "\n")
		sb.WriteString(fmt.Sprintf(`    <sbol:description>Product: %s; knockdown: %.0f%%; disabled: %t</sbol:description>`,
			xmlEscape(gene.Product), gene.Knockdown*100, gene.Disabled) + "\n")
		sb.WriteString(fmt.Sprintf(`    <sbol:type rdf:resource="%s"/>`, sbolDnaRegion) + "\n")
		sb.WriteString(fmt.Sprintf(`    <sbol:role rdf:resource="%s"/>`, sbolSOCDS) + "\n")
		sb.WriteString(`  </sbol:ComponentDefinition>` + "\n")

		if gene.CDSEnd <= len(c.Genome.DNA) && gene.CDSStart < gene.CDSEnd {
			cds := string(c.Genome.DNA[gene.CDSStart:gene.CDSEnd])
			cdsSeqID := fmt.Sprintf("relix_cds_%s", gid)
			sb.WriteString(fmt.Sprintf(`  <sbol:Sequence rdf:about="http://relix.local/%s">`, cdsSeqID) + "\n")
			sb.WriteString(fmt.Sprintf(`    <sbol:displayId>%s</sbol:displayId>`, cdsSeqID) + "\n")
			sb.WriteString(fmt.Sprintf(`    <sbol:elements>%s</sbol:elements>`, cds) + "\n")
			sb.WriteString(fmt.Sprintf(`    <sbol:encoding rdf:resource="%s"/>`, sbolEncoding) + "\n")
			sb.WriteString(`  </sbol:Sequence>` + "\n")
		}
		sb.WriteString("\n")
	}

	sb.WriteString(`</rdf:RDF>` + "\n")
	return sb.String()
}

func (u *UI) exportSBOL() {
	d := dialog.NewFileSave(func(w fyne.URIWriteCloser, err error) {
		if err != nil || w == nil {
			return
		}
		defer w.Close()
		fmt.Fprint(w, u.buildSBOL())
	}, u.window)
	d.SetFileName("relix-export.xml")
	d.Show()
}

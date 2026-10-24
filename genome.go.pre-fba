package main

import (
	"fmt"
	"math/rand"
	"strings"
)

const Nucleotides = "ATGC"

type Gene struct {
	Name          string
	PromoterStart int
	PromoterEnd   int
	RBSStart      int
	RBSEnd        int
	CDSStart      int
	CDSEnd        int

	Product    string
	Activators []string
	Repressors []string

	BasalTranscription float64
	ActivationK        float64
	RepressionK        float64
	HillN              float64
	Disabled           bool

	Operon string
	Knockdown float64
	// Inducer regulation (small-molecule inputs, IPTG/aTc/Arabinose)
	Inducers     []string
	InducerModes []string
	InducerKs    []float64

	// Quorum sensing (AHL)
	QuorumReg  bool
	QuorumMode string
	QuorumK    float64

	// AHL production (LuxI-like)
	ProducesAHL bool
	AHLRate     float64
}

func (g *Gene) Length() int { return g.CDSEnd - g.PromoterStart }

type Genome struct {
	DNA   []byte
	Genes []*Gene
	Pathways []PathwayStep
}

func NewGenome() *Genome {
	g := &Genome{}
	// metA: metabolic enzyme. Activated by RegA_TF and AHL, repressed by RegB_TF.
	g.addGene(&Gene{
		Name: "metA", Product: "MetA_enzyme",
		Activators:         []string{"RegA_TF"},
		Repressors:         []string{"RegB_TF"},
		BasalTranscription: 0.02,
		ActivationK:        20.0, RepressionK: 20.0, HillN: 2.0,
		QuorumReg: true, QuorumMode: "activate", QuorumK: 5.0,
	}, "AAAAAAAAAAAAAAAAAAAA", "GGGAGGAGG", "ATGCGCTAGCTAGCTAGCTAGCTAGCTAGCTAGCTAGCTAGCTAGCTAGCTAGCTAGC")
	// regA: autoregulatory TF, also activated by IPTG inducer.
	g.addGene(&Gene{
		Name: "regA", Product: "RegA_TF",
		Activators:         []string{"RegA_TF"},
		BasalTranscription: 0.01,
		ActivationK:        20.0, HillN: 2.5,
		Inducers:     []string{"IPTG"},
		InducerModes: []string{"activate"},
		InducerKs:    []float64{1.0},
	}, "TTTTTTTTTTTTTTTTTTTT", "GGGAGGAGG", "ATGCGCGCGCGCGCGCGCGCGCGCGCGCGCGCGCGCGCGCGCGCGC")
	// regB: activated by MetA_enzyme, represses metA (delayed negative feedback).
	g.addGene(&Gene{
		Name: "regB", Product: "RegB_TF",
		Activators:         []string{"MetA_enzyme"},
		BasalTranscription: 0.01,
		ActivationK:        30.0, HillN: 2.0,
	}, "CCCCCCCCCCCCCCCCCCCC", "GGGAGGAGG", "ATGCGTAGCTAGCTAGCTAGCTAGCTAGCTAGCTAGCTAGCTAGCTAGCTAGCTAG")
	// luxI: AHL synthase, activated by IPTG, also inducible by aTc (repressible).
	g.addGene(&Gene{
		Name: "luxI", Product: "LuxI_enzyme",
		BasalTranscription: 0.02,
		Inducers:     []string{"IPTG", "aTc"},
		InducerModes: []string{"activate", "repress"},
		InducerKs:    []float64{1.0, 1.0},
		ProducesAHL: true, AHLRate: 0.5,
	}, "GGGGGGGGGGGGGGGGGGGG", "GGGAGGAGG", "ATGCATGCATGCATGCATGCATGCATGCATGCATGCATGCATGC")
	return g
}

func (g *Genome) addGene(gene *Gene, promoter, rbs, cds string) {
	gene.PromoterStart = len(g.DNA)
	g.DNA = append(g.DNA, []byte(promoter)...)
	gene.PromoterEnd = len(g.DNA)

	gene.RBSStart = len(g.DNA)
	g.DNA = append(g.DNA, []byte(rbs)...)
	gene.RBSEnd = len(g.DNA)

	gene.CDSStart = len(g.DNA)
	g.DNA = append(g.DNA, []byte(cds)...)
	gene.CDSEnd = len(g.DNA)

	g.Genes = append(g.Genes, gene)
}

func (g *Genome) Clone() *Genome {
	ng := &Genome{DNA: append([]byte{}, g.DNA...)}
	for _, gene := range g.Genes {
		c := *gene
		c.Activators = append([]string{}, gene.Activators...)
		c.Repressors = append([]string{}, gene.Repressors...)
		c.Inducers = append([]string{}, gene.Inducers...)
		c.InducerModes = append([]string{}, gene.InducerModes...)
		c.InducerKs = append([]float64{}, gene.InducerKs...)
		ng.Genes = append(ng.Genes, &c)
	}
	return ng
}

func (g *Genome) RBSStrength(gene *Gene) float64 {
	if gene.RBSEnd <= gene.RBSStart {
		return 0
	}
	s := g.DNA[gene.RBSStart:gene.RBSEnd]
	if len(s) == 0 {
		return 0
	}
	gc := 0
	for _, b := range s {
		if b == 'G' {
			gc++
		}
	}
	return float64(gc) / float64(len(s))
}

func (g *Genome) mutate(rng *rand.Rand) (string, int) {
	if len(g.DNA) == 0 {
		return "none", -1
	}
	pos := rng.Intn(len(g.DNA))
	switch rng.Intn(3) {
	case 0:
		_ = g.DNA[pos]
		g.DNA[pos] = Nucleotides[rng.Intn(4)]
		return "point", pos
	case 1:
		g.DNA = append(g.DNA[:pos], g.DNA[pos+1:]...)
		g.shiftGenes(pos, -1)
		return "deletion", pos
	case 2:
		newBase := Nucleotides[rng.Intn(4)]
		g.DNA = append(g.DNA, 0)
		copy(g.DNA[pos+1:], g.DNA[pos:])
		g.DNA[pos] = newBase
		g.shiftGenes(pos, 1)
		return "insertion", pos
	}
	return "none", -1
}

func (g *Genome) shiftGenes(pos, delta int) {
	for _, gene := range g.Genes {
		if gene.PromoterStart >= pos { gene.PromoterStart += delta }
		if gene.PromoterEnd > pos { gene.PromoterEnd += delta }
		if gene.RBSStart >= pos { gene.RBSStart += delta }
		if gene.RBSEnd > pos { gene.RBSEnd += delta }
		if gene.CDSStart >= pos { gene.CDSStart += delta }
		if gene.CDSEnd > pos { gene.CDSEnd += delta }
	}
	for _, gene := range g.Genes {
		if gene.PromoterStart < 0 { gene.PromoterStart = 0 }
		if gene.PromoterEnd > len(g.DNA) { gene.PromoterEnd = len(g.DNA) }
		if gene.RBSStart < 0 { gene.RBSStart = 0 }
		if gene.RBSEnd > len(g.DNA) { gene.RBSEnd = len(g.DNA) }
		if gene.CDSStart < 0 { gene.CDSStart = 0 }
		if gene.CDSEnd > len(g.DNA) { gene.CDSEnd = len(g.DNA) }
	}
	var alive []*Gene
	for _, gene := range g.Genes {
		if gene.PromoterEnd > gene.PromoterStart &&
			gene.RBSEnd > gene.RBSStart &&
			gene.CDSEnd > gene.CDSStart &&
			gene.CDSEnd <= len(g.DNA) {
			alive = append(alive, gene)
		}
	}
	g.Genes = alive
}

func codonOf(dna []byte) byte {
	if len(dna) < 3 { return '?' }
	h := int(dna[0])*31 + int(dna[1])*7 + int(dna[2])
	return byte('A' + h%20)
}

func (g *Genome) TranslateCDS(gene *Gene) string {
	if gene.CDSEnd > len(g.DNA) || gene.CDSEnd-gene.CDSStart < 3 {
		return ""
	}
	var sb strings.Builder
	for i := gene.CDSStart; i+2 < gene.CDSEnd; i += 3 {
		codon := g.DNA[i : i+3]
		if codon[0] == 'T' && codon[1] == 'A' && (codon[2] == 'A' || codon[2] == 'G') { break }
		if codon[0] == 'T' && codon[1] == 'G' && codon[2] == 'A' { break }
		sb.WriteByte(codonOf(codon))
	}
	return sb.String()
}

func (g *Genome) String() string {
	var sb strings.Builder
	for i, b := range g.DNA {
		for _, gene := range g.Genes {
			if i == gene.PromoterStart { sb.WriteString(fmt.Sprintf("\n[%s|prom] ", gene.Name)) }
			if i == gene.RBSStart { sb.WriteString("|rbs|") }
			if i == gene.CDSStart { sb.WriteString("|cds|") }
			if i == gene.CDSEnd { sb.WriteString(" [end]\n") }
		}
		sb.WriteByte(b)
	}
	sb.WriteString("\n")
	return sb.String()
}

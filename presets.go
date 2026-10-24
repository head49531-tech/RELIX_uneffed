package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

type DNAPreset struct {
	Name         string             `json:"name"`
	Category     string             `json:"category,omitempty"`
	Description  string             `json:"description"`
	DNA          []byte             `json:"dna"`
	Genes        []*Gene            `json:"genes"`
	Pathways     []PathwayStep      `json:"pathways,omitempty"`
	InitMRNA     map[string]float64 `json:"init_mrna,omitempty"`
	InitProteins map[string]float64 `json:"init_proteins,omitempty"`
	Builtin      bool               `json:"-"`
}

type presetSpec struct {
	Name         string
	Category     string
	Description  string
	Genes        []Gene
	InitMRNA     map[string]float64
	InitProteins map[string]float64
}

const (
	presetPromoter = "AAAAAAAAAAAAAAAAAAAA"
	presetRBS      = "GGGAGGAGG"
	presetCDS      = "ATGCGCTAGCTAGCTAGCTAGCTAGCTAGCTAGCTAGCTAGCTAGCTAGCTAGCTAGC"
)

// --- gene helpers ---

func sg(name, prod string, basal float64) Gene {
	return Gene{Name: name, Product: prod, BasalTranscription: basal}
}

func rg(name, prod string, basal, k, n float64, act, rep []string) Gene {
	return Gene{Name: name, Product: prod, BasalTranscription: basal,
		ActivationK: k, RepressionK: k, HillN: n,
		Activators: act, Repressors: rep}
}

func ig(name, prod string, basal float64, inds, modes []string, ks []float64) Gene {
	return Gene{Name: name, Product: prod, BasalTranscription: basal,
		Inducers: inds, InducerModes: modes, InducerKs: ks}
}

func qg(name, prod string, basal, qk float64, qmode string) Gene {
	return Gene{Name: name, Product: prod, BasalTranscription: basal,
		QuorumReg: true, QuorumMode: qmode, QuorumK: qk}
}

func ahg(name, prod string, basal, rate float64) Gene {
	return Gene{Name: name, Product: prod, BasalTranscription: basal,
		ProducesAHL: true, AHLRate: rate}
}

// --- preset helpers ---

func ps(name, cat, desc string, genes ...Gene) presetSpec {
	return presetSpec{Name: name, Category: cat, Description: desc, Genes: genes}
}

func (s presetSpec) withMRNA(m map[string]float64) presetSpec { s.InitMRNA = m; return s }
func (s presetSpec) withProt(m map[string]float64) presetSpec { s.InitProteins = m; return s }

func (s presetSpec) makePreset() *DNAPreset {
	g := &Genome{}
	for i := range s.Genes {
		g.addGene(&s.Genes[i], presetPromoter, presetRBS, presetCDS)
	}
	return &DNAPreset{
		Name:         s.Name,
		Category:     s.Category,
		Description:  s.Description,
		DNA:          append([]byte{}, g.DNA...),
		Genes:        cloneGenes(g.Genes),
		InitMRNA:     s.InitMRNA,
		InitProteins: s.InitProteins,
		Builtin:      true,
	}
}

// --- 100 preset table ---

var presetSpecs = []presetSpec{
	// ===== BASAL (10) =====
	ps("Basal 01", "Basal", "Weak promoter (0.01). Slow basal expression.", sg("g1", "P1", 0.01)),
	ps("Basal 02", "Basal", "Medium promoter (0.1).", sg("g1", "P1", 0.1)),
	ps("Basal 03", "Basal", "Strong promoter (0.5).", sg("g1", "P1", 0.5)),
	ps("Basal 04", "Basal", "Overexpression (2.0). Massive mRNA.", sg("g1", "P1", 2.0)),
	ps("Basal 05", "Basal", "Leaky (0.02). Minimal background.", sg("g1", "P1", 0.02)),
	ps("Basal 06", "Basal", "Two independent genes.", sg("gA", "PA", 0.1), sg("gB", "PB", 0.1)),
	ps("Basal 07", "Basal", "Three independent genes.", sg("gA", "PA", 0.1), sg("gB", "PB", 0.1), sg("gC", "PC", 0.1)),
	ps("Basal 08", "Basal", "Reporter only. No TF cascade.", sg("reporter", "GFP", 0.3)),
	ps("Basal 09", "Basal", "Housekeeping enzyme.", sg("housekeeping", "HK_enzyme", 0.5)),
	ps("Basal 10", "Basal", "Dual reporters.", sg("rfp", "RFP", 0.3), sg("gfp", "GFP", 0.3)),

	// ===== AUTOREGULATION (8) =====
	ps("Auto Weak+", "Autoreg", "Weak positive autoregulation.",
		rg("reg", "Reg_TF", 0.05, 20, 2, []string{"Reg_TF"}, nil)).withMRNA(map[string]float64{"reg": 0.5}),
	ps("Auto Strong+", "Autoreg", "Strong positive autoregulation.",
		rg("reg", "Reg_TF", 0.05, 10, 3, []string{"Reg_TF"}, nil)).withMRNA(map[string]float64{"reg": 2.0}),
	ps("Auto+ Reporter", "Autoreg", "Positive autoreg driving a reporter.",
		rg("reg", "Reg_TF", 0.05, 15, 2, []string{"Reg_TF"}, nil),
		rg("reporter", "GFP", 0.02, 20, 2, []string{"Reg_TF"}, nil)).withMRNA(map[string]float64{"reg": 2.0}),
	ps("Auto Weak-", "Autoreg", "Weak negative autoregulation.",
		rg("reg", "Reg_TF", 0.3, 20, 2, nil, []string{"Reg_TF"})),
	ps("Auto Strong-", "Autoreg", "Strong negative autoregulation.",
		rg("reg", "Reg_TF", 0.3, 10, 3, nil, []string{"Reg_TF"})),
	ps("Auto Fast-", "Autoreg", "Fast negative autoreg (n=4).",
		rg("reg", "Reg_TF", 0.3, 15, 4, nil, []string{"Reg_TF"})),
	ps("Auto Slow-", "Autoreg", "Slow negative autoreg (n=1).",
		rg("reg", "Reg_TF", 0.3, 15, 1, nil, []string{"Reg_TF"})),
	ps("Auto- Reporter", "Autoreg", "Negative autoreg with reporter.",
		rg("reg", "Reg_TF", 0.3, 15, 3, nil, []string{"Reg_TF"}),
		rg("reporter", "GFP", 0.02, 20, 2, []string{"Reg_TF"}, nil)),

	// ===== TOGGLE (8) =====
	ps("Toggle Classic", "Toggle", "2-gene bistable mutual repressor.",
		rg("togA", "TogA_TF", 0.3, 15, 3, nil, []string{"TogB_TF"}),
		rg("togB", "TogB_TF", 0.3, 15, 3, nil, []string{"TogA_TF"})).withMRNA(map[string]float64{"togA": 3.0}),
	ps("Toggle Strong", "Toggle", "High Hill coefficient (n=4) — sharp bistability.",
		rg("togA", "TogA_TF", 0.3, 12, 4, nil, []string{"TogB_TF"}),
		rg("togB", "TogB_TF", 0.3, 12, 4, nil, []string{"TogA_TF"})).withMRNA(map[string]float64{"togA": 3.0}),
	ps("Toggle Weak", "Toggle", "Low Hill (n=2) — softer transition.",
		rg("togA", "TogA_TF", 0.3, 20, 2, nil, []string{"TogB_TF"}),
		rg("togB", "TogB_TF", 0.3, 20, 2, nil, []string{"TogA_TF"})).withMRNA(map[string]float64{"togA": 3.0}),
	ps("Toggle BiasA", "Toggle", "Toggle seeded toward state A.",
		rg("togA", "TogA_TF", 0.3, 15, 3, nil, []string{"TogB_TF"}),
		rg("togB", "TogB_TF", 0.3, 15, 3, nil, []string{"TogA_TF"})).withMRNA(map[string]float64{"togA": 5.0}),
	ps("Toggle BiasB", "Toggle", "Toggle seeded toward state B.",
		rg("togA", "TogA_TF", 0.3, 15, 3, nil, []string{"TogB_TF"}),
		rg("togB", "TogB_TF", 0.3, 15, 3, nil, []string{"TogA_TF"})).withMRNA(map[string]float64{"togB": 5.0}),
	ps("Toggle Asym", "Toggle", "Asymmetric toggle — A dominates.",
		rg("togA", "TogA_TF", 0.5, 15, 3, nil, []string{"TogB_TF"}),
		rg("togB", "TogB_TF", 0.2, 15, 3, nil, []string{"TogA_TF"})).withMRNA(map[string]float64{"togA": 3.0}),
	ps("Toggle + Report", "Toggle", "Toggle with dual reporters.",
		rg("togA", "TogA_TF", 0.3, 15, 3, nil, []string{"TogB_TF"}),
		rg("togB", "TogB_TF", 0.3, 15, 3, nil, []string{"TogA_TF"}),
		rg("gfp", "GFP", 0.02, 20, 2, []string{"TogA_TF"}, nil),
		rg("rfp", "RFP", 0.02, 20, 2, []string{"TogB_TF"}, nil)).withMRNA(map[string]float64{"togA": 3.0}),
	ps("Triple Toggle", "Toggle", "3-gene cyclic mutual repression.",
		rg("gA", "GA_TF", 0.3, 15, 3, nil, []string{"GB_TF"}),
		rg("gB", "GB_TF", 0.3, 15, 3, nil, []string{"GC_TF"}),
		rg("gC", "GC_TF", 0.3, 15, 3, nil, []string{"GA_TF"})).withMRNA(map[string]float64{"gA": 3.0}),

	// ===== OSCILLATOR (10) =====
	ps("Repressilator", "Oscillator", "Classic 3-gene ring oscillator (Elowitz 2000).",
		rg("cI", "cI_TF", 0.05, 15, 3, nil, []string{"TetR_TF"}),
		rg("LacI", "LacI_TF", 0.05, 15, 3, nil, []string{"cI_TF"}),
		rg("TetR", "TetR_TF", 0.05, 15, 3, nil, []string{"LacI_TF"})).withMRNA(map[string]float64{"cI": 3.0, "LacI": 0.1, "TetR": 0.1}),
	ps("Repress Slow", "Oscillator", "Repressilator with softer Hill (n=2).",
		rg("cI", "cI_TF", 0.05, 20, 2, nil, []string{"TetR_TF"}),
		rg("LacI", "LacI_TF", 0.05, 20, 2, nil, []string{"cI_TF"}),
		rg("TetR", "TetR_TF", 0.05, 20, 2, nil, []string{"LacI_TF"})).withMRNA(map[string]float64{"cI": 3.0, "LacI": 0.1, "TetR": 0.1}),
	ps("Repress Strong", "Oscillator", "Repressilator with steep Hill (n=4).",
		rg("cI", "cI_TF", 0.05, 12, 4, nil, []string{"TetR_TF"}),
		rg("LacI", "LacI_TF", 0.05, 12, 4, nil, []string{"cI_TF"}),
		rg("TetR", "TetR_TF", 0.05, 12, 4, nil, []string{"LacI_TF"})).withMRNA(map[string]float64{"cI": 3.0, "LacI": 0.1, "TetR": 0.1}),
	ps("Ring 4", "Oscillator", "4-gene repressor ring.",
		rg("g1", "G1_TF", 0.05, 15, 3, nil, []string{"G4_TF"}),
		rg("g2", "G2_TF", 0.05, 15, 3, nil, []string{"G1_TF"}),
		rg("g3", "G3_TF", 0.05, 15, 3, nil, []string{"G2_TF"}),
		rg("g4", "G4_TF", 0.05, 15, 3, nil, []string{"G3_TF"})).withMRNA(map[string]float64{"g1": 3.0}),
	ps("Ring 5", "Oscillator", "5-gene repressor ring.",
		rg("g1", "G1_TF", 0.05, 15, 3, nil, []string{"G5_TF"}),
		rg("g2", "G2_TF", 0.05, 15, 3, nil, []string{"G1_TF"}),
		rg("g3", "G3_TF", 0.05, 15, 3, nil, []string{"G2_TF"}),
		rg("g4", "G4_TF", 0.05, 15, 3, nil, []string{"G3_TF"}),
		rg("g5", "G5_TF", 0.05, 15, 3, nil, []string{"G4_TF"})).withMRNA(map[string]float64{"g1": 3.0}),
	ps("Ring 6", "Oscillator", "6-gene repressor ring.",
		rg("g1", "G1_TF", 0.05, 15, 3, nil, []string{"G6_TF"}),
		rg("g2", "G2_TF", 0.05, 15, 3, nil, []string{"G1_TF"}),
		rg("g3", "G3_TF", 0.05, 15, 3, nil, []string{"G2_TF"}),
		rg("g4", "G4_TF", 0.05, 15, 3, nil, []string{"G3_TF"}),
		rg("g5", "G5_TF", 0.05, 15, 3, nil, []string{"G4_TF"}),
		rg("g6", "G6_TF", 0.05, 15, 3, nil, []string{"G5_TF"})).withMRNA(map[string]float64{"g1": 3.0}),
	ps("Relaxation", "Oscillator", "Relaxation oscillator: fast activator + slow repressor.",
		rg("act", "Act_TF", 0.05, 15, 2, []string{"Act_TF"}, nil),
		rg("rep", "Rep_TF", 0.02, 25, 3, []string{"Act_TF"}, nil),
		rg("target", "Target", 0.02, 20, 2, []string{"Act_TF"}, []string{"Rep_TF"})).withMRNA(map[string]float64{"act": 2.0}),
	ps("Activ Osc", "Oscillator", "Activator-based 2-gene oscillator with mutual antagonism.",
		rg("activ", "Activ_TF", 0.05, 15, 3, []string{"Activ_TF"}, []string{"Repr_TF"}),
		rg("repr", "Repr_TF", 0.05, 20, 2, []string{"Activ_TF"}, []string{"Activ_TF"})).withMRNA(map[string]float64{"activ": 2.0}),
	ps("Amp Osc", "Oscillator", "Ring oscillator with amplified output.",
		rg("cI", "cI_TF", 0.05, 15, 3, nil, []string{"TetR_TF"}),
		rg("LacI", "LacI_TF", 0.05, 15, 3, nil, []string{"cI_TF"}),
		rg("TetR", "TetR_TF", 0.05, 15, 3, nil, []string{"LacI_TF"}),
		rg("out", "Out_TF", 0.02, 20, 2, []string{"cI_TF"}, nil)).withMRNA(map[string]float64{"cI": 3.0, "LacI": 0.1, "TetR": 0.1}),
	ps("Fast Osc", "Oscillator", "Compact 3-gene oscillator tuned for speed.",
		rg("cI", "cI_TF", 0.15, 10, 4, nil, []string{"TetR_TF"}),
		rg("LacI", "LacI_TF", 0.15, 10, 4, nil, []string{"cI_TF"}),
		rg("TetR", "TetR_TF", 0.15, 10, 4, nil, []string{"LacI_TF"})).withMRNA(map[string]float64{"cI": 3.0}),

	// ===== LOGIC (10) =====
	ps("AND Gate", "Logic", "Two activators required for output.",
		rg("inA", "InA_TF", 0.5, 20, 2, nil, nil),
		rg("inB", "InB_TF", 0.5, 20, 2, nil, nil),
		rg("out", "Out", 0.02, 15, 3, []string{"InA_TF", "InB_TF"}, nil)),
	ps("OR Gate", "Logic", "Either activator drives output.",
		rg("inA", "InA_TF", 0.5, 20, 2, nil, nil),
		rg("inB", "InB_TF", 0.5, 20, 2, nil, nil),
		rg("outA", "OutA", 0.02, 15, 2, []string{"InA_TF"}, nil),
		rg("outB", "OutB", 0.02, 15, 2, []string{"InB_TF"}, nil)),
	ps("NOT Gate", "Logic", "Constitutive input represses output.",
		sg("in", "In_TF", 0.5),
		rg("out", "Out", 0.5, 15, 3, nil, []string{"In_TF"})),
	ps("NAND Gate", "Logic", "AND inverted by repressor.",
		rg("inA", "InA_TF", 0.5, 20, 2, nil, nil),
		rg("inB", "InB_TF", 0.5, 20, 2, nil, nil),
		rg("andOut", "AndOut", 0.02, 15, 3, []string{"InA_TF", "InB_TF"}, nil),
		rg("nandOut", "NandOut", 0.5, 15, 3, nil, []string{"AndOut"})),
	ps("NOR Gate", "Logic", "NOR combiner.",
		rg("inA", "InA_TF", 0.5, 20, 2, nil, nil),
		rg("inB", "InB_TF", 0.5, 20, 2, nil, nil),
		rg("orOut", "OrOut", 0.02, 15, 2, []string{"InA_TF", "InB_TF"}, nil),
		rg("norOut", "NorOut", 0.5, 15, 3, nil, []string{"OrOut"})),
	ps("XOR Circuit", "Logic", "XOR via two parallel ANDs with cross-repression.",
		rg("inA", "InA_TF", 0.5, 20, 2, nil, nil),
		rg("inB", "InB_TF", 0.5, 20, 2, nil, nil),
		rg("notA", "NotA", 0.5, 15, 3, nil, []string{"InA_TF"}),
		rg("notB", "NotB", 0.5, 15, 3, nil, []string{"InB_TF"}),
		rg("andA", "AndA", 0.02, 15, 3, []string{"InA_TF", "NotB"}, nil),
		rg("andB", "AndB", 0.02, 15, 3, []string{"InB_TF", "NotA"}, nil)),
	ps("Buffer", "Logic", "Signal buffer — input drives output.",
		rg("in", "In_TF", 0.5, 20, 2, nil, nil),
		rg("out", "Out", 0.02, 15, 3, []string{"In_TF"}, nil)),
	ps("Inverter 2", "Logic", "Two-stage inverter chain.",
		sg("in", "In_TF", 0.5),
		rg("not1", "Not1", 0.5, 15, 3, nil, []string{"In_TF"}),
		rg("out", "Out", 0.5, 15, 3, nil, []string{"Not1"})),
	ps("Inverter 3", "Logic", "Three-stage inverter chain.",
		sg("in", "In_TF", 0.5),
		rg("n1", "N1", 0.5, 15, 3, nil, []string{"In_TF"}),
		rg("n2", "N2", 0.5, 15, 3, nil, []string{"N1"}),
		rg("n3", "N3", 0.5, 15, 3, nil, []string{"N2"})),
	ps("Half Adder", "Logic", "Boolean half-adder (sum + carry).",
		rg("a", "A_TF", 0.5, 20, 2, nil, nil),
		rg("b", "B_TF", 0.5, 20, 2, nil, nil),
		rg("notA", "NotA", 0.5, 15, 3, nil, []string{"A_TF"}),
		rg("notB", "NotB", 0.5, 15, 3, nil, []string{"B_TF"}),
		rg("sum1", "Sum1", 0.02, 15, 3, []string{"A_TF", "NotB"}, nil),
		rg("sum2", "Sum2", 0.02, 15, 3, []string{"NotA", "B_TF"}, nil),
		rg("carry", "Carry", 0.02, 15, 3, []string{"A_TF", "B_TF"}, nil)),

	// ===== FFL (6) =====
	ps("FFL Coherent", "FFL", "Coherent FFL: A → B and A+B → C.",
		rg("A", "A_TF", 0.3, 20, 2, nil, nil),
		rg("B", "B_TF", 0.02, 20, 2, []string{"A_TF"}, nil),
		rg("C", "C_out", 0.02, 15, 3, []string{"A_TF", "B_TF"}, nil)).withMRNA(map[string]float64{"A": 1.0}),
	ps("FFL Incoherent", "FFL", "Incoherent FFL: A+B → C, B represses C.",
		rg("A", "A_TF", 0.3, 20, 2, nil, nil),
		rg("B", "B_TF", 0.05, 20, 2, []string{"A_TF"}, nil),
		rg("C", "C_out", 0.02, 15, 3, []string{"A_TF", "B_TF"}, []string{"B_TF"})).withMRNA(map[string]float64{"A": 1.0}),
	ps("FFL Buffer", "FFL", "FFL used as signal buffer.",
		rg("in", "In_TF", 0.3, 20, 2, nil, nil),
		rg("mid", "Mid", 0.02, 20, 2, []string{"In_TF"}, nil),
		rg("out", "Out", 0.02, 15, 3, []string{"In_TF", "Mid"}, nil)),
	ps("FFL Pulse", "FFL", "Incoherent FFL generates pulses.",
		rg("s", "S_TF", 0.3, 20, 2, nil, nil),
		rg("f", "F_TF", 0.05, 20, 2, []string{"S_TF"}, nil),
		rg("pulse", "Pulse", 0.02, 15, 3, []string{"S_TF"}, []string{"F_TF"})).withMRNA(map[string]float64{"s": 1.0}),
	ps("FFL Delay", "FFL", "Chained FFL creates delay.",
		rg("in", "In_TF", 0.3, 20, 2, nil, nil),
		rg("m1", "M1", 0.02, 20, 2, []string{"In_TF"}, nil),
		rg("m2", "M2", 0.02, 20, 2, []string{"M1"}, nil),
		rg("out", "Out", 0.02, 15, 3, []string{"In_TF", "M2"}, nil)),
	ps("FFL Band", "FFL", "Band detector — output peaks at intermediate input.",
		rg("in", "In_TF", 0.05, 20, 2, nil, nil),
		rg("amplifier", "Amp_TF", 0.02, 15, 2, []string{"In_TF"}, nil),
		rg("attenuator", "Att_TF", 0.02, 25, 2, []string{"In_TF"}, nil),
		rg("band", "Band", 0.02, 15, 3, []string{"Amp_TF"}, []string{"Att_TF"})),

	// ===== QUORUM (6) =====
	ps("Quorum Sensor", "Quorum", "LuxI + AHL-activated response.",
		ahg("luxI", "LuxI_enzyme", 0.5, 0.5),
		qg("resp", "Resp", 0.02, 5, "activate")),
	ps("Quorum Report", "Quorum", "AHL drives a reporter.",
		ahg("luxI", "LuxI_enzyme", 0.5, 0.5),
		qg("gfp", "GFP", 0.02, 5, "activate")),
	ps("Quorum Kill", "Quorum", "AHL activates a toxic output.",
		ahg("luxI", "LuxI_enzyme", 0.5, 0.5),
		qg("toxin", "Toxin", 0.01, 5, "activate")),
	ps("Quorum Timer", "Quorum", "AHL slowly activates a reporter.",
		ahg("luxI", "LuxI_enzyme", 0.3, 0.3),
		qg("gfp", "GFP", 0.01, 8, "activate")),
	ps("Quorum Suppress", "Quorum", "AHL represses an existing gene.",
		sg("baseline", "Baseline", 0.5),
		ahg("luxI", "LuxI_enzyme", 0.5, 0.5),
		qg("resp", "Resp", 0.5, 5, "repress")),
	ps("Quorum Pulse", "Quorum", "Transient AHL-induced reporter pulse.",
		ahg("luxI", "LuxI_enzyme", 0.3, 0.3),
		qg("amp", "Amp_TF", 0.02, 5, "activate"),
		qg("repressor", "Rep_TF", 0.02, 8, "activate"),
		rg("pulse", "Pulse", 0.02, 15, 3, []string{"Amp_TF"}, []string{"Rep_TF"})),

	// ===== INDUCER (10) =====
	ps("IPTG-Inducible", "Inducer", "IPTG activates gene.", ig("g1", "P1", 0.02, []string{"IPTG"}, []string{"activate"}, []float64{1})),
	ps("aTc-Repressible", "Inducer", "aTc represses gene.", ig("g1", "P1", 0.5, []string{"aTc"}, []string{"repress"}, []float64{1})),
	ps("Ara-Inducible", "Inducer", "Arabinose activates gene.", ig("g1", "P1", 0.02, []string{"Arabinose"}, []string{"activate"}, []float64{1})),
	ps("IPTG + aTc", "Inducer", "Two-inducer: IPTG on, aTc off.", ig("g1", "P1", 0.02, []string{"IPTG", "aTc"}, []string{"activate", "repress"}, []float64{1, 1})),
	ps("Ara + IPTG", "Inducer", "Dual activators.", ig("g1", "P1", 0.02, []string{"Arabinose", "IPTG"}, []string{"activate", "activate"}, []float64{1, 1})),
	ps("IPTG Cascade", "Inducer", "IPTG-induced TF activates reporter.",
		ig("iptgTF", "IptgTF", 0.02, []string{"IPTG"}, []string{"activate"}, []float64{1}),
		rg("reporter", "GFP", 0.02, 15, 3, []string{"IptgTF"}, nil)),
	ps("IPTG Report", "Inducer", "Direct IPTG reporter.", ig("gfp", "GFP", 0.02, []string{"IPTG"}, []string{"activate"}, []float64{1})),
	ps("aTc Feedback", "Inducer", "aTc-repressed gene autoactivates.",
		rg("reg", "Reg_TF", 0.02, 15, 2, []string{"Reg_TF"}, nil)),
	ps("IPTG Feedback", "Inducer", "IPTG-induced autoregulatory gene.", ig("reg", "Reg_TF", 0.02, []string{"IPTG"}, []string{"activate"}, []float64{1})),
	ps("All Inducers", "Inducer", "Combinatorial inducer circuit.", ig("g1", "P1", 0.02, []string{"IPTG", "aTc", "Arabinose"}, []string{"activate", "repress", "activate"}, []float64{1, 1, 1})),

	// ===== METABOLISM (8) =====
	ps("MetA Basic", "Metabolism", "Single metA gene.", sg("metA", "MetA_enzyme", 0.5)),
	ps("MetA Strong", "Metabolism", "Overexpressed metA.", sg("metA", "MetA_enzyme", 1.5)),
	ps("Dual Enzyme", "Metabolism", "Two metabolic enzymes.",
		sg("metA", "MetA_enzyme", 0.5),
		sg("metB", "MetB_enzyme", 0.5)),
	ps("ATP Synthase", "Metabolism", "Dedicated ATP synthesizer.",
		sg("atp", "ATP_synthase", 1.0)),
	ps("Citrate Cycle", "Metabolism", "Citrate + metA dual pathway.",
		sg("metA", "MetA_enzyme", 0.5),
		sg("cit", "Citrate_enzyme", 0.5)),
	ps("Sugar Import", "Metabolism", "Glucose importer protein.",
		sg("imp", "Importer", 1.0)),
	ps("Stress Circuit", "Metabolism", "MetA + stress sensor.",
		sg("metA", "MetA_enzyme", 0.5),
		rg("sensor", "Sensor_TF", 0.02, 20, 2, []string{"MetA_enzyme"}, nil)),
	ps("Catabolite Rep", "Metabolism", "catabolite-like repression of metA.",
		sg("sig", "Sig_TF", 0.5),
		rg("metA", "MetA_enzyme", 0.5, 15, 3, nil, []string{"Sig_TF"})),

	// ===== STRESS (6) =====
	ps("Heat Shock", "Stress", "Constitutive stress responder.", sg("hsp", "HSP_protein", 0.5)),
	ps("SOS Response", "Stress", "Toxin-antitoxin-like SOS sensor.",
		sg("recA", "RecA_TF", 0.5),
		rg("sos", "SOS_out", 0.02, 15, 2, []string{"RecA_TF"}, nil)),
	ps("Oxidative", "Stress", "Oxidant-responsive reporter.",
		rg("oxR", "OxR_TF", 0.02, 10, 2, nil, nil),
		rg("oxOut", "OxOut", 0.02, 15, 2, []string{"OxR_TF"}, nil)),
	ps("Cold Shock", "Stress", "Cold-shock protein producer.", sg("csp", "CSP_protein", 0.5)),
	ps("pH Stress", "Stress", "Acid-responsive TF.",
		rg("asR", "AsR_TF", 0.02, 8, 2, nil, nil)),
	ps("Toxin-Antitox", "Stress", "Toxin plus antitoxin pair.",
		sg("tox", "Toxin", 0.5),
		rg("anti", "Antitoxin", 0.5, 15, 2, []string{"Toxin"}, nil)),

	// ===== NAMED (8) =====
	ps("Elowitz Rep", "Named", "Elowitz-Leibler 2000 repressilator.",
		rg("cI", "cI_TF", 0.05, 15, 3, nil, []string{"TetR_TF"}),
		rg("LacI", "LacI_TF", 0.05, 15, 3, nil, []string{"cI_TF"}),
		rg("TetR", "TetR_TF", 0.05, 15, 3, nil, []string{"LacI_TF"})).withMRNA(map[string]float64{"cI": 3, "LacI": 0.1, "TetR": 0.1}),
	ps("Gardner Toggle", "Named", "Gardner-Cantor-Collins 2000 toggle switch.",
		rg("LacI", "LacI_TF", 0.3, 15, 3, nil, []string{"cI_TF"}),
		rg("cI", "cI_TF", 0.3, 15, 3, nil, []string{"LacI_TF"})).withMRNA(map[string]float64{"LacI": 3}),
	ps("Collins Toggle", "Named", "Collins lab toggle variant.",
		rg("cI", "cI_TF", 0.4, 10, 4, nil, []string{"TetR_TF"}),
		rg("TetR", "TetR_TF", 0.4, 10, 4, nil, []string{"cI_TF"})).withMRNA(map[string]float64{"cI": 3}),
	ps("Weiss Osc", "Named", "Weiss lab ring oscillator.",
		rg("g1", "G1_TF", 0.1, 12, 3, nil, []string{"G3_TF"}),
		rg("g2", "G2_TF", 0.1, 12, 3, nil, []string{"G1_TF"}),
		rg("g3", "G3_TF", 0.1, 12, 3, nil, []string{"G2_TF"})).withMRNA(map[string]float64{"g1": 3}),
	ps("Basu Quorum", "Named", "Basu-Weiss quorum sender-receiver.",
		ahg("luxI", "LuxI_enzyme", 0.5, 0.5),
		qg("recv", "Recv", 0.02, 5, "activate")),
	ps("Yokobayashi FFL", "Named", "Yokobayashi FFL pulse generator.",
		rg("s", "S_TF", 0.3, 20, 2, nil, nil),
		rg("f", "F_TF", 0.05, 20, 2, []string{"S_TF"}, nil),
		rg("out", "Out", 0.02, 15, 3, []string{"S_TF"}, []string{"F_TF"})).withMRNA(map[string]float64{"s": 1}),
	ps("Endy Osc", "Named", "Endy-lab-derived oscillator.",
		rg("g1", "G1_TF", 0.1, 15, 3, nil, []string{"G3_TF"}),
		rg("g2", "G2_TF", 0.1, 15, 3, nil, []string{"G1_TF"}),
		rg("g3", "G3_TF", 0.1, 15, 3, nil, []string{"G2_TF"})).withMRNA(map[string]float64{"g1": 3}),
	ps("Hasty Osc", "Named", "Hasty-lab-style dual-feedback oscillator.",
		rg("act", "Act_TF", 0.05, 15, 2, []string{"Act_TF"}, nil),
		rg("rep", "Rep_TF", 0.02, 25, 3, []string{"Act_TF"}, nil),
		rg("target", "Target", 0.02, 20, 2, []string{"Act_TF"}, []string{"Rep_TF"})).withMRNA(map[string]float64{"act": 2}),

	// ===== COMPOSITE (10) =====
	ps("Cascade 2", "Composite", "2-stage cascade.",
		rg("in", "In_TF", 0.5, 20, 2, nil, nil),
		rg("out", "Out", 0.02, 15, 3, []string{"In_TF"}, nil)),
	ps("Cascade 3", "Composite", "3-stage cascade.",
		rg("in", "In_TF", 0.5, 20, 2, nil, nil),
		rg("mid", "Mid", 0.02, 20, 2, []string{"In_TF"}, nil),
		rg("out", "Out", 0.02, 20, 2, []string{"Mid"}, nil)),
	ps("Cascade 4", "Composite", "4-stage cascade.",
		rg("in", "In_TF", 0.5, 20, 2, nil, nil),
		rg("m1", "M1", 0.02, 20, 2, []string{"In_TF"}, nil),
		rg("m2", "M2", 0.02, 20, 2, []string{"M1"}, nil),
		rg("out", "Out", 0.02, 20, 2, []string{"M2"}, nil)),
	ps("Amplifier", "Composite", "Cascade amplifier — small input, large output.",
		rg("in", "In_TF", 0.05, 20, 2, nil, nil),
		rg("amp1", "Amp1", 0.02, 15, 3, []string{"In_TF"}, nil),
		rg("amp2", "Amp2", 0.02, 15, 3, []string{"Amp1"}, nil)),
	ps("Attenuator", "Composite", "Attenuator — signal dampening chain.",
		rg("in", "In_TF", 0.5, 20, 2, nil, nil),
		rg("att1", "Att1", 0.5, 15, 3, nil, []string{"In_TF"}),
		rg("att2", "Att2", 0.5, 15, 3, nil, []string{"Att1"})),
	ps("Differentiator", "Composite", "Differentiator — responds to changes.",
		rg("in", "In_TF", 0.3, 20, 2, nil, nil),
		rg("fast", "Fast_TF", 0.02, 15, 2, []string{"In_TF"}, nil),
		rg("slow", "Slow_TF", 0.02, 25, 2, []string{"In_TF"}, nil),
		rg("diff", "Diff", 0.02, 15, 3, []string{"Fast_TF"}, []string{"Slow_TF"})),
	ps("Integrator", "Composite", "Integrator — accumulates input signal.",
		rg("in", "In_TF", 0.1, 20, 2, nil, nil),
		rg("acc", "Acc_TF", 0.02, 10, 1, []string{"In_TF", "Acc_TF"}, nil)),
	ps("Memory Latch", "Composite", "SR-style latch via toggle.",
		rg("set", "Set_TF", 0.3, 20, 2, nil, nil),
		rg("reset", "Reset_TF", 0.3, 20, 2, nil, nil),
		rg("latch", "Latch", 0.3, 15, 3, []string{"Set_TF", "Latch"}, []string{"Reset_TF"})).withMRNA(map[string]float64{"latch": 2}),
	ps("Clock", "Composite", "Slow clock — periodic pulse train.",
		rg("osc1", "Osc1_TF", 0.1, 15, 3, nil, []string{"Osc2_TF"}),
		rg("osc2", "Osc2_TF", 0.1, 15, 3, nil, []string{"Osc1_TF"}),
		rg("clock", "Clock", 0.02, 15, 3, []string{"Osc1_TF"}, []string{"Osc2_TF"})).withMRNA(map[string]float64{"osc1": 2}),
	ps("Randomizer", "Composite", "Cross-coupled toggle + noise-ish reporter.",
		rg("n1", "N1_TF", 0.3, 10, 3, nil, []string{"N2_TF"}),
		rg("n2", "N2_TF", 0.3, 10, 3, nil, []string{"N1_TF"}),
		rg("reporter", "GFP", 0.02, 15, 2, []string{"N1_TF"}, nil)),
}

// --- registry & IO ---

func presetsDir() string {
	base, err := os.UserConfigDir()
	if err != nil || base == "" {
		base = "."
	}
	return filepath.Join(base, "relix", "presets")
}

func (p *DNAPreset) filePath() string {
	return filepath.Join(presetsDir(), sanitizeID(p.Name)+".json")
}

func cloneGenes(src []*Gene) []*Gene {
	out := make([]*Gene, 0, len(src))
	for _, g := range src {
		c := *g
		c.Activators = append([]string{}, g.Activators...)
		c.Repressors = append([]string{}, g.Repressors...)
		c.Inducers = append([]string{}, g.Inducers...)
		c.InducerModes = append([]string{}, g.InducerModes...)
		c.InducerKs = append([]float64{}, g.InducerKs...)
		out = append(out, &c)
	}
	return out
}

func (p *DNAPreset) genomeFromPreset() *Genome {
	return &Genome{
		DNA:      append([]byte{}, p.DNA...),
		Genes:    cloneGenes(p.Genes),
		Pathways: append([]PathwayStep{}, p.Pathways...),
	}
}

func loadUserPresets() ([]*DNAPreset, []error) {
	dir := presetsDir()
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, []error{err}
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, []error{err}
	}
	var out []*DNAPreset
	var errs []error
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			errs = append(errs, err)
			continue
		}
		var p DNAPreset
		if err := json.Unmarshal(data, &p); err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", e.Name(), err))
			continue
		}
		if p.Name == "" {
			p.Name = strings.TrimSuffix(e.Name(), ".json")
		}
		p.Builtin = false
		out = append(out, &p)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, errs
}

func saveUserPreset(p *DNAPreset) error {
	dir := presetsDir()
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(p.filePath(), data, 0644)
}

func builtinFullGRN() *DNAPreset {
	g := NewGenome()
	return &DNAPreset{
		Name: "Full GRN", Category: "Default",
		Description: "Default 4-gene network with inducer + quorum.",
		DNA:         g.DNA, Genes: cloneGenes(g.Genes), Builtin: true,
	}
}

func builtinPresets() []*DNAPreset {
	out := make([]*DNAPreset, 0, len(presetSpecs)+1)
	out = append(out, builtinFullGRN())
	for _, spec := range presetSpecs {
		out = append(out, spec.makePreset())
	}
	return out
}

// --- UI methods (unchanged from before) ---

func (u *UI) refreshPresetsList() {
	builtins := builtinPresets()
	users, _ := loadUserPresets()
	u.presetsCache = append(builtins, users...)
	u.applyPresetFilters()
}

func (u *UI) selectedPreset() *DNAPreset {
	if u.selectedPresetIdx < 0 || u.selectedPresetIdx >= len(u.presetsFiltered) {
		return nil
	}
	return u.presetsFiltered[u.selectedPresetIdx]
}

func (u *UI) refreshPresetPreview() {
	if u.presetPreview == nil {
		return
	}
	p := u.selectedPreset()
	if p == nil {
		u.presetPreview.SetText("Select a preset from the list.")
		return
	}
	tag := "user"
	if p.Builtin {
		tag = "built-in"
	}
	var sb strings.Builder
	badge := u.presetBadge(p.Name)
	fmt.Fprintf(&sb, "%s [%s, %s] %s\n\n", p.Name, p.Category, tag, badge)
	fmt.Fprintf(&sb, "%s\n\n", p.Description)
	fmt.Fprintf(&sb, "Genes:     %d\n", len(p.Genes))
	fmt.Fprintf(&sb, "DNA:       %d bp\n", len(p.DNA))
	fmt.Fprintf(&sb, "Pathways:  %d\n", len(p.Pathways))
	if len(p.InitMRNA) > 0 {
		fmt.Fprintf(&sb, "Seed mRNA: %v\n", p.InitMRNA)
	}
	if len(p.InitProteins) > 0 {
		fmt.Fprintf(&sb, "Seed prot: %v\n", p.InitProteins)
	}
	if len(p.Genes) > 0 {
		sb.WriteString("\nGenes:\n")
		for _, g := range p.Genes {
			fmt.Fprintf(&sb, "  %-10s → %-14s (basal %.2f)\n", g.Name, g.Product, g.BasalTranscription)
		}
	}
	u.presetPreview.SetText(sb.String())
}

func (u *UI) saveCurrentAsPreset() {
	c := u.getSelectedCell()
	if c == nil {
		dialog.ShowInformation("Save preset", "Select a cell first.", u.window)
		return
	}
	nameIn := widget.NewEntry()
	nameIn.SetPlaceHolder("My circuit")
	descIn := widget.NewMultiLineEntry()
	descIn.SetPlaceHolder("Optional description")

	dialog.ShowForm("Save current cell as preset", "Save", "Cancel",
		[]*widget.FormItem{
			{Text: "Name", Widget: nameIn},
			{Text: "Description", Widget: descIn},
		}, func(ok bool) {
			if !ok {
				return
			}
			name := strings.TrimSpace(nameIn.Text)
			if name == "" {
				dialog.ShowInformation("Save preset", "Name required.", u.window)
				return
			}
			p := &DNAPreset{
				Name:         name,
				Description:  strings.TrimSpace(descIn.Text),
				DNA:          append([]byte{}, c.Genome.DNA...),
				Genes:        cloneGenes(c.Genome.Genes),
				Pathways:     append([]PathwayStep{}, c.Genome.Pathways...),
				InitMRNA:     map[string]float64{},
				InitProteins: map[string]float64{},
			}
			for k, v := range c.MRNA {
				p.InitMRNA[k] = v
			}
			for k, v := range c.Proteins {
				p.InitProteins[k] = v
			}
			if err := saveUserPreset(p); err != nil {
				dialog.ShowError(err, u.window)
				return
			}
			u.refreshPresetsList()
			dialog.ShowInformation("Saved", "Preset saved as "+p.Name, u.window)
		}, u.window)
}

func (u *UI) loadPresetIntoCurrentCell() {
	p := u.selectedPreset()
	if p == nil {
		dialog.ShowInformation("Load preset", "Select a preset first.", u.window)
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
	dialog.ShowInformation("Loaded", fmt.Sprintf("Preset '%s' loaded into cell #%d", p.Name, c.ID), u.window)
}

func (u *UI) spawnCellFromPreset() {
	p := u.selectedPreset()
	if p == nil {
		dialog.ShowInformation("Spawn cell", "Select a preset first.", u.window)
		return
	}
	if len(u.sim.Cells) >= u.sim.Params.MaxCells {
		dialog.ShowInformation("Cannot spawn", "Max cells reached.", u.window)
		return
	}
	c := &Cell{
		ID:          u.sim.NextID,
		Generation:  0,
		Alive:       true,
		Genome:      p.genomeFromPreset(),
		MRNA:        map[string]float64{},
		Proteins:    map[string]float64{},
		Metabolites: map[string]float64{},
		Glucose:     10,
		ATP:         5,
	}
	u.sim.NextID++
	for k, v := range p.InitMRNA {
		c.MRNA[k] = v
	}
	for k, v := range p.InitProteins {
		c.Proteins[k] = v
	}
	u.sim.Cells = append(u.sim.Cells, c)
	u.selectedCell = c.ID
	u.render()
	dialog.ShowInformation("Spawned", fmt.Sprintf("New cell #%d from preset '%s'", c.ID, p.Name), u.window)
}

func (u *UI) deleteSelectedPreset() {
	p := u.selectedPreset()
	if p == nil {
		dialog.ShowInformation("Delete preset", "Select a preset first.", u.window)
		return
	}
	if p.Builtin {
		dialog.ShowInformation("Delete preset", "Built-in presets cannot be deleted.", u.window)
		return
	}
	dialog.ShowConfirm("Delete preset", "Delete '"+p.Name+"'?", func(ok bool) {
		if !ok {
			return
		}
		if err := os.Remove(p.filePath()); err != nil {
			dialog.ShowError(err, u.window)
			return
		}
		u.selectedPresetIdx = -1
		u.refreshPresetsList()
	}, u.window)
}

func (u *UI) buildPresetsTab() fyne.CanvasObject {
	saveBtn := widget.NewButton("Save current cell as preset…", u.saveCurrentAsPreset)
	refreshBtn := widget.NewButton("Refresh", u.refreshPresetsList)
	folderBtn := widget.NewButton("Show folder path", func() {
		dialog.ShowInformation("Presets folder", presetsDir(), u.window)
	})

	u.presetList = widget.NewList(
		func() int { return len(u.presetsFiltered) },
		func() fyne.CanvasObject { return widget.NewLabel("template") },
		func(i widget.ListItemID, o fyne.CanvasObject) {
			if i < 0 || i >= len(u.presetsFiltered) {
				return
			}
			p := u.presetsFiltered[i]
			tag := "(user)"
			if p.Builtin {
				tag = "(built-in)"
			}
			o.(*widget.Label).SetText(fmt.Sprintf("%-13s %-10s %s %s", tag, "["+p.Category+"]", p.Name, u.presetBadge(p.Name)))
		},
	)
	u.presetList.OnSelected = func(id widget.ListItemID) {
		if id < 0 || id >= len(u.presetsFiltered) {
			return
		}
		u.selectedPresetIdx = id
		u.refreshPresetPreview()
	}

	u.presetPreview = widget.NewLabel("Select a preset from the list.")
	u.presetPreview.Wrapping = fyne.TextWrapWord
	u.presetPreview.TextStyle = fyne.TextStyle{Monospace: true}

	loadBtn := widget.NewButton("Load into selected cell", u.loadPresetIntoCurrentCell)
	spawnBtn := widget.NewButton("Spawn new cell from preset", u.spawnCellFromPreset)
	deleteBtn := widget.NewButton("Delete preset", u.deleteSelectedPreset)

	screenBtn := widget.NewButton("Screen behaviors", u.screenAllPresets)

	u.presetPinALabel = mono("A: (unset)")
	u.presetPinBLabel = mono("B: (unset)")
	pinA := widget.NewButton("Pin A", func() {
		if p := u.selectedPreset(); p != nil {
			u.presetPinA = p
			u.presetPinALabel.SetText("A: " + p.Name)
		}
	})
	pinB := widget.NewButton("Pin B", func() {
		if p := u.selectedPreset(); p != nil {
			u.presetPinB = p
			u.presetPinBLabel.SetText("B: " + p.Name)
		}
	})
	compareBtn := widget.NewButton("Compare A vs B", u.showPresetCompare)
	editBtn := widget.NewButton("Edit circuit…", func() {
		if p := u.selectedPreset(); p != nil {
			u.openCircuitEditor(p)
		}
	})

	filterBar := u.buildPresetFilterBar()

	top := container.NewVBox(
		container.NewHBox(saveBtn, refreshBtn, folderBtn, screenBtn),
		filterBar,
		widget.NewSeparator(),
	)
	actions := container.NewVBox(
		container.NewHBox(loadBtn, spawnBtn, deleteBtn, editBtn),
		container.NewHBox(pinA, pinB, compareBtn, u.presetPinALabel, widget.NewLabel("  "), u.presetPinBLabel),
	)
	right := container.NewBorder(nil, actions, nil, nil, container.NewVScroll(u.presetPreview))
	split := container.NewHSplit(u.presetList, right)
	split.SetOffset(0.4)

	u.refreshPresetsList()
	return container.NewBorder(top, nil, nil, nil, split)
}

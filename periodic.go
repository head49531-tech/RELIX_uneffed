package main

import "fmt"

// ElementMeta carries periodic-table data for each element.
// Group 0 marks lanthanide/actinide f-block rows.
type ElementMeta struct {
	Symbol            string
	Z                 int
	Period            int
	Group             int
	Block             string
	Electronegativity float64
	OxidationStates   string
}

var elementMeta = []ElementMeta{
	// Period 1
	{"H", 1, 1, 1, "s", 2.20, "-1,+1"},
	{"He", 2, 1, 18, "s", 0.0, "0"},
	// Period 2
	{"Li", 3, 2, 1, "s", 0.98, "+1"},
	{"Be", 4, 2, 2, "s", 1.57, "+2"},
	{"B", 5, 2, 13, "p", 2.04, "+3"},
	{"C", 6, 2, 14, "p", 2.55, "-4,+2,+4"},
	{"N", 7, 2, 15, "p", 3.04, "-3,+3,+5"},
	{"O", 8, 2, 16, "p", 3.44, "-2,-1"},
	{"F", 9, 2, 17, "p", 3.98, "-1"},
	{"Ne", 10, 2, 18, "p", 0.0, "0"},
	// Period 3
	{"Na", 11, 3, 1, "s", 0.93, "+1"},
	{"Mg", 12, 3, 2, "s", 1.31, "+2"},
	{"Al", 13, 3, 13, "p", 1.61, "+3"},
	{"Si", 14, 3, 14, "p", 1.90, "-4,+4"},
	{"P", 15, 3, 15, "p", 2.19, "-3,+3,+5"},
	{"S", 16, 3, 16, "p", 2.58, "-2,+4,+6"},
	{"Cl", 17, 3, 17, "p", 3.16, "-1,+1,+5,+7"},
	{"Ar", 18, 3, 18, "p", 0.0, "0"},
	// Period 4
	{"K", 19, 4, 1, "s", 0.82, "+1"},
	{"Ca", 20, 4, 2, "s", 1.00, "+2"},
	{"Sc", 21, 4, 3, "d", 1.36, "+3"},
	{"Ti", 22, 4, 4, "d", 1.54, "+2,+3,+4"},
	{"V", 23, 4, 5, "d", 1.63, "+2,+3,+4,+5"},
	{"Cr", 24, 4, 6, "d", 1.66, "+2,+3,+6"},
	{"Mn", 25, 4, 7, "d", 1.55, "+2,+3,+4,+6,+7"},
	{"Fe", 26, 4, 8, "d", 1.83, "+2,+3"},
	{"Co", 27, 4, 9, "d", 1.88, "+2,+3"},
	{"Ni", 28, 4, 10, "d", 1.91, "+2,+3"},
	{"Cu", 29, 4, 11, "d", 1.90, "+1,+2"},
	{"Zn", 30, 4, 12, "d", 1.65, "+2"},
	{"Ga", 31, 4, 13, "p", 1.81, "+3"},
	{"Ge", 32, 4, 14, "p", 2.01, "-4,+2,+4"},
	{"As", 33, 4, 15, "p", 2.18, "-3,+3,+5"},
	{"Se", 34, 4, 16, "p", 2.55, "-2,+4,+6"},
	{"Br", 35, 4, 17, "p", 2.96, "-1,+1,+5"},
	{"Kr", 36, 4, 18, "p", 3.00, "0,+2"},
	// Period 5
	{"Rb", 37, 5, 1, "s", 0.82, "+1"},
	{"Sr", 38, 5, 2, "s", 0.95, "+2"},
	{"Y", 39, 5, 3, "d", 1.22, "+3"},
	{"Zr", 40, 5, 4, "d", 1.33, "+4"},
	{"Nb", 41, 5, 5, "d", 1.60, "+3,+5"},
	{"Mo", 42, 5, 6, "d", 2.16, "+4,+6"},
	{"Tc", 43, 5, 7, "d", 1.90, "+4,+7"},
	{"Ru", 44, 5, 8, "d", 2.20, "+3,+4,+8"},
	{"Rh", 45, 5, 9, "d", 2.28, "+3"},
	{"Pd", 46, 5, 10, "d", 2.20, "+2,+4"},
	{"Ag", 47, 5, 11, "d", 1.93, "+1"},
	{"Cd", 48, 5, 12, "d", 1.69, "+2"},
	{"In", 49, 5, 13, "p", 1.78, "+3"},
	{"Sn", 50, 5, 14, "p", 1.96, "+2,+4"},
	{"Sb", 51, 5, 15, "p", 2.05, "-3,+3,+5"},
	{"Te", 52, 5, 16, "p", 2.10, "-2,+4,+6"},
	{"I", 53, 5, 17, "p", 2.66, "-1,+1,+5,+7"},
	{"Xe", 54, 5, 18, "p", 2.60, "0,+2,+4,+6,+8"},
	// Period 6
	{"Cs", 55, 6, 1, "s", 0.79, "+1"},
	{"Ba", 56, 6, 2, "s", 0.89, "+2"},
	{"La", 57, 6, 3, "f", 1.10, "+3"},
	{"Ce", 58, 6, 0, "f", 1.12, "+3,+4"},
	{"Pr", 59, 6, 0, "f", 1.13, "+3,+4"},
	{"Nd", 60, 6, 0, "f", 1.14, "+3"},
	{"Pm", 61, 6, 0, "f", 1.13, "+3"},
	{"Sm", 62, 6, 0, "f", 1.17, "+2,+3"},
	{"Eu", 63, 6, 0, "f", 1.20, "+2,+3"},
	{"Gd", 64, 6, 0, "f", 1.20, "+3"},
	{"Tb", 65, 6, 0, "f", 1.10, "+3,+4"},
	{"Dy", 66, 6, 0, "f", 1.22, "+3"},
	{"Ho", 67, 6, 0, "f", 1.23, "+3"},
	{"Er", 68, 6, 0, "f", 1.24, "+3"},
	{"Tm", 69, 6, 0, "f", 1.25, "+3"},
	{"Yb", 70, 6, 0, "f", 1.10, "+2,+3"},
	{"Lu", 71, 6, 3, "f", 1.27, "+3"},
	{"Hf", 72, 6, 4, "d", 1.30, "+4"},
	{"Ta", 73, 6, 5, "d", 1.50, "+5"},
	{"W", 74, 6, 6, "d", 2.36, "+4,+6"},
	{"Re", 75, 6, 7, "d", 1.90, "+4,+7"},
	{"Os", 76, 6, 8, "d", 2.20, "+3,+4,+8"},
	{"Ir", 77, 6, 9, "d", 2.20, "+3,+4,+6"},
	{"Pt", 78, 6, 10, "d", 2.28, "+2,+4"},
	{"Au", 79, 6, 11, "d", 2.54, "+1,+3"},
	{"Hg", 80, 6, 12, "d", 2.00, "+1,+2"},
	{"Tl", 81, 6, 13, "p", 1.62, "+1,+3"},
	{"Pb", 82, 6, 14, "p", 2.33, "+2,+4"},
	{"Bi", 83, 6, 15, "p", 2.02, "+3,+5"},
	{"Po", 84, 6, 16, "p", 2.00, "+2,+4,+6"},
	{"At", 85, 6, 17, "p", 2.20, "-1,+1,+3,+5"},
	{"Rn", 86, 6, 18, "p", 2.20, "0,+2"},
	// Period 7
	{"Fr", 87, 7, 1, "s", 0.70, "+1"},
	{"Ra", 88, 7, 2, "s", 0.90, "+2"},
	{"Ac", 89, 7, 3, "f", 1.10, "+3"},
	{"Th", 90, 7, 0, "f", 1.30, "+4"},
	{"Pa", 91, 7, 0, "f", 1.50, "+4,+5"},
	{"U", 92, 7, 0, "f", 1.38, "+3,+4,+5,+6"},
	{"Np", 93, 7, 0, "f", 1.36, "+3,+4,+5,+6"},
	{"Pu", 94, 7, 0, "f", 1.28, "+3,+4,+5,+6"},
	{"Am", 95, 7, 0, "f", 1.30, "+3"},
	{"Cm", 96, 7, 0, "f", 1.30, "+3"},
	{"Bk", 97, 7, 0, "f", 1.30, "+3"},
	{"Cf", 98, 7, 0, "f", 1.30, "+3"},
	{"Es", 99, 7, 0, "f", 1.30, "+3"},
	{"Fm", 100, 7, 0, "f", 1.30, "+3"},
	{"Md", 101, 7, 0, "f", 1.30, "+3"},
	{"No", 102, 7, 0, "f", 1.30, "+2,+3"},
	{"Lr", 103, 7, 3, "d", 1.30, "+3"},
	{"Rf", 104, 7, 4, "d", 0.0, "+4"},
	{"Db", 105, 7, 5, "d", 0.0, "+5"},
	{"Sg", 106, 7, 6, "d", 0.0, "+6"},
	{"Bh", 107, 7, 7, "d", 0.0, "+7"},
	{"Hs", 108, 7, 8, "d", 0.0, "+8"},
	{"Mt", 109, 7, 9, "d", 0.0, "?"},
	{"Ds", 110, 7, 10, "d", 0.0, "?"},
	{"Rg", 111, 7, 11, "d", 0.0, "?"},
	{"Cn", 112, 7, 12, "d", 0.0, "?"},
	{"Nh", 113, 7, 13, "p", 0.0, "?"},
	{"Fl", 114, 7, 14, "p", 0.0, "?"},
	{"Mc", 115, 7, 15, "p", 0.0, "?"},
	{"Lv", 116, 7, 16, "p", 0.0, "?"},
	{"Ts", 117, 7, 17, "p", 0.0, "?"},
	{"Og", 118, 7, 18, "p", 0.0, "?"},
}

// ==================== Lookups ====================

func ElementMetaBySymbol(symbol string) *ElementMeta {
	for i := range elementMeta {
		if elementMeta[i].Symbol == symbol {
			return &elementMeta[i]
		}
	}
	return nil
}

func ElementMetaByZ(z int) *ElementMeta {
	for i := range elementMeta {
		if elementMeta[i].Z == z {
			return &elementMeta[i]
		}
	}
	return nil
}

func ElementsInPeriod(p int) []ElementMeta {
	var out []ElementMeta
	for _, e := range elementMeta {
		if e.Period == p {
			out = append(out, e)
		}
	}
	return out
}

func ElementsInGroup(g int) []ElementMeta {
	var out []ElementMeta
	for _, e := range elementMeta {
		if e.Group == g {
			out = append(out, e)
		}
	}
	return out
}

func ElementsInBlock(block string) []ElementMeta {
	var out []ElementMeta
	for _, e := range elementMeta {
		if e.Block == block {
			out = append(out, e)
		}
	}
	return out
}

// ==================== Category classification ====================

func (e *ElementMeta) Category() string {
	if e == nil {
		return "unknown"
	}
	z := e.Z
	switch {
	case z == 1 || z == 6 || z == 7 || z == 8 || z == 15 || z == 16 || z == 34:
		return "nonmetal"
	case z == 2 || z == 10 || z == 18 || z == 36 || z == 54 || z == 86 || z == 118:
		return "noble"
	case z == 3 || z == 11 || z == 19 || z == 37 || z == 55 || z == 87:
		return "alkali"
	case z == 4 || z == 12 || z == 20 || z == 38 || z == 56 || z == 88:
		return "alkaline"
	case z == 5 || z == 14 || z == 32 || z == 33 || z == 51 || z == 52:
		return "metalloid"
	case z == 9 || z == 17 || z == 35 || z == 53 || z == 85 || z == 117:
		return "halogen"
	case z >= 57 && z <= 71:
		return "lanthanide"
	case z >= 89 && z <= 103:
		return "actinide"
	case (z >= 21 && z <= 30) || (z >= 39 && z <= 48) || (z >= 72 && z <= 80) || (z >= 104 && z <= 112):
		return "transition"
	case z == 13 || (z >= 31 && z <= 31) || (z >= 49 && z <= 50) || (z >= 81 && z <= 84) || (z >= 113 && z <= 116):
		return "post-transition"
	}
	return "unknown"
}

// ==================== Predicate helpers ====================

func (e *ElementMeta) IsAlkali() bool      { return e != nil && e.Category() == "alkali" }
func (e *ElementMeta) IsAlkaline() bool    { return e != nil && e.Category() == "alkaline" }
func (e *ElementMeta) IsTransition() bool  { return e != nil && e.Category() == "transition" }
func (e *ElementMeta) IsLanthanide() bool  { return e != nil && e.Category() == "lanthanide" }
func (e *ElementMeta) IsActinide() bool    { return e != nil && e.Category() == "actinide" }
func (e *ElementMeta) IsHalogen() bool     { return e != nil && e.Category() == "halogen" }
func (e *ElementMeta) IsNoble() bool       { return e != nil && e.Category() == "noble" }
func (e *ElementMeta) IsMetalloid() bool   { return e != nil && e.Category() == "metalloid" }
func (e *ElementMeta) IsNonmetal() bool    { return e != nil && e.Category() == "nonmetal" }
func (e *ElementMeta) IsPostTransition() bool { return e != nil && e.Category() == "post-transition" }
func (e *ElementMeta) IsMetal() bool {
	if e == nil {
		return false
	}
	c := e.Category()
	return c == "alkali" || c == "alkaline" || c == "transition" || c == "lanthanide" || c == "actinide" || c == "post-transition"
}

// ==================== Derived properties ====================

// ShellCount returns the number of electron shells (= period).
func (e *ElementMeta) ShellCount() int {
	if e == nil {
		return 0
	}
	return e.Period
}

// ValenceCount estimates valence electrons using group number.
// For d-block elements this is a rough approximation.
func (e *ElementMeta) ValenceCount() int {
	if e == nil {
		return 0
	}
	g := e.Group
	if g >= 1 && g <= 2 {
		return g
	}
	if g >= 13 && g <= 18 {
		return g - 10
	}
	if g >= 3 && g <= 12 {
		return 2 // simplified for d-block
	}
	return 0
}

// ==================== Display ====================

func (e *ElementMeta) Summary() string {
	if e == nil {
		return "(unknown element)"
	}
	en := "n/a"
	if e.Electronegativity > 0 {
		en = fmt.Sprintf("%.2f", e.Electronegativity)
	}
	return fmt.Sprintf("%s (Z=%d) · period %d · group %d · block %s · EN %s · ox %s · %s",
		e.Symbol, e.Z, e.Period, e.Group, e.Block, en, e.OxidationStates, e.Category())
}

func ElementSummary(symbol string) string {
	return ElementMetaBySymbol(symbol).Summary()
}

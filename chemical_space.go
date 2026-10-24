package main

import "fmt"

// ToxicityLevel buckets a compound's Toxicity into a human-readable label.
func ToxicityLevel(sp ChemSpecies) string {
	t := sp.Toxicity
	switch {
	case t <= 0.0:
		return "safe"
	case t < 0.2:
		return "mild"
	case t < 0.4:
		return "irritant"
	case t < 0.6:
		return "moderately toxic"
	case t < 0.8:
		return "very toxic"
	}
	return "deadly"
}

// NutrientLevel buckets Nutrient into a descriptor.
func NutrientLevel(sp ChemSpecies) string {
	n := sp.Nutrient
	switch {
	case n <= 0.0:
		return "no nutrient value"
	case n < 0.3:
		return "trace nutrient"
	case n < 0.6:
		return "moderate nutrient"
	}
	return "rich nutrient"
}

// RadiationLevel buckets Radiation into a descriptor.
func RadiationLevel(sp ChemSpecies) string {
	r := sp.Radiation
	switch {
	case r <= 0.0:
		return "non-radioactive"
	case r < 0.3:
		return "weakly radioactive"
	case r < 0.6:
		return "moderately radioactive"
	case r < 0.85:
		return "highly radioactive"
	}
	return "extremely radioactive"
}

// EnvironmentClass returns a one-word summary of how this compound would
// affect a growing cell.
func EnvironmentClass(sp ChemSpecies) string {
	if sp.Radiation > 0.5 {
		return "radiotoxic"
	}
	if sp.Toxicity > 0.7 {
		return "lethal"
	}
	if sp.Oxidant > 0.7 {
		return "oxidative-stress"
	}
	if sp.Nutrient > 0.5 {
		return "nutritive"
	}
	if sp.Class == "acid" && sp.pH < 4 {
		return "acidifying"
	}
	if sp.Class == "base" && sp.pH > 10 {
		return "alkalinizing"
	}
	if sp.Class == "detergent" || sp.Class == "organic" {
		return "membrane-active"
	}
	return "benign"
}

// CommonCompoundsByEnvironment returns compounds grouped by what they do.
// Each entry is [category, example-compound-formula].
func CommonCompoundsByEnvironment() [][2]string {
	return [][2]string{
		{"nutrient", "Glc"},
		{"nutrient", "VitB12"},
		{"nutrient", "ATP"},
		{"toxin", "Hg2+"},
		{"toxin", "CN-"},
		{"toxin", "HCN"},
		{"toxin", "Po210"},
		{"oxidizer", "H2O2"},
		{"oxidizer", "MnO4-"},
		{"oxidizer", "O3"},
		{"acid", "HCl"},
		{"acid", "H2SO4"},
		{"base", "NaOH"},
		{"base", "NH4OH"},
		{"metal", "Fe2+"},
		{"metal", "Zn2+"},
		{"metal", "Cu2+"},
		{"detergent", "SDS"},
		{"detergent", "TX100"},
	}
}

// CompoundInteractionProfile returns a plain-text description of what a
// compound does to a cell environment.
func CompoundInteractionProfile(formula string) string {
	sp := FindSpeciesByFormula(formula)
	if sp == nil {
		return fmt.Sprintf("Unknown compound %q", formula)
	}
	return fmt.Sprintf(
		"%s (%s)\n  toxicity:    %s (%.2f)\n  nutrient:    %s (%.2f)\n  oxidant:     %.2f\n  radiation:   %s (%.2f)\n  pH:          %.2f\n  environment: %s",
		sp.Name, sp.Formula,
		ToxicityLevel(*sp), sp.Toxicity,
		NutrientLevel(*sp), sp.Nutrient,
		sp.Oxidant,
		RadiationLevel(*sp), sp.Radiation,
		sp.pH,
		EnvironmentClass(*sp))
}

// ToxicCompounds returns compounds above a toxicity threshold.
func ToxicCompounds(threshold float64) []ChemSpecies {
	var out []ChemSpecies
	for _, s := range ChemSpeciesTable {
		if s.Toxicity >= threshold {
			out = append(out, s)
		}
	}
	return out
}

// RadioactiveCompounds returns every radioactive entry.
func RadioactiveCompounds() []ChemSpecies {
	var out []ChemSpecies
	for _, s := range ChemSpeciesTable {
		if s.Radiation > 0.0 {
			out = append(out, s)
		}
	}
	return out
}

// OxidizerCompounds returns compounds flagged as oxidizers.
func OxidizerCompounds() []ChemSpecies {
	var out []ChemSpecies
	for _, s := range ChemSpeciesTable {
		if s.Oxidant >= 0.5 {
			out = append(out, s)
		}
	}
	return out
}

// CountByEnvironmentClass tallies how many compounds fall in each
// environment class.
func CountByEnvironmentClass() map[string]int {
	counts := map[string]int{}
	for _, s := range ChemSpeciesTable {
		counts[EnvironmentClass(s)]++
	}
	return counts
}

// CompatibilityVerdict returns a short label for how risky it is to have
// both compounds in the same environment.
func CompatibilityVerdict(a, b ChemSpecies) string {
	if a.Class == "radioactive" || b.Class == "radioactive" {
		return "radioactive contamination"
	}
	if a.Class == "noble" || b.Class == "noble" {
		return "inert"
	}
	if (a.Class == "acid" && b.Class == "base") || (a.Class == "base" && b.Class == "acid") {
		return "will neutralize"
	}
	if (a.Class == "oxidizer" && b.Class == "organic") || (b.Class == "oxidizer" && a.Class == "organic") {
		return "vigorous oxidation"
	}
	if a.Class == "chelator" && b.Class == "metal" {
		return "chelates the metal"
	}
	if a.Toxicity+b.Toxicity > 1.2 {
		return "combined lethal toxicity"
	}
	return "compatible"
}

package main

import "fmt"

// ReactionRule matches two ChemSpecies and produces a reaction product.
// Specific formula matchers take priority over class matchers; the first
// rule in reactionRules whose matchers both succeed wins.
type ReactionRule struct {
	Name         string
	LeftFormula  string // "" means match by class instead
	RightFormula string
	LeftClass    string
	RightClass   string
	Symmetric    bool
	Product      func(a, b ChemSpecies) ChemSpecies
}

// ==================== Product helpers ====================

func makeNeutralSalt(a, b ChemSpecies) ChemSpecies {
	return cmpd("Generated", "salt",
		a.Name+" + "+b.Name+" (neutralized)",
		"SALT_"+sanitizeID(a.Formula)+"_"+sanitizeID(b.Formula),
		0.0, 0.2, 0.0, 0.0, 7.0,
		"Acid-base neutralization → salt + water")
}

func makeGasEvolution(a, b ChemSpecies, gasName string) ChemSpecies {
	return cmpd("Generated", "neutral",
		gasName+" + salt",
		"GAS_"+sanitizeID(a.Formula)+"_"+sanitizeID(b.Formula),
		0.1, 0.1, 0.0, 0.0, 6.5,
		"Gas evolution: "+gasName+" released")
}

func makeMetalSalt(a, b ChemSpecies) ChemSpecies {
	return cmpd("Generated", "salt",
		"Metal salt + H2 gas",
		"MSALT_"+sanitizeID(a.Formula)+"_"+sanitizeID(b.Formula),
		0.2, 0.0, 0.0, 0.0, 3.5,
		"Acid-metal dissolution; H2 released")
}

func makeOxidationProduct(a, b ChemSpecies) ChemSpecies {
	return cmpd("Generated", "neutral",
		"Oxidation products",
		"OX_"+sanitizeID(a.Formula)+"_"+sanitizeID(b.Formula),
		0.3, 0.0, 0.2, 0.0, 5.0,
		"Partial oxidation")
}

func makeCombustionProduct(a, b ChemSpecies) ChemSpecies {
	return cmpd("Generated", "acid",
		"Combustion products (CO2 + H2O)",
		"COMB_"+sanitizeID(a.Formula)+"_"+sanitizeID(b.Formula),
		0.05, 0.0, 0.0, 0.0, 5.5,
		"Complete combustion")
}

func makeChelate(a, b ChemSpecies) ChemSpecies {
	return cmpd("Generated", "neutral",
		"Chelate complex",
		"CHEL_"+sanitizeID(a.Formula)+"_"+sanitizeID(b.Formula),
		0.05, 0.0, 0.0, 0.0, 7.0,
		"Metal-chelator complex; reduces free metal toxicity")
}

func makeRadioactiveMixture(a, b ChemSpecies) ChemSpecies {
	rad := a.Radiation + b.Radiation
	return cmpd("Generated", "radioactive",
		"Radioactive mixture",
		"RMIX_"+sanitizeID(a.Formula)+"_"+sanitizeID(b.Formula),
		0.5, 0.0, 0.0, rad, 7.0,
		"Retains radioactivity; contamination propagates")
}

func makeHalideSalt(a, b ChemSpecies) ChemSpecies {
	return cmpd("Generated", "salt",
		"Metal halide",
		"HAL_"+sanitizeID(a.Formula)+"_"+sanitizeID(b.Formula),
		0.3, 0.0, 0.0, 0.0, 5.0,
		"Halogen + metal → metal halide")
}

func makeRedoxProduct(a, b ChemSpecies) ChemSpecies {
	return cmpd("Generated", "neutral",
		"Redox products",
		"RED_"+sanitizeID(a.Formula)+"_"+sanitizeID(b.Formula),
		0.2, 0.0, 0.1, 0.0, 7.0,
		"Redox couple: oxidizer + reductant")
}

func makeSulfideGas(a, b ChemSpecies) ChemSpecies {
	return cmpd("Generated", "acid",
		"H2S + salt",
		"H2S_"+sanitizeID(a.Formula)+"_"+sanitizeID(b.Formula),
		0.4, 0.0, 0.2, 0.0, 4.5,
		"Acid + sulfide → H2S")
}

func makeSulfiteGas(a, b ChemSpecies) ChemSpecies {
	return cmpd("Generated", "acid",
		"SO2 + salt",
		"SO2_"+sanitizeID(a.Formula)+"_"+sanitizeID(b.Formula),
		0.3, 0.0, 0.3, 0.0, 3.5,
		"Acid + sulfite → SO2")
}

func makeAmmoniumSalt(a, b ChemSpecies) ChemSpecies {
	return cmpd("Generated", "salt",
		"Ammonium salt",
		"NH4_"+sanitizeID(a.Formula)+"_"+sanitizeID(b.Formula),
		0.15, 0.15, 0.0, 0.0, 6.0,
		"Acid + ammonia → ammonium salt")
}

func makeWater(a, b ChemSpecies) ChemSpecies {
	return cmpd("Water & ROS", "neutral", "Water", "H2O",
		0.0, 1.0, 0.0, 0.0, 7.0, "Water from combustion")
}

// ==================== Rule table ====================

var reactionRules = []ReactionRule{
	// -------- Specific named pairs --------
	{
		Name: "H2 + O2 → H2O", LeftFormula: "H2", RightFormula: "O2", Symmetric: true,
		Product: makeWater,
	},
	{
		Name: "CH4 + O2 → CO2 + H2O", LeftFormula: "CH4", RightFormula: "O2", Symmetric: true,
		Product: makeCombustionProduct,
	},
	{
		Name: "HCl + NaOH → NaCl + H2O", LeftFormula: "HCl", RightFormula: "NaOH", Symmetric: true,
		Product: makeNeutralSalt,
	},
	{
		Name: "H2SO4 + NaOH → Na2SO4 + H2O", LeftFormula: "H2SO4", RightFormula: "NaOH", Symmetric: true,
		Product: makeNeutralSalt,
	},
	{
		Name: "CaCO3 + HCl → CaCl2 + CO2 + H2O", LeftFormula: "CaCO3", RightFormula: "HCl", Symmetric: true,
		Product: func(a, b ChemSpecies) ChemSpecies {
			return makeGasEvolution(a, b, "CO2")
		},
	},
	{
		Name: "Na2S + HCl → NaCl + H2S", LeftFormula: "Na2S", RightFormula: "HCl", Symmetric: true,
		Product: makeSulfideGas,
	},
	{
		Name: "EDTA + Pb2+ → chelate", LeftFormula: "EDTA", RightFormula: "Pb2+", Symmetric: true,
		Product: makeChelate,
	},
	{
		Name: "EDTA + Hg2+ → chelate", LeftFormula: "EDTA", RightFormula: "Hg2+", Symmetric: true,
		Product: makeChelate,
	},
	{
		Name: "EDTA + Cd2+ → chelate", LeftFormula: "EDTA", RightFormula: "Cd2+", Symmetric: true,
		Product: makeChelate,
	},
	{
		Name: "DTT + H2O2 → oxidized DTT", LeftFormula: "DTT", RightFormula: "H2O2", Symmetric: true,
		Product: makeRedoxProduct,
	},
	{
		Name: "BME + H2O2 → oxidized BME", LeftFormula: "BME", RightFormula: "H2O2", Symmetric: true,
		Product: makeRedoxProduct,
	},
	{
		Name: "NaOH + Al → aluminate + H2", LeftFormula: "NaOH", RightFormula: "Al", Symmetric: true,
		Product: func(a, b ChemSpecies) ChemSpecies {
			return cmpd("Generated", "salt", "Aluminate + H2",
				"ALUM_"+sanitizeID(a.Formula)+"_"+sanitizeID(b.Formula),
				0.3, 0.0, 0.0, 0.0, 12.0, "Base dissolves aluminum")
		},
	},
	{
		Name: "Cl2 + Na → NaCl", LeftFormula: "Cl2", RightFormula: "Na", Symmetric: true,
		Product: makeHalideSalt,
	},
	{
		Name: "Ag + Cl → AgCl", LeftFormula: "Ag", RightFormula: "Cl", Symmetric: true,
		Product: makeHalideSalt,
	},
	{
		Name: "H2O2 + Fe2+ → Fenton reaction", LeftFormula: "H2O2", RightFormula: "Fe2+", Symmetric: true,
		Product: func(a, b ChemSpecies) ChemSpecies {
			return cmpd("Water & ROS", "oxidizer", "Hydroxyl radical",
				"OH*_FENTON", 0.9, 0.0, 1.0, 0.0, 7.0,
				"Fenton chemistry: H2O2 + Fe2+ → OH* + OH- + Fe3+")
		},
	},
	{
		Name: "HCl + NaHCO3 → NaCl + CO2 + H2O", LeftFormula: "HCl", RightFormula: "NaHCO3", Symmetric: true,
		Product: func(a, b ChemSpecies) ChemSpecies {
			return makeGasEvolution(a, b, "CO2")
		},
	},
	{
		Name: "HCl + CaCO3 → CaCl2 + CO2 + H2O", LeftFormula: "HCl", RightFormula: "CaCO3", Symmetric: true,
		Product: func(a, b ChemSpecies) ChemSpecies {
			return makeGasEvolution(a, b, "CO2")
		},
	},
	{
		Name: "CaCO3 + H2SO4 → CaSO4 + CO2", LeftFormula: "CaCO3", RightFormula: "H2SO4", Symmetric: true,
		Product: func(a, b ChemSpecies) ChemSpecies {
			return makeGasEvolution(a, b, "CO2")
		},
	},

	// -------- Class-based general rules --------
	{
		Name: "acid + base → salt", LeftClass: "acid", RightClass: "base", Symmetric: true,
		Product: makeNeutralSalt,
	},
	{
		Name: "acid + carbonate → salt + CO2", LeftClass: "acid", RightClass: "carbonate", Symmetric: true,
		Product: func(a, b ChemSpecies) ChemSpecies {
			return makeGasEvolution(a, b, "CO2")
		},
	},
	{
		Name: "acid + metal → salt + H2", LeftClass: "acid", RightClass: "metal", Symmetric: true,
		Product: makeMetalSalt,
	},
	{
		Name: "acid + sulfide → H2S", LeftClass: "acid", RightClass: "sulfide", Symmetric: true,
		Product: makeSulfideGas,
	},
	{
		Name: "acid + oxidizer → oxidation products", LeftClass: "acid", RightClass: "oxidizer", Symmetric: true,
		Product: makeOxidationProduct,
	},
	{
		Name: "base + organic → saponification", LeftClass: "base", RightClass: "organic", Symmetric: true,
		Product: func(a, b ChemSpecies) ChemSpecies {
			return cmpd("Generated", "detergent", "Soap + alcohol",
				"SAP_"+sanitizeID(a.Formula)+"_"+sanitizeID(b.Formula),
				0.3, 0.0, 0.0, 0.0, 9.0,
				"Base-catalyzed ester hydrolysis")
		},
	},
	{
		Name: "base + halogen → hypohalite", LeftClass: "base", RightClass: "halogen", Symmetric: true,
		Product: func(a, b ChemSpecies) ChemSpecies {
			return cmpd("Chlorine & Halogens", "oxidizer", "Hypohalite salt",
				"HYPO_"+sanitizeID(a.Formula)+"_"+sanitizeID(b.Formula),
				0.4, 0.0, 0.7, 0.0, 10.0, "Base + halogen bleach")
		},
	},
	{
		Name: "oxidizer + organic → partial oxidation", LeftClass: "oxidizer", RightClass: "organic", Symmetric: true,
		Product: makeOxidationProduct,
	},
	{
		Name: "oxidizer + nutrient → burnt nutrients", LeftClass: "oxidizer", RightClass: "nutrient", Symmetric: true,
		Product: makeOxidationProduct,
	},
	{
		Name: "oxidizer + metal → metal oxide", LeftClass: "oxidizer", RightClass: "metal", Symmetric: true,
		Product: func(a, b ChemSpecies) ChemSpecies {
			return cmpd("Generated", "salt", "Metal oxide",
				"MOX_"+sanitizeID(a.Formula)+"_"+sanitizeID(b.Formula),
				0.3, 0.0, 0.0, 0.0, 6.0, "Oxidizer attacks metal")
		},
	},
	{
		Name: "oxidizer + chelator → oxidized chelator", LeftClass: "oxidizer", RightClass: "chelator", Symmetric: true,
		Product: makeRedoxProduct,
	},
	{
		Name: "chelator + metal → chelate", LeftClass: "chelator", RightClass: "metal", Symmetric: true,
		Product: makeChelate,
	},
	{
		Name: "halogen + metal → metal halide", LeftClass: "halogen", RightClass: "metal", Symmetric: true,
		Product: makeHalideSalt,
	},
	{
		Name: "halogen + nonmetal → interhalogen", LeftClass: "halogen", RightClass: "nonmetal", Symmetric: true,
		Product: func(a, b ChemSpecies) ChemSpecies {
			return cmpd("Chlorine & Halogens", "oxidizer", "Interhalogen compound",
				"INTER_"+sanitizeID(a.Formula)+"_"+sanitizeID(b.Formula),
				0.5, 0.0, 0.5, 0.0, 5.0, "Interhalogen reaction")
		},
	},
	{
		Name: "acid + acid → mixed acid", LeftClass: "acid", RightClass: "acid", Symmetric: true,
		Product: func(a, b ChemSpecies) ChemSpecies {
			return cmpd("Generated", "acid", "Mixed acid",
				"MIXACID_"+sanitizeID(a.Formula)+"_"+sanitizeID(b.Formula),
				(a.Toxicity+b.Toxicity)/2, 0.0, (a.Oxidant+b.Oxidant)/2, 0.0,
				(a.pH+b.pH)/2, "Mixed acid; superacidic environment")
		},
	},
	{
		Name: "base + base → mixed base", LeftClass: "base", RightClass: "base", Symmetric: true,
		Product: func(a, b ChemSpecies) ChemSpecies {
			return cmpd("Generated", "base", "Mixed base",
				"MIXBASE_"+sanitizeID(a.Formula)+"_"+sanitizeID(b.Formula),
				(a.Toxicity+b.Toxicity)/2, 0.0, 0.0, 0.0,
				(a.pH+b.pH)/2, "Mixed base; hyperbasic environment")
		},
	},
	{
		Name: "noble + anything → no reaction", LeftClass: "noble",
		Product: func(a, b ChemSpecies) ChemSpecies {
			return cmpd("Generated", "neutral", "Inert mixture",
				"INERT_"+sanitizeID(a.Formula)+"_"+sanitizeID(b.Formula),
				0.0, 0.0, 0.0, 0.0, 7.0, "Noble gas does not react")
		},
	},

	// -------- Radioactivity (last before generic) --------
	{
		Name: "radioactive + anything → radioactive mixture", LeftClass: "radioactive",
		Product: makeRadioactiveMixture,
	},
}

// ==================== Matching ====================

func matchesRuleSide(formula, class string, sp ChemSpecies) bool {
	if formula != "" {
		return sp.Formula == formula
	}
	if class != "" {
		return sp.Class == class
	}
	return true
}

// matchReactionRule returns the first rule that matches (a, b) in order,
// trying the reverse orientation if the rule is symmetric.
func matchReactionRule(a, b ChemSpecies) *ReactionRule {
	for i := range reactionRules {
		r := &reactionRules[i]
		if matchesRuleSide(r.LeftFormula, r.LeftClass, a) &&
			matchesRuleSide(r.RightFormula, r.RightClass, b) {
			return r
		}
		if r.Symmetric &&
			matchesRuleSide(r.LeftFormula, r.LeftClass, b) &&
			matchesRuleSide(r.RightFormula, r.RightClass, a) {
			return r
		}
	}
	// Final radioactive fallback in either direction
	for i := range reactionRules {
		r := &reactionRules[i]
		if r.Name == "radioactive + anything → radioactive mixture" {
			if a.Class == "radioactive" || b.Class == "radioactive" {
				return r
			}
		}
	}
	return nil
}

// ==================== Combustion classifier ====================

// isCombustible reports whether a species is likely to combust in oxygen.
func isCombustible(sp ChemSpecies) bool {
	if sp.Class == "organic" || sp.Class == "nutrient" || sp.Class == "detergent" {
		return true
	}
	if sp.Formula == "CH4" || sp.Formula == "H2" || sp.Formula == "CO" {
		return true
	}
	return false
}

// ==================== Public entry point ====================

// reactSpeciesBetter is the rule-table-driven replacement for reactSpecies.
// Returns a synthetic ChemSpecies describing the products of (a, b).
func reactSpeciesBetter(a, b ChemSpecies) ChemSpecies {
	if r := matchReactionRule(a, b); r != nil {
		product := r.Product(a, b)
		if product.Description == "" {
			product.Description = r.Name
		}
		return product
	}

	// Combustion check: combustible + oxygen
	if (isCombustible(a) && b.Formula == "O2") || (isCombustible(b) && a.Formula == "O2") {
		return makeCombustionProduct(a, b)
	}

	// Generic mixture
	return cmpd("Generated", "neutral",
		a.Name+" / "+b.Name+" mixture",
		"MIX_"+sanitizeID(a.Formula)+"_"+sanitizeID(b.Formula),
		(a.Toxicity+b.Toxicity)/2,
		(a.Nutrient+b.Nutrient)/2,
		(a.Oxidant+b.Oxidant)/2,
		(a.Radiation+b.Radiation)/2,
		weightedPH(a.pH, b.pH),
		"Generic mixture")
}

// ==================== Solubility ====================

// SolubilityClass returns a rough solubility tag for a compound formula.
func SolubilityClass(formula string) string {
	switch formula {
	case "NaCl", "KCl", "NaNO3", "KNO3", "Na2SO4":
		return "soluble"
	case "CaCO3", "PbCl2", "AgCl", "BaSO4":
		return "insoluble"
	case "NaOH", "KOH", "HCl", "H2SO4", "HNO3":
		return "soluble (ionizes)"
	}
	return "unknown"
}

// ==================== Reaction energy estimation ====================

// ReactionEnergyClass returns "exothermic", "endothermic", or "neutral"
// based on the reaction type.
func ReactionEnergyClass(a, b ChemSpecies) string {
	if a.Class == "acid" && b.Class == "base" {
		return "exothermic"
	}
	if a.Class == "base" && b.Class == "acid" {
		return "exothermic"
	}
	if a.Class == "oxidizer" || b.Class == "oxidizer" {
		return "exothermic"
	}
	if a.Class == "radioactive" || b.Class == "radioactive" {
		return "energy-releasing"
	}
	if a.Formula == "H2" && b.Formula == "O2" {
		return "highly exothermic"
	}
	if a.Formula == "CH4" && b.Formula == "O2" {
		return "highly exothermic"
	}
	return "neutral"
}

// SummarizeReaction produces a one-line summary of what would happen.
func SummarizeReaction(a, b ChemSpecies) string {
	r := matchReactionRule(a, b)
	if r != nil {
		return r.Name
	}
	if (isCombustible(a) && b.Formula == "O2") || (isCombustible(b) && a.Formula == "O2") {
		return "Combustion"
	}
	return fmt.Sprintf("Mixture of %s and %s (no reaction)", a.Name, b.Name)
}

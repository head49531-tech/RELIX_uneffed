package main

// extraCompoundSpecies extends the compound table with biologically and
// chemically important molecules that are not covered by the base table.
var extraCompoundSpecies = []ChemSpecies{
	// -------- Vitamins (13) --------
	cmpd("Vitamins", "nutrient", "Retinol (Vitamin A)", "VitA", 0.05, 0.6, 0.05, 0.0, 7.0, "Fat-soluble; vision and immunity"),
	cmpd("Vitamins", "nutrient", "Thiamine (B1)", "VitB1", 0.0, 0.7, 0.0, 0.0, 7.0, "Coenzyme in carbohydrate metabolism"),
	cmpd("Vitamins", "nutrient", "Riboflavin (B2)", "VitB2", 0.0, 0.7, 0.0, 0.0, 7.0, "FAD/FMN precursor"),
	cmpd("Vitamins", "nutrient", "Niacin (B3)", "VitB3", 0.0, 0.6, 0.0, 0.0, 7.0, "NAD+/NADP+ precursor"),
	cmpd("Vitamins", "nutrient", "Pantothenic acid (B5)", "VitB5", 0.0, 0.6, 0.0, 0.0, 7.0, "Coenzyme A precursor"),
	cmpd("Vitamins", "nutrient", "Pyridoxine (B6)", "VitB6", 0.0, 0.6, 0.0, 0.0, 7.0, "Amino acid metabolism"),
	cmpd("Vitamins", "nutrient", "Biotin (B7)", "VitB7", 0.0, 0.6, 0.0, 0.0, 7.0, "Carboxylase cofactor"),
	cmpd("Vitamins", "nutrient", "Folate (B9)", "VitB9", 0.0, 0.6, 0.0, 0.0, 7.0, "One-carbon metabolism"),
	cmpd("Vitamins", "nutrient", "Cobalamin (B12)", "VitB12", 0.0, 0.7, 0.0, 0.0, 7.0, "Cofactor; requires cobalt"),
	cmpd("Vitamins", "nutrient", "Ascorbic acid (C)", "VitC", 0.1, 0.8, 0.1, 0.0, 3.0, "Antioxidant; collagen synthesis"),
	cmpd("Vitamins", "nutrient", "Calciferol (D)", "VitD", 0.1, 0.5, 0.0, 0.0, 7.0, "Calcium homeostasis"),
	cmpd("Vitamins", "nutrient", "Tocopherol (E)", "VitE", 0.05, 0.6, 0.05, 0.0, 7.0, "Lipid antioxidant"),
	cmpd("Vitamins", "nutrient", "Phylloquinone (K)", "VitK", 0.05, 0.6, 0.0, 0.0, 7.0, "Clotting factor cofactor"),

	// -------- Hormones (12) --------
	cmpd("Hormones", "signaling", "Insulin", "Ins", 0.0, 0.3, 0.0, 0.0, 7.0, "Glucose uptake hormone"),
	cmpd("Hormones", "signaling", "Glucagon", "Gluc", 0.0, 0.3, 0.0, 0.0, 7.0, "Glucose release hormone"),
	cmpd("Hormones", "signaling", "Adrenaline", "Adr", 0.1, 0.0, 0.0, 0.0, 7.0, "Fight-or-flight catecholamine"),
	cmpd("Hormones", "signaling", "Noradrenaline", "NAdr", 0.1, 0.0, 0.0, 0.0, 7.0, "Catecholamine neurotransmitter"),
	cmpd("Hormones", "signaling", "Cortisol", "Cort", 0.15, 0.0, 0.0, 0.0, 7.0, "Glucocorticoid stress hormone"),
	cmpd("Hormones", "signaling", "Aldosterone", "Aldo", 0.05, 0.0, 0.0, 0.0, 7.0, "Mineralocorticoid; Na+ retention"),
	cmpd("Hormones", "signaling", "Testosterone", "Test", 0.1, 0.0, 0.0, 0.0, 7.0, "Androgen"),
	cmpd("Hormones", "signaling", "Estradiol", "E2", 0.1, 0.0, 0.0, 0.0, 7.0, "Estrogen"),
	cmpd("Hormones", "signaling", "Progesterone", "Prog", 0.1, 0.0, 0.0, 0.0, 7.0, "Progestogen"),
	cmpd("Hormones", "signaling", "Thyroxine (T4)", "T4", 0.1, 0.0, 0.0, 0.0, 7.0, "Thyroid hormone"),
	cmpd("Hormones", "signaling", "Melatonin", "Mel", 0.0, 0.0, 0.0, 0.0, 7.0, "Circadian rhythm hormone"),
	cmpd("Hormones", "signaling", "Vasopressin", "ADH", 0.05, 0.0, 0.0, 0.0, 7.0, "Antidiuretic hormone"),

	// -------- Neurotransmitters (8) --------
	cmpd("Neurotransmitters", "signaling", "Dopamine", "DA", 0.1, 0.0, 0.0, 0.0, 7.0, "Reward/motivation"),
	cmpd("Neurotransmitters", "signaling", "Serotonin", "5HT", 0.05, 0.0, 0.0, 0.0, 7.0, "Mood regulation"),
	cmpd("Neurotransmitters", "signaling", "GABA", "GABA", 0.0, 0.0, 0.0, 0.0, 7.0, "Inhibitory neurotransmitter"),
	cmpd("Neurotransmitters", "signaling", "Glutamate", "Glu", 0.1, 0.3, 0.0, 0.0, 7.0, "Excitatory neurotransmitter"),
	cmpd("Neurotransmitters", "signaling", "Acetylcholine", "ACh", 0.1, 0.0, 0.0, 0.0, 7.0, "Neuromuscular junction"),
	cmpd("Neurotransmitters", "signaling", "Glycine", "GlyN", 0.0, 0.3, 0.0, 0.0, 7.0, "Inhibitory amino acid"),
	cmpd("Neurotransmitters", "signaling", "Histamine", "Hist", 0.2, 0.0, 0.0, 0.0, 7.0, "Inflammation, wakefulness"),
	cmpd("Neurotransmitters", "signaling", "Adenosine", "Ado", 0.0, 0.2, 0.0, 0.0, 7.0, "Sleep pressure"),

	// -------- Common drugs (12) --------
	cmpd("Drugs", "neutral", "Morphine", "Morph", 0.7, 0.0, 0.0, 0.0, 7.0, "Opioid analgesic"),
	cmpd("Drugs", "neutral", "Aspirin", "ASA", 0.3, 0.0, 0.0, 0.0, 3.5, "NSAID; COX inhibitor"),
	cmpd("Drugs", "neutral", "Ibuprofen", "Ibu", 0.3, 0.0, 0.0, 0.0, 6.0, "NSAID"),
	cmpd("Drugs", "neutral", "Paracetamol", "Para", 0.4, 0.0, 0.0, 0.0, 6.0, "Analgesic; hepatotoxic at high dose"),
	cmpd("Drugs", "neutral", "Diazepam", "Dz", 0.5, 0.0, 0.0, 0.0, 7.0, "Benzodiazepine; GABA agonist"),
	cmpd("Drugs", "neutral", "Fluoxetine", "Flx", 0.5, 0.0, 0.0, 0.0, 7.0, "SSRI antidepressant"),
	cmpd("Drugs", "neutral", "Metformin", "Met", 0.3, 0.0, 0.0, 0.0, 7.0, "Antidiabetic"),
	cmpd("Drugs", "neutral", "Warfarin", "Warf", 0.6, 0.0, 0.0, 0.0, 7.0, "Vitamin K antagonist"),
	cmpd("Drugs", "neutral", "Atropine", "Atr", 0.7, 0.0, 0.0, 0.0, 7.0, "Muscarinic antagonist"),
	cmpd("Drugs", "neutral", "Nicotine", "Nic", 0.8, 0.0, 0.0, 0.0, 7.0, "Nicotinic agonist"),
	cmpd("Drugs", "neutral", "Caffeine", "Caff", 0.4, 0.0, 0.0, 0.0, 7.0, "Adenosine antagonist"),
	cmpd("Drugs", "neutral", "Cocaine", "Coc", 0.9, 0.0, 0.0, 0.0, 7.0, "Dopamine reuptake inhibitor"),

	// -------- Organic acids (10) --------
	cmpd("Organic acids", "acid", "Pyruvate", "Pyr", 0.1, 0.4, 0.0, 0.0, 2.5, "Glycolysis product"),
	cmpd("Organic acids", "acid", "Succinate", "Succ", 0.05, 0.3, 0.0, 0.0, 4.2, "TCA cycle intermediate"),
	cmpd("Organic acids", "acid", "Fumarate", "Fum", 0.05, 0.3, 0.0, 0.0, 3.0, "TCA cycle intermediate"),
	cmpd("Organic acids", "acid", "Malate", "Mal", 0.05, 0.3, 0.0, 0.0, 3.4, "TCA cycle intermediate"),
	cmpd("Organic acids", "acid", "Oxaloacetate", "OAA", 0.05, 0.3, 0.0, 0.0, 2.8, "TCA cycle intermediate"),
	cmpd("Organic acids", "acid", "Alpha-ketoglutarate", "aKG", 0.05, 0.3, 0.0, 0.0, 2.5, "TCA cycle intermediate"),
	cmpd("Organic acids", "acid", "Isocitrate", "Isocit", 0.05, 0.3, 0.0, 0.0, 3.2, "TCA cycle intermediate"),
	cmpd("Organic acids", "acid", "Aconitate", "Acon", 0.05, 0.3, 0.0, 0.0, 3.0, "TCA cycle intermediate"),
	cmpd("Organic acids", "acid", "Acetyl-CoA", "AcCoA", 0.05, 0.4, 0.0, 0.0, 4.0, "Central metabolite"),
	cmpd("Organic acids", "acid", "Malonate", "Malon", 0.2, 0.1, 0.0, 0.0, 2.8, "Succinate dehydrogenase inhibitor"),

	// -------- Signaling molecules (8) --------
	cmpd("Signaling", "signaling", "cAMP", "cAMP", 0.0, 0.1, 0.0, 0.0, 7.0, "Second messenger"),
	cmpd("Signaling", "signaling", "cGMP", "cGMP", 0.0, 0.1, 0.0, 0.0, 7.0, "Second messenger"),
	cmpd("Signaling", "signaling", "IP3", "IP3", 0.0, 0.1, 0.0, 0.0, 7.0, "Inositol trisphosphate"),
	cmpd("Signaling", "signaling", "DAG", "DAG", 0.0, 0.1, 0.0, 0.0, 7.0, "Diacylglycerol"),
	cmpd("Signaling", "nutrient", "ATP", "ATP", 0.0, 0.9, 0.0, 0.0, 7.0, "Energy currency"),
	cmpd("Signaling", "nutrient", "ADP", "ADP", 0.0, 0.7, 0.0, 0.0, 7.0, "Partially hydrolyzed ATP"),
	cmpd("Signaling", "nutrient", "AMP", "AMP", 0.0, 0.5, 0.0, 0.0, 7.0, "Adenosine monophosphate"),
	cmpd("Signaling", "nutrient", "NAD+", "NAD", 0.0, 0.6, 0.0, 0.0, 7.0, "Redox cofactor"),

	// -------- Ions (8) --------
	cmpd("Ions", "base", "Bicarbonate ion", "HCO3-", 0.0, 0.2, 0.0, 0.0, 8.3, "Buffer"),
	cmpd("Ions", "base", "Carbonate ion", "CO3-2", 0.1, 0.0, 0.0, 0.0, 11.0, "Strong base"),
	cmpd("Ions", "base", "Hydroxide ion", "OH-", 0.4, 0.0, 0.0, 0.0, 13.5, "Strong base"),
	cmpd("Ions", "acid", "Hydronium ion", "H3O+", 0.4, 0.0, 0.0, 0.0, 1.0, "Strong acid"),
	cmpd("Ions", "nutrient", "Phosphate ion", "PO4-3", 0.0, 0.3, 0.0, 0.0, 8.0, "Bio-phosphate"),
	cmpd("Ions", "nutrient", "Sulfate ion", "SO4-2", 0.05, 0.2, 0.0, 0.0, 7.0, "Sulfur source"),
	cmpd("Ions", "nutrient", "Ammonium ion", "NH4+", 0.1, 0.4, 0.0, 0.0, 6.0, "Nitrogen source"),
	cmpd("Ions", "neutral", "Cyanide ion", "CN-", 0.95, 0.0, 0.0, 0.0, 9.0, "Cytochrome oxidase inhibitor"),

	// -------- Fluorophores / dyes (8) --------
	cmpd("Dyes", "neutral", "Fluorescein", "FITC", 0.05, 0.0, 0.0, 0.0, 7.0, "Green fluorophore"),
	cmpd("Dyes", "neutral", "Rhodamine", "Rhod", 0.1, 0.0, 0.0, 0.0, 7.0, "Red fluorophore"),
	cmpd("Dyes", "neutral", "DAPI", "DAPI", 0.3, 0.0, 0.0, 0.0, 7.0, "DNA-binding dye"),
	cmpd("Dyes", "neutral", "Propidium iodide", "PI", 0.5, 0.0, 0.0, 0.0, 7.0, "Nucleic-acid dye"),
	cmpd("Dyes", "neutral", "Ethidium bromide", "EtBr", 0.6, 0.0, 0.0, 0.0, 7.0, "Intercalating mutagen"),
	cmpd("Dyes", "neutral", "Methylene blue", "MeBlue", 0.2, 0.0, 0.1, 0.0, 7.0, "Redox-active dye"),
	cmpd("Dyes", "neutral", "Congo red", "Congo", 0.2, 0.0, 0.0, 0.0, 7.0, "Amyloid stain"),
	cmpd("Dyes", "neutral", "Bromophenol blue", "BPB", 0.1, 0.0, 0.0, 0.0, 7.0, "pH indicator"),

	// -------- Enzymes / proteins (8) --------
	cmpd("Proteins", "nutrient", "Catalase", "Cat", 0.0, 0.6, 0.0, 0.0, 7.0, "H2O2 decomposer"),
	cmpd("Proteins", "nutrient", "Lysozyme", "Lys", 0.1, 0.4, 0.0, 0.0, 7.0, "Cell-wall hydrolase"),
	cmpd("Proteins", "nutrient", "Trypsin", "Try", 0.2, 0.4, 0.0, 0.0, 7.0, "Serine protease"),
	cmpd("Proteins", "nutrient", "Pepsin", "Pep", 0.2, 0.4, 0.0, 0.0, 2.0, "Acid protease"),
	cmpd("Proteins", "nutrient", "Amylase", "Amy", 0.0, 0.4, 0.0, 0.0, 7.0, "Starch hydrolase"),
	cmpd("Proteins", "nutrient", "Lipase", "Lip", 0.0, 0.4, 0.0, 0.0, 7.0, "Lipid hydrolase"),
	cmpd("Proteins", "nutrient", "Bovine serum albumin", "BSA", 0.0, 0.5, 0.0, 0.0, 7.0, "Blocking protein"),
	cmpd("Proteins", "signaling", "IgG antibody", "IgG", 0.0, 0.2, 0.0, 0.0, 7.0, "Immunoglobulin"),

	// -------- Reactive species (8) --------
	cmpd("Reactive species", "oxidizer", "Nitrotyrosine", "NT", 0.5, 0.0, 0.3, 0.0, 7.0, "Oxidative damage marker"),
	cmpd("Reactive species", "oxidizer", "8-oxoguanine", "8oxoG", 0.6, 0.0, 0.4, 0.0, 7.0, "DNA oxidation marker"),
	cmpd("Reactive species", "oxidizer", "Lipid peroxide", "LOOH", 0.4, 0.0, 0.7, 0.0, 7.0, "Membrane oxidation product"),
	cmpd("Reactive species", "oxidizer", "Malondialdehyde", "MDA", 0.5, 0.0, 0.3, 0.0, 7.0, "Lipid peroxidation product"),
	cmpd("Reactive species", "oxidizer", "4-HNE", "4HNE", 0.6, 0.0, 0.3, 0.0, 7.0, "Lipid-derived electrophile"),
	cmpd("Reactive species", "oxidizer", "Protein carbonyl", "PC", 0.5, 0.0, 0.2, 0.0, 7.0, "Protein oxidation marker"),
	cmpd("Reactive species", "neutral", "Isoprostane", "IsoP", 0.3, 0.0, 0.0, 0.0, 7.0, "Lipid peroxidation marker"),
	cmpd("Reactive species", "oxidizer", "Peroxynitrite", "ONOO-_x", 0.8, 0.0, 1.0, 0.0, 7.0, "RNS/ROS hybrid"),

	// -------- Minerals / inorganic ions (8) --------
	cmpd("Minerals", "nutrient", "Selenate", "SeO4", 0.3, 0.2, 0.0, 0.0, 7.0, "Selenium source"),
	cmpd("Minerals", "nutrient", "Molybdate", "MoO4", 0.2, 0.2, 0.0, 0.0, 7.0, "Mo cofactor source"),
	cmpd("Minerals", "nutrient", "Tungstate", "WO4", 0.3, 0.1, 0.0, 0.0, 7.0, "Tungsten source"),
	cmpd("Minerals", "nutrient", "Vanadate", "VO4", 0.4, 0.1, 0.0, 0.0, 7.0, "Vanadium source"),
	cmpd("Minerals", "oxidizer", "Permanganate", "MnO4-", 0.5, 0.0, 0.9, 0.0, 5.0, "Strong oxidizer"),
	cmpd("Minerals", "metal", "Ferricyanide", "FeCN6", 0.3, 0.0, 0.1, 0.0, 7.0, "Redox indicator"),
	cmpd("Minerals", "metal", "Ferrocyanide", "FeCN6r", 0.2, 0.0, 0.0, 0.0, 7.0, "Reduced ferricyanide"),
	cmpd("Minerals", "metal", "Silver nitrate", "AgNO3", 0.6, 0.0, 0.2, 0.0, 6.0, "Precipitating reagent"),

	// -------- More gases (8) --------
	cmpd("Gases", "neutral", "Nitrogen gas", "N2", 0.0, 0.1, 0.0, 0.0, 7.0, "Inert atmospheric gas"),
	cmpd("Gases", "neutral", "Nitrous oxide", "N2O", 0.3, 0.0, 0.1, 0.0, 7.0, "Greenhouse gas"),
	cmpd("Gases", "neutral", "Sulfur hexafluoride", "SF6", 0.4, 0.0, 0.0, 0.0, 7.0, "Potent greenhouse gas"),
	cmpd("Gases", "neutral", "Krypton", "Kr", 0.0, 0.0, 0.0, 0.0, 7.0, "Inert"),
	cmpd("Gases", "neutral", "Xenon", "Xe", 0.0, 0.0, 0.0, 0.0, 7.0, "Anesthetic noble gas"),
	cmpd("Gases", "neutral", "Neon", "Ne", 0.0, 0.0, 0.0, 0.0, 7.0, "Inert"),
	cmpd("Gases", "neutral", "Carbonyl sulfide", "COS", 0.4, 0.0, 0.0, 0.0, 7.0, "Trace gas"),
	cmpd("Gases", "acid", "Carbonyl fluoride", "COF2", 0.6, 0.0, 0.2, 0.0, 3.0, "Toxic acid fluoride"),

	// -------- More solvents (8) --------
	cmpd("Solvents & Detergents", "organic", "Diethyl ether", "Et2O", 0.4, 0.0, 0.0, 0.0, 7.0, "Volatile solvent"),
	cmpd("Solvents & Detergents", "organic", "THF", "THF", 0.5, 0.0, 0.0, 0.0, 7.0, "Cyclic ether"),
	cmpd("Solvents & Detergents", "organic", "Dioxane", "Diox", 0.6, 0.0, 0.0, 0.0, 7.0, "Carcinogenic ether"),
	cmpd("Solvents & Detergents", "organic", "Hexane", "Hex", 0.5, 0.0, 0.0, 0.0, 7.0, "Non-polar solvent"),
	cmpd("Solvents & Detergents", "organic", "Cyclohexane", "Cyhex", 0.4, 0.0, 0.0, 0.0, 7.0, "Non-polar solvent"),
	cmpd("Solvents & Detergents", "organic", "Dichloromethane", "DCM", 0.5, 0.0, 0.0, 0.0, 7.0, "Halogenated solvent"),
	cmpd("Solvents & Detergents", "organic", "Ethyl acetate", "EtOAc", 0.3, 0.0, 0.0, 0.0, 7.0, "Ester solvent"),
	cmpd("Solvents & Detergents", "organic", "Pyridine", "PyrH", 0.6, 0.0, 0.0, 0.0, 8.0, "Aromatic amine"),

	// -------- Common solids / oxides (8) --------
	cmpd("Solids & Oxides", "base", "Quicklime (CaO)", "CaO", 0.3, 0.0, 0.0, 0.0, 11.0, "Basic oxide"),
	cmpd("Solids & Oxides", "neutral", "Alumina (Al2O3)", "Al2O3", 0.1, 0.0, 0.0, 0.0, 7.0, "Inert oxide"),
	cmpd("Solids & Oxides", "neutral", "Silica (SiO2)", "SiO2", 0.1, 0.0, 0.0, 0.0, 7.0, "Silicate dust"),
	cmpd("Solids & Oxides", "neutral", "Titania (TiO2)", "TiO2", 0.0, 0.0, 0.0, 0.0, 7.0, "Photocatalytic oxide"),
	cmpd("Solids & Oxides", "neutral", "Magnesia (MgO)", "MgO", 0.1, 0.0, 0.0, 0.0, 7.0, "Refractory oxide"),
	cmpd("Solids & Oxides", "metal", "Hematite (Fe2O3)", "Fe2O3", 0.2, 0.0, 0.0, 0.0, 7.0, "Iron ore"),
	cmpd("Solids & Oxides", "metal", "Magnetite (Fe3O4)", "Fe3O4", 0.2, 0.0, 0.0, 0.0, 7.0, "Magnetic iron ore"),
	cmpd("Solids & Oxides", "neutral", "Rust (FeOOH)", "FeOOH", 0.1, 0.0, 0.0, 0.0, 7.0, "Corrosion product"),
}

// ==================== Filters ====================

func Vitamins() []ChemSpecies         { return extraCompoundSpeciesIn("Vitamins") }
func Hormones() []ChemSpecies         { return extraCompoundSpeciesIn("Hormones") }
func Neurotransmitters() []ChemSpecies { return extraCompoundSpeciesIn("Neurotransmitters") }
func CommonDrugs() []ChemSpecies      { return extraCompoundSpeciesIn("Drugs") }
func OrganicAcidCompounds() []ChemSpecies { return extraCompoundSpeciesIn("Organic acids") }
func SignalingMolecules() []ChemSpecies { return extraCompoundSpeciesIn("Signaling") }
func Ions() []ChemSpecies             { return extraCompoundSpeciesIn("Ions") }
func Dyes() []ChemSpecies             { return extraCompoundSpeciesIn("Dyes") }
func Proteins() []ChemSpecies         { return extraCompoundSpeciesIn("Proteins") }
func ReactiveSpecies() []ChemSpecies  { return extraCompoundSpeciesIn("Reactive species") }
func Minerals() []ChemSpecies         { return extraCompoundSpeciesIn("Minerals") }

func extraCompoundSpeciesIn(cat string) []ChemSpecies {
	var out []ChemSpecies
	for _, s := range extraCompoundSpecies {
		if s.Category == cat {
			out = append(out, s)
		}
	}
	return out
}

// FindCompoundByName returns the species with the given display name or nil.
func FindCompoundByName(name string) *ChemSpecies {
	for i := range extraCompoundSpecies {
		if extraCompoundSpecies[i].Name == name {
			return &extraCompoundSpecies[i]
		}
	}
	for i := range compoundSpecies {
		if compoundSpecies[i].Name == name {
			return &compoundSpecies[i]
		}
	}
	return nil
}

// ExtraCompoundCount returns how many additional compounds are present.
func ExtraCompoundCount() int { return len(extraCompoundSpecies) }

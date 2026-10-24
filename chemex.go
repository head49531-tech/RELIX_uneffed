package main

import (
	"sort"
	"strings"
)

// ChemSpecies describes one chemical entity in the simulation.
// MolarMass (g/mol) is set for periodic-table elements; compounds leave
// it zero because they are not atoms. IsElement distinguishes the two
// without inferring from the formula string.
type ChemSpecies struct {
	Category    string
	Class       string
	Name        string
	Formula     string
	Toxicity    float64
	Nutrient    float64
	Oxidant     float64
	Radiation   float64
	pH          float64
	Description string
	MolarMass   float64
	IsElement   bool
}

// cmpd constructs a compound entry (mass = 0, not an element).
func cmpd(cat, class, name, formula string, tox, nut, ox, rad, ph float64, desc string) ChemSpecies {
	return ChemSpecies{
		Category: cat, Class: class, Name: name, Formula: formula,
		Toxicity: tox, Nutrient: nut, Oxidant: ox, Radiation: rad,
		pH: ph, Description: desc,
	}
}

// elem constructs a periodic-table entry.
func elem(symbol, name, cat, class string, mass, tox, nut, ox, rad, ph float64, desc string) ChemSpecies {
	return ChemSpecies{
		Category: cat, Class: class, Name: name, Formula: symbol,
		Toxicity: tox, Nutrient: nut, Oxidant: ox, Radiation: rad,
		pH: ph, Description: desc, MolarMass: mass, IsElement: true,
	}
}

// ==================== Periodic table (118 elements) ====================

var elementSpecies = []ChemSpecies{
	// ===== Period 1 =====
	elem("H", "Hydrogen", "Nonmetals", "nonmetal", 1.008, 0.0, 0.2, 0.0, 0.0, 7.0, "Lightest element; fuel"),
	elem("He", "Helium", "Noble gases", "noble", 4.003, 0.0, 0.0, 0.0, 0.0, 7.0, "Inert"),
	// ===== Period 2 =====
	elem("Li", "Lithium", "Alkali metals", "metal", 6.94, 0.4, 0.0, 0.0, 0.0, 7.0, "Reactive; battery metal"),
	elem("Be", "Beryllium", "Alkaline earth", "metal", 9.012, 0.9, 0.0, 0.0, 0.0, 7.0, "Highly toxic dust"),
	elem("B", "Boron", "Metalloids", "metalloid", 10.81, 0.2, 0.2, 0.0, 0.0, 7.0, "Trace element"),
	elem("C", "Carbon", "Nonmetals", "nonmetal", 12.011, 0.0, 0.8, 0.0, 0.0, 7.0, "Backbone of organic chemistry"),
	elem("N", "Nitrogen", "Nonmetals", "nonmetal", 14.007, 0.0, 0.6, 0.0, 0.0, 7.0, "Essential; N2 is inert"),
	elem("O", "Oxygen", "Nonmetals", "nonmetal", 15.999, 0.0, 0.5, 0.5, 0.0, 7.0, "Terminal electron acceptor"),
	elem("F", "Fluorine", "Halogens", "halogen", 18.998, 0.9, 0.0, 1.0, 0.0, 7.0, "Most reactive element"),
	elem("Ne", "Neon", "Noble gases", "noble", 20.18, 0.0, 0.0, 0.0, 0.0, 7.0, "Inert; sign lighting"),
	// ===== Period 3 =====
	elem("Na", "Sodium", "Alkali metals", "metal", 22.99, 0.4, 0.0, 0.0, 0.0, 7.0, "Reactive; osmotic cation"),
	elem("Mg", "Magnesium", "Alkaline earth", "metal", 24.305, 0.0, 0.3, 0.0, 0.0, 7.0, "Cofactor"),
	elem("Al", "Aluminum", "Post-transition", "metal", 26.98, 0.3, 0.0, 0.0, 0.0, 7.0, "Light metal; acidic in solution"),
	elem("Si", "Silicon", "Metalloids", "metalloid", 28.085, 0.0, 0.0, 0.0, 0.0, 7.0, "Semiconductor"),
	elem("P", "Phosphorus", "Nonmetals", "nonmetal", 30.974, 0.1, 0.5, 0.0, 0.0, 7.0, "ATP/phosphate source"),
	elem("S", "Sulfur", "Nonmetals", "nonmetal", 32.06, 0.1, 0.4, 0.0, 0.0, 7.0, "Thiol/disulfide source"),
	elem("Cl", "Chlorine", "Halogens", "halogen", 35.45, 0.8, 0.0, 0.9, 0.0, 7.0, "Bleach precursor"),
	elem("Ar", "Argon", "Noble gases", "noble", 39.948, 0.0, 0.0, 0.0, 0.0, 7.0, "Inert"),
	// ===== Period 4 =====
	elem("K", "Potassium", "Alkali metals", "metal", 39.098, 0.2, 0.4, 0.0, 0.0, 7.0, "Osmolyte"),
	elem("Ca", "Calcium", "Alkaline earth", "metal", 40.078, 0.1, 0.3, 0.0, 0.0, 7.0, "Signaling cation"),
	elem("Sc", "Scandium", "Transition metals", "metal", 44.956, 0.1, 0.0, 0.0, 0.0, 7.0, "Rare earth"),
	elem("Ti", "Titanium", "Transition metals", "metal", 47.867, 0.0, 0.0, 0.0, 0.0, 7.0, "Biocompatible"),
	elem("V", "Vanadium", "Transition metals", "metal", 50.942, 0.3, 0.1, 0.0, 0.0, 7.0, "Trace; vanadate toxic"),
	elem("Cr", "Chromium", "Transition metals", "metal", 51.996, 0.7, 0.1, 0.0, 0.0, 7.0, "Cr6+ is carcinogenic"),
	elem("Mn", "Manganese", "Transition metals", "metal", 54.938, 0.1, 0.2, 0.0, 0.0, 7.0, "Cofactor"),
	elem("Fe", "Iron", "Transition metals", "metal", 55.845, 0.1, 0.4, 0.1, 0.0, 7.0, "Fenton chemistry"),
	elem("Co", "Cobalt", "Transition metals", "metal", 58.933, 0.2, 0.2, 0.0, 0.0, 7.0, "B12 precursor"),
	elem("Ni", "Nickel", "Transition metals", "metal", 58.693, 0.5, 0.0, 0.0, 0.0, 7.0, "Allergen; urease cofactor"),
	elem("Cu", "Copper", "Transition metals", "metal", 63.546, 0.2, 0.3, 0.1, 0.0, 7.0, "Redox-active"),
	elem("Zn", "Zinc", "Transition metals", "metal", 65.38, 0.1, 0.4, 0.0, 0.0, 7.0, "Cofactor; DNA-binding"),
	elem("Ga", "Gallium", "Post-transition", "metal", 69.723, 0.2, 0.0, 0.0, 0.0, 7.0, "Low melting"),
	elem("Ge", "Germanium", "Metalloids", "metalloid", 72.63, 0.0, 0.0, 0.0, 0.0, 7.0, "Semiconductor"),
	elem("As", "Arsenic", "Metalloids", "metalloid", 74.922, 0.8, 0.0, 0.0, 0.05, 7.0, "Toxic; DNA damage"),
	elem("Se", "Selenium", "Nonmetals", "nonmetal", 78.971, 0.3, 0.2, 0.0, 0.0, 7.0, "Trace; selenocysteine"),
	elem("Br", "Bromine", "Halogens", "halogen", 79.904, 0.5, 0.0, 0.6, 0.0, 7.0, "Liquid halogen"),
	elem("Kr", "Krypton", "Noble gases", "noble", 83.798, 0.0, 0.0, 0.0, 0.0, 7.0, "Inert"),
	// ===== Period 5 =====
	elem("Rb", "Rubidium", "Alkali metals", "metal", 85.468, 0.3, 0.0, 0.0, 0.05, 7.0, "Slightly radioactive"),
	elem("Sr", "Strontium", "Alkaline earth", "metal", 87.62, 0.3, 0.0, 0.0, 0.02, 7.0, "Bone-seeking"),
	elem("Y", "Yttrium", "Transition metals", "metal", 88.906, 0.1, 0.0, 0.0, 0.0, 7.0, "Rare earth"),
	elem("Zr", "Zirconium", "Transition metals", "metal", 91.224, 0.0, 0.0, 0.0, 0.0, 7.0, "Corrosion-resistant"),
	elem("Nb", "Niobium", "Transition metals", "metal", 92.906, 0.0, 0.0, 0.0, 0.0, 7.0, "Superconductor"),
	elem("Mo", "Molybdenum", "Transition metals", "metal", 95.95, 0.1, 0.2, 0.0, 0.0, 7.0, "Nitrogenase cofactor"),
	elem("Tc", "Technetium", "Transition metals", "metal", 98, 0.3, 0.0, 0.0, 0.5, 7.0, "First artificial element"),
	elem("Ru", "Ruthenium", "Transition metals", "metal", 101.07, 0.1, 0.0, 0.0, 0.0, 7.0, "Catalyst"),
	elem("Rh", "Rhodium", "Transition metals", "metal", 102.91, 0.1, 0.0, 0.0, 0.0, 7.0, "Catalyst"),
	elem("Pd", "Palladium", "Transition metals", "metal", 106.42, 0.1, 0.0, 0.0, 0.0, 7.0, "Hydrogenation catalyst"),
	elem("Ag", "Silver", "Transition metals", "metal", 107.87, 0.5, 0.0, 0.2, 0.0, 7.0, "Antimicrobial"),
	elem("Cd", "Cadmium", "Transition metals", "metal", 112.41, 0.85, 0.0, 0.0, 0.03, 7.0, "Nephrotoxin"),
	elem("In", "Indium", "Post-transition", "metal", 114.82, 0.1, 0.0, 0.0, 0.0, 7.0, "Soft metal"),
	elem("Sn", "Tin", "Post-transition", "metal", 118.71, 0.2, 0.0, 0.0, 0.0, 7.0, "Organotin toxic"),
	elem("Sb", "Antimony", "Metalloids", "metalloid", 121.76, 0.5, 0.0, 0.0, 0.0, 7.0, "Toxic metalloid"),
	elem("Te", "Tellurium", "Metalloids", "metalloid", 127.6, 0.4, 0.0, 0.0, 0.0, 7.0, "Rare; garlic breath"),
	elem("I", "Iodine", "Halogens", "halogen", 126.9, 0.0, 0.2, 0.0, 0.0, 7.0, "Thyroid precursor"),
	elem("Xe", "Xenon", "Noble gases", "noble", 131.29, 0.0, 0.0, 0.0, 0.0, 7.0, "Anesthetic gas"),
	// ===== Period 6 =====
	elem("Cs", "Cesium", "Alkali metals", "metal", 132.91, 0.2, 0.0, 0.0, 0.1, 7.0, "Most reactive metal"),
	elem("Ba", "Barium", "Alkaline earth", "metal", 137.33, 0.5, 0.0, 0.0, 0.02, 7.0, "Toxic cation"),
	elem("La", "Lanthanum", "Lanthanides", "metal", 138.91, 0.1, 0.0, 0.0, 0.0, 7.0, "Rare earth"),
	elem("Ce", "Cerium", "Lanthanides", "metal", 140.12, 0.1, 0.0, 0.0, 0.0, 7.0, "Oxide polishing"),
	elem("Pr", "Praseodymium", "Lanthanides", "metal", 140.91, 0.1, 0.0, 0.0, 0.0, 7.0, "Rare earth"),
	elem("Nd", "Neodymium", "Lanthanides", "metal", 144.24, 0.1, 0.0, 0.0, 0.0, 7.0, "Magnets"),
	elem("Pm", "Promethium", "Lanthanides", "metal", 145, 0.3, 0.0, 0.0, 0.5, 7.0, "Radioactive lanthanide"),
	elem("Sm", "Samarium", "Lanthanides", "metal", 150.36, 0.1, 0.0, 0.0, 0.0, 7.0, "Magnet alloy"),
	elem("Eu", "Europium", "Lanthanides", "metal", 151.96, 0.1, 0.0, 0.0, 0.0, 7.0, "Phosphor"),
	elem("Gd", "Gadolinium", "Lanthanides", "metal", 157.25, 0.1, 0.0, 0.0, 0.0, 7.0, "MRI contrast"),
	elem("Tb", "Terbium", "Lanthanides", "metal", 158.93, 0.1, 0.0, 0.0, 0.0, 7.0, "Green phosphor"),
	elem("Dy", "Dysprosium", "Lanthanides", "metal", 162.5, 0.1, 0.0, 0.0, 0.0, 7.0, "Magnet additive"),
	elem("Ho", "Holmium", "Lanthanides", "metal", 164.93, 0.1, 0.0, 0.0, 0.0, 7.0, "Laser crystal"),
	elem("Er", "Erbium", "Lanthanides", "metal", 167.26, 0.1, 0.0, 0.0, 0.0, 7.0, "Fiber optics"),
	elem("Tm", "Thulium", "Lanthanides", "metal", 168.93, 0.1, 0.0, 0.0, 0.0, 7.0, "Rare"),
	elem("Yb", "Ytterbium", "Lanthanides", "metal", 173.05, 0.1, 0.0, 0.0, 0.0, 7.0, "Atomic clock"),
	elem("Lu", "Lutetium", "Lanthanides", "metal", 174.97, 0.1, 0.0, 0.0, 0.0, 7.0, "Densest lanthanide"),
	elem("Hf", "Hafnium", "Transition metals", "metal", 178.49, 0.1, 0.0, 0.0, 0.0, 7.0, "Nuclear control rod"),
	elem("Ta", "Tantalum", "Transition metals", "metal", 180.95, 0.0, 0.0, 0.0, 0.0, 7.0, "Capacitors"),
	elem("W", "Tungsten", "Transition metals", "metal", 183.84, 0.1, 0.0, 0.0, 0.0, 7.0, "Highest melting point"),
	elem("Re", "Rhenium", "Transition metals", "metal", 186.21, 0.0, 0.0, 0.0, 0.0, 7.0, "Superalloy"),
	elem("Os", "Osmium", "Transition metals", "metal", 190.23, 0.5, 0.0, 0.1, 0.0, 7.0, "OsO4 is highly toxic"),
	elem("Ir", "Iridium", "Transition metals", "metal", 192.22, 0.1, 0.0, 0.0, 0.0, 7.0, "Corrosion-resistant"),
	elem("Pt", "Platinum", "Transition metals", "metal", 195.08, 0.1, 0.0, 0.0, 0.0, 7.0, "Catalyst"),
	elem("Au", "Gold", "Transition metals", "metal", 196.97, 0.0, 0.0, 0.0, 0.0, 7.0, "Noble; biocompatible"),
	elem("Hg", "Mercury", "Transition metals", "metal", 200.59, 0.95, 0.0, 0.0, 0.05, 7.0, "Thiol-reactive neurotoxin"),
	elem("Tl", "Thallium", "Post-transition", "metal", 204.38, 0.9, 0.0, 0.0, 0.0, 7.0, "Highly toxic"),
	elem("Pb", "Lead", "Post-transition", "metal", 207.2, 0.85, 0.0, 0.0, 0.05, 7.0, "Neurotoxin"),
	elem("Bi", "Bismuth", "Post-transition", "metal", 208.98, 0.2, 0.0, 0.0, 0.0, 7.0, "Low toxicity heavy metal"),
	elem("Po", "Polonium", "Post-transition", "metal", 209, 0.95, 0.0, 0.0, 0.85, 7.0, "Alpha emitter"),
	elem("At", "Astatine", "Halogens", "halogen", 210, 0.9, 0.0, 0.5, 0.5, 7.0, "Rarest naturally-occurring element"),
	elem("Rn", "Radon", "Noble gases", "noble", 222, 0.8, 0.0, 0.0, 0.9, 7.0, "Noble radioactive gas"),
	// ===== Period 7 =====
	elem("Fr", "Francium", "Alkali metals", "metal", 223, 0.7, 0.0, 0.0, 0.6, 7.0, "Highly radioactive"),
	elem("Ra", "Radium", "Alkaline earth", "metal", 226, 0.9, 0.0, 0.0, 0.8, 7.0, "Bone-seeking alpha emitter"),
	elem("Ac", "Actinium", "Actinides", "metal", 227, 0.5, 0.0, 0.0, 0.5, 7.0, "Radioactive"),
	elem("Th", "Thorium", "Actinides", "metal", 232.04, 0.6, 0.0, 0.0, 0.25, 7.0, "Fertile nuclear fuel"),
	elem("Pa", "Protactinium", "Actinides", "metal", 231.04, 0.6, 0.0, 0.0, 0.4, 7.0, "Radioactive"),
	elem("U", "Uranium", "Actinides", "metal", 238.03, 0.7, 0.0, 0.0, 0.3, 7.0, "Fissile alpha emitter"),
	elem("Np", "Neptunium", "Actinides", "metal", 237, 0.8, 0.0, 0.0, 0.5, 7.0, "Transuranic"),
	elem("Pu", "Plutonium", "Actinides", "metal", 244, 0.9, 0.0, 0.0, 0.6, 7.0, "Fissile alpha emitter"),
	elem("Am", "Americium", "Actinides", "metal", 243, 0.7, 0.0, 0.0, 0.4, 7.0, "Smoke detector source"),
	elem("Cm", "Curium", "Actinides", "metal", 247, 0.7, 0.0, 0.0, 0.7, 7.0, "Radioactive"),
	elem("Bk", "Berkelium", "Actinides", "metal", 247, 0.6, 0.0, 0.0, 0.4, 7.0, "Synthetic"),
	elem("Cf", "Californium", "Actinides", "metal", 251, 0.7, 0.0, 0.0, 0.5, 7.0, "Neutron source"),
	elem("Es", "Einsteinium", "Actinides", "metal", 252, 0.6, 0.0, 0.0, 0.4, 7.0, "Synthetic"),
	elem("Fm", "Fermium", "Actinides", "metal", 257, 0.5, 0.0, 0.0, 0.3, 7.0, "Synthetic"),
	elem("Md", "Mendelevium", "Actinides", "metal", 258, 0.5, 0.0, 0.0, 0.3, 7.0, "Synthetic"),
	elem("No", "Nobelium", "Actinides", "metal", 259, 0.5, 0.0, 0.0, 0.3, 7.0, "Synthetic"),
	elem("Lr", "Lawrencium", "Actinides", "metal", 266, 0.5, 0.0, 0.0, 0.3, 7.0, "Synthetic"),
	elem("Rf", "Rutherfordium", "Transition metals", "metal", 267, 0.5, 0.0, 0.0, 0.3, 7.0, "Synthetic"),
	elem("Db", "Dubnium", "Transition metals", "metal", 268, 0.5, 0.0, 0.0, 0.3, 7.0, "Synthetic"),
	elem("Sg", "Seaborgium", "Transition metals", "metal", 269, 0.5, 0.0, 0.0, 0.3, 7.0, "Synthetic"),
	elem("Bh", "Bohrium", "Transition metals", "metal", 270, 0.5, 0.0, 0.0, 0.3, 7.0, "Synthetic"),
	elem("Hs", "Hassium", "Transition metals", "metal", 269, 0.5, 0.0, 0.0, 0.3, 7.0, "Synthetic"),
	elem("Mt", "Meitnerium", "Transition metals", "metal", 278, 0.5, 0.0, 0.0, 0.3, 7.0, "Synthetic"),
	elem("Ds", "Darmstadtium", "Transition metals", "metal", 281, 0.5, 0.0, 0.0, 0.3, 7.0, "Synthetic"),
	elem("Rg", "Roentgenium", "Transition metals", "metal", 282, 0.5, 0.0, 0.0, 0.3, 7.0, "Synthetic"),
	elem("Cn", "Copernicium", "Transition metals", "metal", 285, 0.5, 0.0, 0.0, 0.3, 7.0, "Synthetic"),
	elem("Nh", "Nihonium", "Post-transition", "metal", 286, 0.5, 0.0, 0.0, 0.3, 7.0, "Synthetic"),
	elem("Fl", "Flerovium", "Post-transition", "metal", 289, 0.5, 0.0, 0.0, 0.3, 7.0, "Synthetic"),
	elem("Mc", "Moscovium", "Post-transition", "metal", 290, 0.5, 0.0, 0.0, 0.3, 7.0, "Synthetic"),
	elem("Lv", "Livermorium", "Post-transition", "metal", 293, 0.5, 0.0, 0.0, 0.3, 7.0, "Synthetic"),
	elem("Ts", "Tennessine", "Halogens", "halogen", 294, 0.7, 0.0, 0.3, 0.3, 7.0, "Synthetic halogen-like"),
	elem("Og", "Oganesson", "Noble gases", "noble", 294, 0.6, 0.0, 0.0, 0.3, 7.0, "Synthetic noble gas"),
}

// ==================== Molecular compounds (existing table) ====================

var compoundSpecies = []ChemSpecies{
	// -------- Water & ROS --------
	cmpd("Water & ROS", "neutral", "Water", "H2O", 0.0, 1.0, 0.0, 0.0, 7.0, "Neutral; mild nutrient"),
	cmpd("Water & ROS", "neutral", "Deuterium oxide", "D2O", 0.3, 0.5, 0.0, 0.0, 7.0, "Heavy water"),
	cmpd("Water & ROS", "oxidizer", "Hydrogen peroxide", "H2O2", 0.3, 0.0, 0.8, 0.0, 5.5, "Oxidant"),
	cmpd("Water & ROS", "oxidizer", "Hydroxyl radical", "OH*", 0.9, 0.0, 1.0, 0.0, 7.0, "Highly reactive ROS"),
	cmpd("Water & ROS", "oxidizer", "Superoxide", "O2-", 0.4, 0.0, 0.9, 0.0, 8.0, "ROS"),
	cmpd("Water & ROS", "oxidizer", "Peroxynitrite", "ONOO-", 0.8, 0.0, 1.0, 0.0, 7.0, "RNS/ROS hybrid"),
	cmpd("Water & ROS", "oxidizer", "Singlet oxygen", "1O2", 0.7, 0.0, 0.9, 0.0, 7.0, "Excited-state O2"),

	// -------- Gases --------
	cmpd("Gases", "neutral", "Hydrogen gas", "H2", 0.0, 0.3, 0.0, 0.0, 7.0, "Electron donor"),
	cmpd("Gases", "oxidizer", "Oxygen", "O2", 0.0, 0.8, 0.0, 0.0, 7.0, "Terminal acceptor"),
	cmpd("Gases", "oxidizer", "Ozone", "O3", 0.6, 0.0, 0.9, 0.0, 7.0, "Strong oxidant"),
	cmpd("Gases", "oxidizer", "Chlorine gas", "Cl2", 0.9, 0.0, 0.7, 0.0, 3.0, "WWI gas"),
	cmpd("Gases", "neutral", "Nitric oxide", "NO", 0.3, 0.0, 0.5, 0.1, 7.0, "Signaling"),
	cmpd("Gases", "acid", "Sulfur dioxide", "SO2", 0.4, 0.0, 0.4, 0.0, 3.5, "Irritant"),
	cmpd("Gases", "acid", "Hydrogen sulfide", "H2S", 0.5, 0.0, 0.3, 0.0, 4.0, "Signaling at low conc"),
	cmpd("Gases", "base", "Ammonia", "NH3", 0.4, 0.5, 0.0, 0.0, 11.0, "Nitrogen source"),
	cmpd("Gases", "acid", "Carbon dioxide", "CO2", 0.0, 0.2, 0.0, 0.0, 4.0, "Respiratory"),
	cmpd("Gases", "neutral", "Carbon monoxide", "CO", 0.9, 0.0, 0.0, 0.0, 7.0, "Hemoglobin binder"),
	cmpd("Gases", "acid", "Hydrogen cyanide", "HCN", 0.95, 0.0, 0.0, 0.0, 5.0, "Cytochrome oxidase inhibitor"),
	cmpd("Gases", "acid", "Nitrogen dioxide", "NO2", 0.7, 0.0, 0.4, 0.0, 3.0, "Irritant"),
	cmpd("Gases", "neutral", "Methane", "CH4", 0.1, 0.3, 0.0, 0.0, 7.0, "Anaerobic product"),

	// -------- Acids --------
	cmpd("Acids", "acid", "Proton", "H+", 0.1, 0.0, 0.2, 0.0, 1.0, "pH effect"),
	cmpd("Acids", "acid", "Hydrochloric acid", "HCl", 0.6, 0.0, 0.1, 0.0, 1.0, "Strong acid"),
	cmpd("Acids", "acid", "Sulfuric acid", "H2SO4", 0.7, 0.0, 0.1, 0.0, 1.0, "Strong acid"),
	cmpd("Acids", "acid", "Nitric acid", "HNO3", 0.7, 0.0, 0.3, 0.0, 1.0, "Oxidizing acid"),
	cmpd("Acids", "acid", "Phosphoric acid", "H3PO4", 0.5, 0.0, 0.1, 0.0, 2.0, "Weak acid"),
	cmpd("Acids", "acid", "Acetic acid", "AcOH", 0.3, 0.2, 0.1, 0.0, 3.0, "Weak acid"),
	cmpd("Acids", "acid", "Lactic acid", "LacH", 0.2, 0.3, 0.1, 0.0, 3.5, "Fermentation"),
	cmpd("Acids", "acid", "Citric acid", "CitH", 0.2, 0.3, 0.1, 0.0, 3.5, "TCA cycle"),
	cmpd("Acids", "acid", "Formic acid", "HCOOH", 0.5, 0.0, 0.1, 0.0, 2.0, "C1 acid"),
	cmpd("Acids", "acid", "Oxalic acid", "H2C2O4", 0.6, 0.0, 0.1, 0.0, 1.5, "Dicarboxylic"),
	cmpd("Acids", "acid", "Trichloroacetic acid", "TCA", 0.6, 0.0, 0.1, 0.0, 0.5, "Protein precipitant"),

	// -------- Bases --------
	cmpd("Bases", "base", "Sodium hydroxide", "NaOH", 0.7, 0.0, 0.0, 0.0, 13.0, "Strong base"),
	cmpd("Bases", "base", "Potassium hydroxide", "KOH", 0.7, 0.0, 0.0, 0.0, 13.0, "Strong base"),
	cmpd("Bases", "base", "Calcium hydroxide", "Ca(OH)2", 0.5, 0.0, 0.0, 0.0, 12.0, "Strong base"),
	cmpd("Bases", "base", "Ammonium hydroxide", "NH4OH", 0.4, 0.0, 0.0, 0.0, 11.0, "Weak base"),
	cmpd("Bases", "base", "Lithium hydroxide", "LiOH", 0.7, 0.0, 0.0, 0.0, 13.0, "Strong base"),
	cmpd("Bases", "base", "Barium hydroxide", "Ba(OH)2", 0.6, 0.0, 0.0, 0.0, 12.0, "Strong base"),

	// -------- Carbonates --------
	cmpd("Carbonates", "carbonate", "Sodium carbonate", "Na2CO3", 0.2, 0.0, 0.0, 0.0, 11.0, "Basic salt"),
	cmpd("Carbonates", "carbonate", "Sodium bicarbonate", "NaHCO3", 0.1, 0.0, 0.0, 0.0, 8.3, "Buffer"),
	cmpd("Carbonates", "carbonate", "Calcium carbonate", "CaCO3", 0.0, 0.0, 0.0, 0.0, 9.0, "Shells; chalk"),
	cmpd("Carbonates", "carbonate", "Potassium bicarbonate", "KHCO3", 0.1, 0.0, 0.0, 0.0, 8.3, "Buffer"),

	// -------- Chlorine & Halogens --------
	cmpd("Chlorine & Halogens", "oxidizer", "Hypochlorous acid", "HOCl", 0.5, 0.0, 0.9, 0.0, 4.0, "Bleach"),
	cmpd("Chlorine & Halogens", "oxidizer", "Sodium hypochlorite", "NaClO", 0.4, 0.0, 0.8, 0.0, 11.0, "Household bleach"),
	cmpd("Chlorine & Halogens", "oxidizer", "Chlorine dioxide", "ClO2", 0.6, 0.0, 1.0, 0.0, 3.0, "Industrial oxidant"),
	cmpd("Chlorine & Halogens", "neutral", "Chloride ion", "Cl-", 0.0, 0.4, 0.0, 0.0, 7.0, "Essential anion"),
	cmpd("Chlorine & Halogens", "neutral", "Fluoride ion", "F-", 0.4, 0.0, 0.0, 0.0, 8.0, "Enolase inhibitor"),
	cmpd("Chlorine & Halogens", "neutral", "Bromide ion", "Br-", 0.1, 0.1, 0.0, 0.0, 7.0, "Trace nutrient"),
	cmpd("Chlorine & Halogens", "neutral", "Iodide ion", "I-", 0.0, 0.2, 0.0, 0.0, 7.0, "Thyroid precursor"),

	// -------- Sulfur --------
	cmpd("Sulfur", "neutral", "Elemental sulfur", "S8", 0.0, 0.0, 0.0, 0.0, 7.0, "Inert solid"),
	cmpd("Sulfur", "base", "Sodium sulfide", "Na2S", 0.6, 0.0, 0.2, 0.0, 12.0, "H2S source"),
	cmpd("Sulfur", "neutral", "DMSO", "DMSO", 0.2, 0.0, 0.0, 0.0, 7.0, "Cryoprotectant"),

	// -------- Nitrogen --------
	cmpd("Nitrogen", "nutrient", "Nitrate", "NO3-", 0.0, 0.6, 0.0, 0.0, 7.0, "Nitrogen source"),
	cmpd("Nitrogen", "nutrient", "Nitrite", "NO2-", 0.3, 0.4, 0.2, 0.0, 7.5, "Toxic >1mM"),
	cmpd("Nitrogen", "neutral", "Urea", "Urea", 0.6, 0.0, 0.0, 0.0, 7.2, "Chaotrope"),

	// -------- Essential metals --------
	cmpd("Essential metals", "metal", "Magnesium ion", "Mg2+", 0.0, 0.3, 0.0, 0.0, 7.0, "Cofactor"),
	cmpd("Essential metals", "metal", "Calcium ion", "Ca2+", 0.1, 0.2, 0.0, 0.0, 7.0, "Signaling"),
	cmpd("Essential metals", "metal", "Potassium ion", "K+", 0.0, 0.3, 0.0, 0.0, 7.0, "Osmolyte"),
	cmpd("Essential metals", "metal", "Iron(II)", "Fe2+", 0.1, 0.5, 0.1, 0.0, 6.0, "Fenton"),
	cmpd("Essential metals", "metal", "Iron(III)", "Fe3+", 0.1, 0.4, 0.0, 0.0, 4.0, "Carrier"),
	cmpd("Essential metals", "metal", "Zinc ion", "Zn2+", 0.1, 0.4, 0.0, 0.0, 6.0, "Cofactor"),
	cmpd("Essential metals", "metal", "Copper(II)", "Cu2+", 0.2, 0.3, 0.2, 0.0, 5.5, "Redox-active"),
	cmpd("Essential metals", "metal", "Cobalt(II)", "Co2+", 0.2, 0.3, 0.0, 0.0, 6.0, "B12 precursor"),
	cmpd("Essential metals", "metal", "Manganese(II)", "Mn2+", 0.1, 0.3, 0.0, 0.0, 6.0, "Cofactor"),
	cmpd("Essential metals", "metal", "Barium ion", "Ba2+", 0.6, 0.0, 0.0, 0.0, 7.0, "Toxic metal"),
	cmpd("Essential metals", "acid", "Aluminum ion", "Al3+", 0.4, 0.0, 0.0, 0.0, 3.0, "Hydrolyzes acidic"),
	cmpd("Essential metals", "metal", "Silver ion", "Ag+", 0.7, 0.0, 0.2, 0.0, 6.0, "Antimicrobial"),

	// -------- Heavy metals --------
	cmpd("Heavy metals", "metal", "Lead ion", "Pb2+", 0.85, 0.0, 0.0, 0.05, 6.0, "Neurotoxin"),
	cmpd("Heavy metals", "metal", "Mercury ion", "Hg2+", 0.95, 0.0, 0.0, 0.05, 5.0, "Thiol-reactive"),
	cmpd("Heavy metals", "metal", "Cadmium ion", "Cd2+", 0.85, 0.0, 0.0, 0.03, 6.0, "Nephrotoxin"),
	cmpd("Heavy metals", "metal", "Arsenate", "AsO4", 0.70, 0.0, 0.0, 0.10, 7.0, "DNA damage"),
	cmpd("Heavy metals", "oxidizer", "Chromate", "Cr6+", 0.90, 0.0, 0.3, 0.00, 4.0, "Carcinogen"),
	cmpd("Heavy metals", "metal", "Nickel ion", "Ni2+", 0.60, 0.0, 0.0, 0.02, 6.0, "Allergen"),

	// -------- Salts & Buffers --------
	cmpd("Salts & Buffers", "salt", "Sodium chloride", "NaCl", 0.0, 0.3, 0.0, 0.0, 7.0, "Osmolyte"),
	cmpd("Salts & Buffers", "salt", "Potassium chloride", "KCl", 0.0, 0.3, 0.0, 0.0, 7.0, "K source"),
	cmpd("Salts & Buffers", "salt", "Magnesium chloride", "MgCl2", 0.0, 0.3, 0.0, 0.0, 7.0, "Cofactor"),
	cmpd("Salts & Buffers", "salt", "Calcium chloride", "CaCl2", 0.1, 0.2, 0.0, 0.0, 7.0, "Signaling"),
	cmpd("Salts & Buffers", "salt", "Tris buffer", "Tris", 0.0, 0.0, 0.0, 0.0, 7.5, "pH buffering"),
	cmpd("Salts & Buffers", "salt", "Sodium sulfate", "Na2SO4", 0.1, 0.0, 0.0, 0.0, 7.0, "Salt"),
	cmpd("Salts & Buffers", "salt", "Potassium nitrate", "KNO3", 0.1, 0.2, 0.0, 0.0, 7.0, "Oxidizer salt"),
	cmpd("Salts & Buffers", "salt", "Sodium nitrate", "NaNO3", 0.1, 0.2, 0.0, 0.0, 7.0, "Preservative"),
	cmpd("Salts & Buffers", "salt", "Ammonium chloride", "NH4Cl", 0.2, 0.1, 0.0, 0.0, 5.5, "Slightly acidic"),
	cmpd("Salts & Buffers", "salt", "Copper sulfate", "CuSO4", 0.4, 0.0, 0.1, 0.0, 4.5, "Algicide"),
	cmpd("Salts & Buffers", "salt", "Iron sulfate", "FeSO4", 0.3, 0.1, 0.0, 0.0, 4.0, "Iron supplement"),
	cmpd("Salts & Buffers", "salt", "Zinc sulfate", "ZnSO4", 0.2, 0.1, 0.0, 0.0, 5.5, "Zinc supplement"),
	cmpd("Salts & Buffers", "salt", "Manganese sulfate", "MnSO4", 0.2, 0.1, 0.0, 0.0, 5.5, "Mn supplement"),

	// -------- Antibiotics --------
	cmpd("Antibiotics", "neutral", "Ampicillin", "Amp", 0.6, 0.0, 0.0, 0.0, 7.0, "beta-lactam"),
	cmpd("Antibiotics", "neutral", "Kanamycin", "Kan", 0.6, 0.0, 0.0, 0.0, 7.0, "Aminoglycoside"),
	cmpd("Antibiotics", "neutral", "Chloramphenicol", "Cm", 0.5, 0.0, 0.0, 0.0, 7.0, "Ribosome"),
	cmpd("Antibiotics", "neutral", "Tetracycline", "Tet", 0.5, 0.0, 0.0, 0.0, 7.0, "Bacteriostatic"),
	cmpd("Antibiotics", "neutral", "Puromycin", "Puro", 0.7, 0.0, 0.0, 0.0, 7.0, "Premature stop"),
	cmpd("Antibiotics", "neutral", "Rifampicin", "Rif", 0.6, 0.0, 0.0, 0.0, 7.0, "RNA pol"),
	cmpd("Antibiotics", "neutral", "Ciprofloxacin", "Cip", 0.6, 0.0, 0.0, 0.0, 7.0, "Gyrase"),
	cmpd("Antibiotics", "neutral", "Streptomycin", "Str", 0.7, 0.0, 0.0, 0.0, 7.0, "Aminoglycoside"),
	cmpd("Antibiotics", "neutral", "Erythromycin", "Ery", 0.6, 0.0, 0.0, 0.0, 7.0, "Macrolide"),
	cmpd("Antibiotics", "neutral", "Vancomycin", "Van", 0.8, 0.0, 0.0, 0.0, 7.0, "Cell wall"),
	cmpd("Antibiotics", "neutral", "Azithromycin", "Azi", 0.6, 0.0, 0.0, 0.0, 7.0, "Macrolide"),

	// -------- Solvents & Detergents --------
	cmpd("Solvents & Detergents", "organic", "Ethanol", "EtOH", 0.5, 0.0, 0.1, 0.0, 7.0, "Membrane disrupt"),
	cmpd("Solvents & Detergents", "organic", "Methanol", "MeOH", 0.7, 0.0, 0.1, 0.0, 7.0, "Toxic solvent"),
	cmpd("Solvents & Detergents", "detergent", "SDS", "SDS", 0.9, 0.0, 0.0, 0.0, 7.0, "Membrane disrupt"),
	cmpd("Solvents & Detergents", "detergent", "Triton X-100", "TX100", 0.7, 0.0, 0.0, 0.0, 7.0, "Nonionic detergent"),
	cmpd("Solvents & Detergents", "organic", "Chloroform", "CHCl3", 0.6, 0.0, 0.0, 0.0, 7.0, "Hepatotoxic"),
	cmpd("Solvents & Detergents", "organic", "Formaldehyde", "FA", 0.7, 0.0, 0.0, 0.0, 7.0, "Crosslinker"),
	cmpd("Solvents & Detergents", "organic", "Acetonitrile", "ACN", 0.5, 0.0, 0.0, 0.0, 7.0, "Aprotic"),
	cmpd("Solvents & Detergents", "organic", "Benzene", "Bz", 0.8, 0.0, 0.0, 0.0, 7.0, "Carcinogen"),
	cmpd("Solvents & Detergents", "organic", "Toluene", "Tol", 0.6, 0.0, 0.0, 0.0, 7.0, "Solvent"),
	cmpd("Solvents & Detergents", "acid", "Phenol", "PhOH", 0.7, 0.0, 0.0, 0.0, 6.0, "Toxic"),
	cmpd("Solvents & Detergents", "organic", "Acetone", "Ac2O", 0.3, 0.0, 0.0, 0.0, 7.0, "Polar aprotic"),
	cmpd("Solvents & Detergents", "organic", "Isopropanol", "iPrOH", 0.3, 0.0, 0.0, 0.0, 7.0, "Disinfectant"),

	// -------- Sugars & Nutrients --------
	cmpd("Sugars & Nutrients", "nutrient", "Glucose", "Glc", 0.0, 1.0, 0.0, 0.0, 7.0, "Primary carbon"),
	cmpd("Sugars & Nutrients", "nutrient", "Fructose", "Fru", 0.0, 0.8, 0.0, 0.0, 7.0, "Carbon"),
	cmpd("Sugars & Nutrients", "nutrient", "Sucrose", "Suc", 0.0, 0.7, 0.0, 0.0, 7.0, "Disaccharide"),
	cmpd("Sugars & Nutrients", "nutrient", "Glycerol", "Gly", 0.0, 0.5, 0.0, 0.0, 7.0, "Osmolyte"),
	cmpd("Sugars & Nutrients", "nutrient", "Ribose", "Rib", 0.0, 0.6, 0.0, 0.0, 7.0, "RNA sugar"),

	// -------- Chelators & Inhibitors --------
	cmpd("Chelators & Inhibitors", "chelator", "EDTA", "EDTA", 0.3, 0.0, 0.0, 0.0, 7.0, "Divalent chelator"),
	cmpd("Chelators & Inhibitors", "chelator", "EGTA", "EGTA", 0.2, 0.0, 0.0, 0.0, 7.0, "Ca chelator"),
	cmpd("Chelators & Inhibitors", "chelator", "DTT", "DTT", 0.2, 0.0, 0.0, 0.0, 7.0, "Reducing"),
	cmpd("Chelators & Inhibitors", "chelator", "BME", "BME", 0.2, 0.0, 0.0, 0.0, 7.0, "Reducing"),
	cmpd("Chelators & Inhibitors", "inhibitor", "Sodium azide", "NaN3", 0.8, 0.0, 0.0, 0.0, 7.0, "Cox inhibitor"),
	cmpd("Chelators & Inhibitors", "chelator", "Deferoxamine", "DFO", 0.3, 0.0, 0.0, 0.0, 7.0, "Fe chelator"),
	cmpd("Chelators & Inhibitors", "chelator", "Dimercaprol", "BAL", 0.4, 0.0, 0.0, 0.0, 7.0, "Heavy-metal antidote"),

	// -------- Radioactive --------
	cmpd("Radioactive", "radioactive", "Uranium-238", "U238", 0.70, 0.0, 0.0, 0.30, 7.0, "Alpha"),
	cmpd("Radioactive", "radioactive", "Uranium-235", "U235", 0.70, 0.0, 0.0, 0.40, 7.0, "Fissile"),
	cmpd("Radioactive", "radioactive", "Plutonium-239", "Pu239", 0.90, 0.0, 0.0, 0.60, 7.0, "Alpha+fission"),
	cmpd("Radioactive", "radioactive", "Radium-226", "Ra226", 0.90, 0.0, 0.0, 0.80, 7.0, "Alpha"),
	cmpd("Radioactive", "radioactive", "Radon-222", "Rn222", 0.80, 0.0, 0.0, 0.90, 7.0, "Noble gas"),
	cmpd("Radioactive", "radioactive", "Cesium-137", "Cs137", 0.80, 0.0, 0.0, 0.70, 7.0, "Gamma"),
	cmpd("Radioactive", "radioactive", "Strontium-90", "Sr90", 0.70, 0.0, 0.0, 0.50, 7.0, "Beta; bone"),
	cmpd("Radioactive", "radioactive", "Iodine-131", "I131", 0.70, 0.0, 0.0, 0.60, 7.0, "Beta+gamma"),
	cmpd("Radioactive", "radioactive", "Cobalt-60", "Co60", 0.60, 0.0, 0.0, 0.90, 7.0, "Gamma"),
	cmpd("Radioactive", "radioactive", "Americium-241", "Am241", 0.70, 0.0, 0.0, 0.40, 7.0, "Alpha"),
	cmpd("Radioactive", "radioactive", "Thorium-232", "Th232", 0.60, 0.0, 0.0, 0.25, 7.0, "Alpha"),
	cmpd("Radioactive", "radioactive", "Phosphorus-32", "P32", 0.70, 0.0, 0.0, 0.60, 7.0, "Beta"),
	cmpd("Radioactive", "radioactive", "Polonium-210", "Po210", 0.95, 0.0, 0.0, 0.85, 7.0, "Alpha"),
	cmpd("Radioactive", "radioactive", "Technetium-99m", "Tc99m", 0.30, 0.0, 0.0, 0.55, 7.0, "SPECT tracer"),
	cmpd("Radioactive", "radioactive", "Gallium-67", "Ga67", 0.40, 0.0, 0.0, 0.45, 7.0, "Imaging isotope"),
	cmpd("Radioactive", "radioactive", "Yttrium-90", "Y90", 0.60, 0.0, 0.0, 0.75, 7.0, "Beta therapy"),
	cmpd("Radioactive", "radioactive", "Strontium-89", "Sr89", 0.55, 0.0, 0.0, 0.55, 7.0, "Bone palliative"),
	cmpd("Radioactive", "radioactive", "Lutetium-177", "Lu177", 0.50, 0.0, 0.0, 0.60, 7.0, "Therapy"),
}

// ==================== Registry ====================

// ChemSpeciesTable is compounds followed by elements. Elements are at the
// end so periodic-table entries appear at the bottom of the Chemex scroll
// without disrupting the existing compound grouping.
var ChemSpeciesTable = func() []ChemSpecies {
	out := make([]ChemSpecies, 0, len(compoundSpecies)+len(extraCompoundSpecies)+len(elementSpecies))
	out = append(out, compoundSpecies...)
	out = append(out, extraCompoundSpecies...)
	out = append(out, elementSpecies...)
	return out
}()

// ==================== Chemex ====================

type Chemex struct {
	Concentrations map[string]float64
}

func NewChemex() *Chemex {
	c := &Chemex{Concentrations: map[string]float64{}}
	for _, s := range ChemSpeciesTable {
		c.Concentrations[s.Formula] = 0
	}
	return c
}

// ==================== Aggregates ====================

// MaxNutrientMult caps NutrientMultiplier so that piling up trace nutrients
// cannot fabricate food from nothing. Set to 5x baseline.
const maxNutrientMult = 5.0

func (c *Chemex) TotalToxicity() float64 {
	total := 0.0
	for _, s := range ChemSpeciesTable {
		conc := c.Concentrations[s.Formula]
		if conc <= 0 {
			continue
		}
		total += s.Toxicity * conc
	}
	return total
}

func (c *Chemex) TotalOxidant() float64 {
	total := 0.0
	for _, s := range ChemSpeciesTable {
		conc := c.Concentrations[s.Formula]
		if conc <= 0 {
			continue
		}
		total += s.Oxidant * conc
	}
	return total
}

func (c *Chemex) TotalRadiation() float64 {
	total := 0.0
	for _, s := range ChemSpeciesTable {
		conc := c.Concentrations[s.Formula]
		if conc <= 0 {
			continue
		}
		total += s.Radiation * conc
	}
	return total
}

func (c *Chemex) NutrientMultiplier() float64 {
	m := 1.0
	for _, s := range ChemSpeciesTable {
		conc := c.Concentrations[s.Formula]
		if conc <= 0 {
			continue
		}
		m += s.Nutrient * conc * 0.05
	}
	if m > maxNutrientMult {
		return maxNutrientMult
	}
	return m
}

// TotalMolarMass sums mass concentration across species that carry a
// mass. Compounds have zero mass so only element concentrations count.
// Useful for sanity-checking closed-system mass balance.
func (c *Chemex) TotalMolarMass() float64 {
	total := 0.0
	for _, s := range ChemSpeciesTable {
		if s.MolarMass <= 0 {
			continue
		}
		conc := c.Concentrations[s.Formula]
		if conc <= 0 {
			continue
		}
		total += s.MolarMass * conc
	}
	return total
}

// PredictedPH estimates environmental pH from the balance of acids and
// bases present. Returns 7.0 when nothing acid or basic is present.
func (c *Chemex) PredictedPH() float64 {
	acid, base := 0.0, 0.0
	for _, s := range ChemSpeciesTable {
		conc := c.Concentrations[s.Formula]
		if conc <= 0 {
			continue
		}
		switch s.Class {
		case "acid":
			acid += conc
		case "base":
			base += conc
		}
	}
	total := acid + base
	if total < 1e-9 {
		return 7.0
	}
	ph := 7.0 - 7.0*(acid-base)/total
	if ph < 0 {
		ph = 0
	}
	if ph > 14 {
		ph = 14
	}
	return ph
}

// ==================== Lookups / filters ====================

// ElementsOnly returns the periodic-table entries.
func ElementsOnly() []ChemSpecies {
	out := make([]ChemSpecies, 0, len(elementSpecies))
	for _, s := range ChemSpeciesTable {
		if s.IsElement {
			out = append(out, s)
		}
	}
	return out
}

// CompoundsOnly returns molecular (non-element) species.
func CompoundsOnly() []ChemSpecies {
	out := make([]ChemSpecies, 0, len(compoundSpecies))
	for _, s := range ChemSpeciesTable {
		if !s.IsElement {
			out = append(out, s)
		}
	}
	return out
}

// ChemCategories returns the sorted unique categories in the table.
func ChemCategories() []string {
	set := map[string]bool{}
	for _, s := range ChemSpeciesTable {
		set[s.Category] = true
	}
	out := make([]string, 0, len(set))
	for c := range set {
		out = append(out, c)
	}
	sort.Strings(out)
	return out
}

// FindSpeciesByFormula returns a pointer to the species with the given
// formula, or nil. Uses a linear scan because the table is small.
func FindSpeciesByFormula(formula string) *ChemSpecies {
	for i := range ChemSpeciesTable {
		if ChemSpeciesTable[i].Formula == formula {
			return &ChemSpeciesTable[i]
		}
	}
	return nil
}

// formulaFromName looks up the formula for a species by display name,
// used when compounds are added dynamically from the creator tab.
func formulaFromName(name string) string {
	for i := range ChemSpeciesTable {
		if ChemSpeciesTable[i].Name == name {
			return ChemSpeciesTable[i].Formula
		}
	}
	return ""
}

var _ = strings.TrimSpace // keep strings imported for future helpers

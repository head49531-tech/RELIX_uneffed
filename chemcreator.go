package main

import (
"fmt"

"fyne.io/fyne/v2"
"fyne.io/fyne/v2/container"
"fyne.io/fyne/v2/widget"
)

func findChemSpecies(formula string) *ChemSpecies {
for i := range ChemSpeciesTable {
if ChemSpeciesTable[i].Formula == formula {
return &ChemSpeciesTable[i]
}
}
return nil
}

func weightedPH(a, b float64) float64 {
if a <= 0 && b <= 0 {
return 7.0
}
if a <= 0 {
return b
}
if b <= 0 {
return a
}
return (a + b) / 2
}

func reactSpecies(a, b ChemSpecies) ChemSpecies {
// acid + base → neutral salt + water
if (a.Class == "acid" && b.Class == "base") || (a.Class == "base" && b.Class == "acid") {
return ChemSpecies{
Category: "Generated", Class: "salt",
Name:    a.Name + " + " + b.Name + " (neutralized)",
Formula: "SALT_" + sanitizeID(a.Formula) + "_" + sanitizeID(b.Formula),
pH:      7.0, Description: "Acid-base neutralization; salt + water",
}
}
// acid + metal → metal salt + H2
if (a.Class == "acid" && b.Class == "metal") || (a.Class == "metal" && b.Class == "acid") {
return ChemSpecies{
Category: "Generated", Class: "salt",
Name:    "Metal salt + H2 gas",
Formula: "MSALT_" + sanitizeID(a.Formula) + "_" + sanitizeID(b.Formula),
Toxicity: 0.2, pH: 3.5, Description: "Acid-metal; releases H2",
}
}
// acid + carbonate → salt + CO2 + H2O
if (a.Class == "acid" && b.Class == "carbonate") || (a.Class == "carbonate" && b.Class == "acid") {
return ChemSpecies{
Category: "Generated", Class: "neutral",
Name:    "CO2 + H2O + salt",
Formula: "CARSALT_" + sanitizeID(a.Formula) + "_" + sanitizeID(b.Formula),
Toxicity: 0.05, pH: 7.0, Description: "Acid-carbonate; releases CO2",
}
}
// oxidizer + organic → partial oxidation
if (a.Class == "oxidizer" && (b.Class == "organic" || b.Class == "nutrient")) ||
(b.Class == "oxidizer" && (a.Class == "organic" || a.Class == "nutrient")) {
return ChemSpecies{
Category: "Generated", Class: "neutral",
Name:    "Oxidation products",
Formula: "OX_" + sanitizeID(a.Formula) + "_" + sanitizeID(b.Formula),
Toxicity: 0.3, Oxidant: 0.2, pH: 5.0, Description: "Partial oxidation",
}
}
// radioactive + anything → mixture retains radiation
if a.Class == "radioactive" || b.Class == "radioactive" {
rad := a.Radiation + b.Radiation
return ChemSpecies{
Category: "Generated", Class: "radioactive",
Name:    "Radioactive mixture",
Formula: "RMIX_" + sanitizeID(a.Formula) + "_" + sanitizeID(b.Formula),
Toxicity: 0.5, Radiation: rad, pH: 7.0, Description: "Retains radioactivity",
}
}
// generic mixture
return ChemSpecies{
Category: "Generated", Class: "neutral",
Name:    a.Name + " / " + b.Name + " mixture",
Formula: "MIX_" + sanitizeID(a.Formula) + "_" + sanitizeID(b.Formula),
Toxicity: (a.Toxicity + b.Toxicity) / 2,
Nutrient: (a.Nutrient + b.Nutrient) / 2,
Oxidant:  (a.Oxidant + b.Oxidant) / 2,
Radiation:(a.Radiation + b.Radiation) / 2,
pH:       weightedPH(a.pH, b.pH),
Description: "Generic mixture",
}
}

func (u *UI) appendChemSpecies(sp ChemSpecies) {
	if _, ok := u.sim.Chem.Concentrations[sp.Formula]; ok {
		return
	}
	ChemSpeciesTable = append(ChemSpeciesTable, sp)
	u.sim.Chem.Concentrations[sp.Formula] = 0
	if u.chemSearch != nil {
		u.rebuildChemBox(u.chemSearch.Text)
	} else {
		u.rebuildChemBox("")
	}
}

func (u *UI) buildChemCreatorTab() fyne.CanvasObject {
formulas := make([]string, 0, len(ChemSpeciesTable))
for _, sp := range ChemSpeciesTable {
formulas = append(formulas, sp.Formula)
}
u.chemA = widget.NewSelect(formulas, nil)
u.chemB = widget.NewSelect(formulas, nil)
if len(formulas) > 0 {
u.chemA.SetSelected(formulas[0])
}
if len(formulas) > 1 {
u.chemB.SetSelected(formulas[1])
}

u.chemPreview = widget.NewLabelWithStyle("Select two reagents and click React.",
fyne.TextAlignLeading, fyne.TextStyle{Monospace: true})
u.chemPreview.Wrapping = fyne.TextWrapWord

reactBtn := widget.NewButton("React", func() {
a := findChemSpecies(u.chemA.Selected)
b := findChemSpecies(u.chemB.Selected)
if a == nil || b == nil {
u.chemPreview.SetText("Pick two valid reagents.")
return
}
product := reactSpeciesBetter(*a, *b)
u.pendingProduct = &product
kind := "compound"
if product.IsElement {
	kind = "element"
}
massStr := "—"
if product.MolarMass > 0 {
	massStr = fmt.Sprintf("%.2f g/mol", product.MolarMass)
}
u.chemPreview.SetText(fmt.Sprintf(
"Product: %s\nFormula: %s\nKind: %s\nMolar mass: %s\npH: %.2f\nToxicity: %.2f\nNutrient: %.2f\nOxidant: %.2f\nRadiation: %.2f\n\n%s",
product.Name, product.Formula, kind, massStr, product.pH,
product.Toxicity, product.Nutrient, product.Oxidant, product.Radiation,
product.Description))
})

addBtn := widget.NewButton("Add to Chemex", func() {
if u.pendingProduct == nil {
u.chemPreview.SetText("No product yet — click React first.")
return
}
u.appendChemSpecies(*u.pendingProduct)
u.chemPreview.SetText("Added " + u.pendingProduct.Formula + " to Chemex tab.")
u.pendingProduct = nil
})

top := container.NewVBox(
widget.NewLabelWithStyle("Chem Creator", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
container.NewHBox(widget.NewLabel("A:"), u.chemA),
container.NewHBox(widget.NewLabel("B:"), u.chemB),
container.NewHBox(reactBtn, addBtn),
widget.NewSeparator(),
)
return container.NewBorder(top, nil, nil, nil, u.chemPreview)
}

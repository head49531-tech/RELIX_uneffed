package main

import (
	"fmt"
	"sort"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

const presetClassAll = "All behaviors"
const presetCatAll = "All categories"

var presetClassOptions = []string{
	presetClassAll,
	"osc", "tog", "stable", "growing", "chaotic", "extinct",
}

func (u *UI) presetCategories() []string {
	set := map[string]bool{}
	for _, p := range u.presetsCache {
		if p.Category != "" {
			set[p.Category] = true
		}
	}
	out := []string{presetCatAll}
	rest := make([]string, 0, len(set))
	for c := range set {
		rest = append(rest, c)
	}
	sort.Strings(rest)
	return append(out, rest...)
}

// applyPresetFilters rebuilds presetsFiltered from presetsCache using the
// current class/category/text filters and sort mode, then refreshes the list.
func (u *UI) applyPresetFilters() {
	if u.presetsCache == nil {
		u.presetsFiltered = nil
	} else {
		q := strings.ToLower(strings.TrimSpace(u.presetFilterText))
		var out []*DNAPreset
		for _, p := range u.presetsCache {
			if u.presetFilterClass != "" && u.presetFilterClass != presetClassAll {
				if u.presetClassKey(p.Name) != u.presetFilterClass {
					continue
				}
			}
			if u.presetFilterCat != "" && u.presetFilterCat != presetCatAll {
				if p.Category != u.presetFilterCat {
					continue
				}
			}
			if q != "" {
				if !strings.Contains(strings.ToLower(p.Name), q) &&
					!strings.Contains(strings.ToLower(p.Category), q) &&
					!strings.Contains(strings.ToLower(p.Description), q) {
					continue
				}
			}
			out = append(out, p)
		}
		switch u.presetSortMode {
		case "class":
			sort.SliceStable(out, func(i, j int) bool {
				ci := u.presetClassKey(out[i].Name)
				cj := u.presetClassKey(out[j].Name)
				if ci != cj {
					return ci < cj
				}
				return out[i].Name < out[j].Name
			})
		case "period":
			sort.SliceStable(out, func(i, j int) bool {
				pi := u.presetPeriod(out[i].Name)
				pj := u.presetPeriod(out[j].Name)
				if pi == 0 && pj == 0 {
					return out[i].Name < out[j].Name
				}
				if pi == 0 {
					return false
				}
				if pj == 0 {
					return true
				}
				return pi < pj
			})
		case "category":
			sort.SliceStable(out, func(i, j int) bool {
				if out[i].Category != out[j].Category {
					return out[i].Category < out[j].Category
				}
				return out[i].Name < out[j].Name
			})
		default:
			sort.SliceStable(out, func(i, j int) bool {
				return out[i].Name < out[j].Name
			})
		}
		u.presetsFiltered = out
	}
	u.selectedPresetIdx = -1
	if u.presetList != nil {
		u.presetList.UnselectAll()
		u.presetList.Refresh()
	}
	u.refreshPresetPreview()
	if u.presetCatSelect != nil {
		newCats := u.presetCategories()
		if !stringSliceEqual(u.presetCatSelect.Options, newCats) {
			u.presetCatSelect.SetOptions(newCats)
		}
	}
}

func (u *UI) presetClassKey(name string) string {
	if u.behaviorCache != nil {
		if rep, ok := u.behaviorCache[name]; ok && rep != nil {
			return string(rep.Class)
		}
	}
	return ""
}

func (u *UI) presetPeriod(name string) float64 {
	if u.behaviorCache != nil {
		if rep, ok := u.behaviorCache[name]; ok && rep != nil {
			return rep.Period
		}
	}
	return 0
}

// ==================== Compare ====================

func (u *UI) showPresetCompare() {
	a := u.presetPinA
	b := u.presetPinB
	if a == nil || b == nil {
		dialog.ShowInformation("Compare", "Pin two presets (A and B) first.", u.window)
		return
	}
	var sb strings.Builder
	fmt.Fprintf(&sb, "%-14s  %-28s  %-28s\n", "", truncateName(a.Name, 28), truncateName(b.Name, 28))
	fmt.Fprintf(&sb, "%s\n", strings.Repeat("-", 74))
	sb.WriteString(fmt.Sprintf("%-14s  %-28s  %-28s\n", "Category", a.Category, b.Category))
	sb.WriteString(fmt.Sprintf("%-14s  %-28s  %-28s\n", "Class", u.compareClass(a.Name), u.compareClass(b.Name)))
	sb.WriteString(fmt.Sprintf("%-14s  %-28s  %-28s\n", "Period", u.comparePeriod(a.Name), u.comparePeriod(b.Name)))
	sb.WriteString(fmt.Sprintf("%-14s  %-28s  %-28s\n", "Alive", u.compareAlive(a.Name), u.compareAlive(b.Name)))
	sb.WriteString(fmt.Sprintf("%-14s  %-28s  %-28s\n", "PopCV", u.comparePopCV(a.Name), u.comparePopCV(b.Name)))
	sb.WriteString(fmt.Sprintf("%-14s  %-28d  %-28d\n", "Genes", len(a.Genes), len(b.Genes)))
	sb.WriteString(fmt.Sprintf("%-14s  %-28d  %-28d\n", "DNA bp", len(a.DNA), len(b.DNA)))

	sb.WriteString("\nGenes:\n")
	maxGenes := len(a.Genes)
	if len(b.Genes) > maxGenes {
		maxGenes = len(b.Genes)
	}
	for i := 0; i < maxGenes; i++ {
		var ca, cb string
		if i < len(a.Genes) {
			ca = fmt.Sprintf("%s→%s", a.Genes[i].Name, a.Genes[i].Product)
		}
		if i < len(b.Genes) {
			cb = fmt.Sprintf("%s→%s", b.Genes[i].Name, b.Genes[i].Product)
		}
		sb.WriteString(fmt.Sprintf("%-14s  %-28s  %-28s\n", "", ca, cb))
	}

	if a.Description != "" || b.Description != "" {
		sb.WriteString("\nDescriptions:\n")
		fmt.Fprintf(&sb, "  A: %s\n", a.Description)
		fmt.Fprintf(&sb, "  B: %s\n", b.Description)
	}

	ta := widget.NewMultiLineEntry()
	ta.Wrapping = fyne.TextWrapOff
	ta.TextStyle = fyne.TextStyle{Monospace: true}
	ta.SetText(sb.String())
	d := dialog.NewCustom("Compare Presets", "Close", container.NewScroll(ta), u.window)
	d.Resize(fyne.NewSize(880, 600))
	d.Show()
}

func (u *UI) compareClass(name string) string {
	if u.behaviorCache != nil {
		if rep, ok := u.behaviorCache[name]; ok && rep != nil {
			return string(rep.Class) + " " + rep.ShortLabel()
		}
	}
	return "—"
}

func (u *UI) comparePeriod(name string) string {
	if u.behaviorCache != nil {
		if rep, ok := u.behaviorCache[name]; ok && rep != nil && rep.Period > 0 {
			return fmt.Sprintf("%.0f t", rep.Period)
		}
	}
	return "—"
}

func (u *UI) compareAlive(name string) string {
	if u.behaviorCache != nil {
		if rep, ok := u.behaviorCache[name]; ok && rep != nil {
			return fmt.Sprintf("%.1f", rep.Alive)
		}
	}
	return "—"
}

func (u *UI) comparePopCV(name string) string {
	if u.behaviorCache != nil {
		if rep, ok := u.behaviorCache[name]; ok && rep != nil {
			return fmt.Sprintf("%.2f", rep.PopCV)
		}
	}
	return "—"
}

// ==================== Filter bar ====================

func (u *UI) buildPresetFilterBar() fyne.CanvasObject {
	u.presetClassSelect = widget.NewSelect(presetClassOptions, func(s string) {
		if s != presetClassAll && len(u.behaviorCache) == 0 {
			dialog.ShowInformation("Behavior filter",
				"Run 'Screen behaviors' first to enable behavior-based filtering.",
				u.window)
			u.presetClassSelect.SetSelected(presetClassAll)
			return
		}
		if s == presetClassAll {
			u.presetFilterClass = ""
		} else {
			u.presetFilterClass = s
		}
		u.applyPresetFilters()
	})
	u.presetClassSelect.SetSelected(presetClassAll)

	u.presetCatSelect = widget.NewSelect(u.presetCategories(), func(s string) {
		if s == presetCatAll {
			u.presetFilterCat = ""
		} else {
			u.presetFilterCat = s
		}
		u.applyPresetFilters()
	})
	u.presetCatSelect.SetSelected(presetCatAll)

	u.presetSortSelect = widget.NewSelect([]string{"Name", "Category", "Class", "Period"}, func(s string) {
		if (s == "Class" || s == "Period") && len(u.behaviorCache) == 0 {
			dialog.ShowInformation("Sort",
				"Run 'Screen behaviors' first to enable "+s+"-based sorting.",
				u.window)
			u.presetSortSelect.SetSelected("Name")
			return
		}
		switch s {
		case "Category":
			u.presetSortMode = "category"
		case "Class":
			u.presetSortMode = "class"
		case "Period":
			u.presetSortMode = "period"
		default:
			u.presetSortMode = "name"
		}
		u.applyPresetFilters()
	})
	u.presetSortSelect.SetSelected("Name")

	u.presetSearch = widget.NewEntry()
	u.presetSearch.SetPlaceHolder("Search name / category / description…")
	u.presetSearch.OnChanged = func(s string) {
		u.presetFilterText = s
		u.applyPresetFilters()
	}

	return container.NewVBox(
		container.NewHBox(
			widget.NewLabel("Behavior:"), u.presetClassSelect,
			widget.NewLabel("Category:"), u.presetCatSelect,
			widget.NewLabel("Sort:"), u.presetSortSelect,
		),
		u.presetSearch,
	)
}

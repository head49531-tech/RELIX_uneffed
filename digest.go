package main

import (
	"fmt"
	"image"
	"image/color"
	"sort"
	"strings"
)

// Enzyme is a restriction endonuclease with its recognition site.
// Site uses IUPAC codes: A/T/G/C literal; R=A|G, Y=C|T, S=G|C, W=A|T,
// K=G|T, M=A|C, B=C|G|T, D=A|G|T, H=A|C|T, V=A|C|G, N=any.
type Enzyme struct {
	Name string
	Site string
	// Cut is offset from start of site where the phosphodiester bond is broken.
	// For palindromic enzymes, cut is usually on both strands; here we just
	// track the 5' cut position on the top strand.
	Cut int
}

var EnzymeTable = []Enzyme{
	{"EcoRI",   "GAATTC",   1},
	{"BamHI",   "GGATCC",   1},
	{"HindIII", "AAGCTT",   1},
	{"NotI",    "GCGGCCGC", 2},
	{"XhoI",    "CTCGAG",   1},
	{"SalI",    "GTCGAC",   1},
	{"PstI",    "CTGCAG",   5},
	{"KpnI",    "GGTACC",   5},
	{"SacI",    "GAGCTC",   5},
	{"SmaI",    "CCCGGG",   3},
	{"XbaI",    "TCTAGA",   1},
	{"SpeI",    "ACTAGT",   1},
	{"NcoI",    "CCATGG",   1},
	{"NdeI",    "CATATG",   2},
	{"BglII",   "AGATCT",   1},
	{"PvuII",   "CAGCTG",   3},
	{"SphI",    "GCATGC",   5},
	{"TaqI",    "TCGA",     1},
	{"AluI",    "AGCT",     2},
	{"HaeIII",  "GGCC",     2},
}

// iupacMatch reports whether a DNA base b matches IUPAC code c.
func iupacMatch(b byte, c byte) bool {
	switch c {
	case 'A', 'C', 'G', 'T':
		return b == c
	case 'R':
		return b == 'A' || b == 'G'
	case 'Y':
		return b == 'C' || b == 'T'
	case 'S':
		return b == 'G' || b == 'C'
	case 'W':
		return b == 'A' || b == 'T'
	case 'K':
		return b == 'G' || b == 'T'
	case 'M':
		return b == 'A' || b == 'C'
	case 'B':
		return b == 'C' || b == 'G' || b == 'T'
	case 'D':
		return b == 'A' || b == 'G' || b == 'T'
	case 'H':
		return b == 'A' || b == 'C' || b == 'T'
	case 'V':
		return b == 'A' || b == 'C' || b == 'G'
	case 'N':
		return b == 'A' || b == 'C' || b == 'G' || b == 'T'
	}
	return false
}

// findSites returns the positions where enzyme e cuts the DNA.
// Circular mode allows matches across the origin.
func findSites(dna []byte, e Enzyme, circular bool) []int {
	var cuts []int
	n := len(dna)
	m := len(e.Site)
	if m == 0 || n == 0 {
		return cuts
	}
	searchLen := n - m + 1
	if circular {
		searchLen = n
	}
	for i := 0; i < searchLen; i++ {
		match := true
		for j := 0; j < m; j++ {
			b := dna[(i+j)%n]
			if !iupacMatch(b, e.Site[j]) {
				match = false
				break
			}
		}
		if match {
			cutPos := (i + e.Cut) % n
			cuts = append(cuts, cutPos)
		}
	}
	return cuts
}

// DigestResult holds per-enzyme cut sites and the resulting fragments.
type DigestResult struct {
	Enzyme    string
	Site      string
	Cuts      []int
	Fragments []int
}

// Digest runs one enzyme on the DNA, computes fragment sizes.
func Digest(dna []byte, e Enzyme, circular bool) DigestResult {
	cuts := findSites(dna, e, circular)
	sort.Ints(cuts)
	var frags []int
	if len(cuts) == 0 {
		frags = []int{len(dna)}
	} else if circular {
		for i := 0; i < len(cuts); i++ {
			next := cuts[(i+1)%len(cuts)]
			if i == len(cuts)-1 {
				frags = append(frags, (len(dna)-cuts[i])+cuts[0])
			} else {
				frags = append(frags, next-cuts[i])
			}
		}
	} else {
		prev := 0
		for _, c := range cuts {
			frags = append(frags, c-prev)
			prev = c
		}
		frags = append(frags, len(dna)-prev)
	}
	return DigestResult{Enzyme: e.Name, Site: e.Site, Cuts: cuts, Fragments: frags}
}

// DigestAll runs every enzyme in the table.
func DigestAll(dna []byte, circular bool) []DigestResult {
	out := make([]DigestResult, 0, len(EnzymeTable))
	for _, e := range EnzymeTable {
		out = append(out, Digest(dna, e, circular))
	}
	return out
}

// FormatDigestReport returns a text summary of one digest.
func FormatDigestReport(r DigestResult, dnaLen int, circular bool) string {
	var sb strings.Builder
	mode := "linear"
	if circular {
		mode = "circular"
	}
	fmt.Fprintf(&sb, "%s  site=%s  (%s, %d bp)\n", r.Enzyme, r.Site, mode, dnaLen)
	if len(r.Cuts) == 0 {
		fmt.Fprintf(&sb, "  no cut sites\n\n")
		return sb.String()
	}
	fmt.Fprintf(&sb, "  cuts: %v\n", r.Cuts)
	fmt.Fprintf(&sb, "  fragments: %d → sizes (bp): %v\n\n", len(r.Fragments), r.Fragments)
	return sb.String()
}

// renderGel draws an ASCII-style agarose gel on a canvas.
// Each row = one digest result. Fragments shown as horizontal bars,
// positioned by log10(size) so 100bp and 10kb are visually distinguishable.
func renderGel(results []DigestResult, w, h int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	bg := color.RGBA{14, 18, 22, 255}
	for i := 0; i < w*h*4; i += 4 {
		img.Pix[i] = bg.R
		img.Pix[i+1] = bg.G
		img.Pix[i+2] = bg.B
		img.Pix[i+3] = bg.A
	}
	if len(results) == 0 {
		return img
	}
	// Find global max fragment size for normalization
	maxFrag := 1
	for _, r := range results {
		for _, f := range r.Fragments {
			if f > maxFrag {
				maxFrag = f
			}
		}
	}
	leftMargin := 60
	rightMargin := 60
	topMargin := 10
	bottomMargin := 10
	usableW := w - leftMargin - rightMargin
	usableH := h - topMargin - bottomMargin
	if usableW < 20 || usableH < 20 {
		return img
	}
	laneH := usableH / len(results)
	if laneH < 6 {
		laneH = 6
	}

	// lane bars
	for i, r := range results {
		y := topMargin + i*laneH + laneH/2
		// lane background
		for x := leftMargin - 2; x < leftMargin+usableW+2; x++ {
			if x >= 0 && x < w && y >= 0 && y < h {
				img.Set(x, y, color.RGBA{30, 36, 44, 255})
			}
		}
		// fragments as bright bands
		for _, f := range r.Fragments {
			// logarithmic position: log10(f) / log10(maxFrag) maps [1, maxFrag] to [0,1]
			frac := float64(f) / float64(maxFrag)
			_ = frac
			// log-scale width: bigger fragments render wider
			width := int(float64(usableW) * (1.0 - float64(logApprox(f))/float64(logApprox(maxFrag)+1)))
			if width < 4 {
				width = 4
			}
			if width > usableW {
				width = usableW
			}
			x0 := leftMargin + (usableW-width)/2
			var c color.Color = color.RGBA{64, 196, 180, 255}
			if f < 500 {
				c = color.RGBA{90, 210, 130, 255}
			} else if f < 5000 {
				c = color.RGBA{64, 196, 180, 255}
			} else {
				c = color.RGBA{180, 140, 255, 255}
			}
			for dx := 0; dx < width; dx++ {
				px := x0 + dx
				for dy := -1; dy <= 1; dy++ {
					py := y + dy
					if px >= 0 && px < w && py >= 0 && py < h {
						img.Set(px, py, c)
					}
				}
			}
		}
	}
	return img
}

// logApprox gives a positive log-like value for scaling.
func logApprox(n int) float64 {
	if n <= 1 {
		return 1
	}
	// natural log approximation via iteration
	x := float64(n)
	// rough ln via repeated sqrt
	result := 0.0
	for x > 2 {
		x = sqrtApprox(x)
		result += 0.693
	}
	return result + (x - 1)
}

func sqrtApprox(x float64) float64 {
	if x <= 0 {
		return 0
	}
	guess := x / 2
	for i := 0; i < 20; i++ {
		guess = (guess + x/guess) / 2
	}
	return guess
}

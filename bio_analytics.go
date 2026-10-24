package main

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"sort"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
)

// ==================== Lineage ====================

type LineageNode struct {
	Cell     *Cell
	Children []*LineageNode
	Depth    int
	X, Y     float64
}

func (s *Sim) BuildLineage(maxNodes int) (*LineageNode, int) {
	byParent := map[int][]*Cell{}
	var roots []*Cell
	for _, c := range s.Cells {
		if c.ParentID == 0 {
			roots = append(roots, c)
		} else {
			byParent[c.ParentID] = append(byParent[c.ParentID], c)
		}
	}
	count := 0
	var build func(c *Cell, depth int) *LineageNode
	build = func(c *Cell, depth int) *LineageNode {
		count++
		n := &LineageNode{Cell: c, Depth: depth}
		if count < maxNodes {
			for _, child := range byParent[c.ID] {
				n.Children = append(n.Children, build(child, depth+1))
			}
		}
		return n
	}
	var root *LineageNode
	if len(roots) == 1 {
		root = build(roots[0], 0)
	} else {
		root = &LineageNode{Depth: -1}
		for _, r := range roots {
			root.Children = append(root.Children, build(r, 0))
		}
	}
	leaf := 0
	maxDepth := 0
	var assign func(n *LineageNode)
	assign = func(n *LineageNode) {
		if len(n.Children) == 0 {
			n.X = float64(leaf)
			leaf++
		} else {
			for _, c := range n.Children {
				assign(c)
			}
			n.X = (n.Children[0].X + n.Children[len(n.Children)-1].X) / 2
		}
		n.Y = float64(n.Depth)
		if n.Depth > maxDepth {
			maxDepth = n.Depth
		}
	}
	assign(root)
	return root, maxDepth
}

func drawLinePix(img *image.RGBA, x0, y0, x1, y1 int, c color.Color) {
	dx, dy := x1-x0, y1-y0
	if dx < 0 { dx = -dx }
	if dy < 0 { dy = -dy }
	dy = -dy
	sx, sy := -1, -1
	if x0 < x1 { sx = 1 }
	if y0 < y1 { sy = 1 }
	err := dx + dy
	w, h := img.Bounds().Dx(), img.Bounds().Dy()
	for {
		if x0 >= 0 && x0 < w && y0 >= 0 && y0 < h {
			img.Set(x0, y0, c)
		}
		if x0 == x1 && y0 == y1 {
			break
		}
		e2 := 2 * err
		if e2 >= dy { err += dy; x0 += sx }
		if e2 <= dx { err += dx; y0 += sy }
	}
}

func renderLineage(root *LineageNode, maxDepth, w, h int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	bg := color.RGBA{13, 13, 16, 255}
	for i := 0; i < w*h*4; i += 4 {
		img.Pix[i] = bg.R
		img.Pix[i+1] = bg.G
		img.Pix[i+2] = bg.B
		img.Pix[i+3] = bg.A
	}
	maxX := 0.0
	var walk func(n *LineageNode)
	walk = func(n *LineageNode) {
		if n.X > maxX { maxX = n.X }
		for _, c := range n.Children { walk(c) }
	}
	walk(root)
	if maxX == 0 { maxX = 1 }
	if maxDepth == 0 { maxDepth = 1 }
	pxPerX := float64(w-20) / maxX
	pxPerY := float64(h-20) / float64(maxDepth+1)
	toPx := func(x, y float64) (int, int) {
		return 10 + int(x*pxPerX), 10 + int(y*pxPerY)
	}
	var drawEdges func(n *LineageNode)
	drawEdges = func(n *LineageNode) {
		x0, y0 := toPx(n.X, n.Y)
		for _, c := range n.Children {
			x1, y1 := toPx(c.X, c.Y)
			ec := color.RGBA{80, 80, 100, 255}
			if c.Cell != nil && !c.Cell.Alive {
				ec = color.RGBA{50, 30, 30, 255}
			}
			drawLinePix(img, x0, y0, x1, y1, ec)
			drawEdges(c)
		}
	}
	drawEdges(root)
	walk = func(n *LineageNode) {
		if n.Cell != nil {
			x, y := toPx(n.X, n.Y)
			var c color.Color
			if n.Cell.Alive {
				c = hslToRGB(float64((n.Cell.Generation*37)%360), 0.65, 0.55)
			} else {
				c = color.RGBA{60, 60, 70, 255}
			}
			for dx := -2; dx <= 2; dx++ {
				for dy := -2; dy <= 2; dy++ {
					if dx*dx+dy*dy <= 4 {
						px, py := x+dx, y+dy
						if px >= 0 && px < w && py >= 0 && py < h {
							img.Set(px, py, c)
						}
					}
				}
			}
		}
		for _, ch := range n.Children {
			walk(ch)
		}
	}
	walk(root)
	return img
}

// ==================== Phase portrait ====================

func (s *Sim) speciesValue(sp SpeciesPoint, name string) float64 {
	switch name {
	case "Alive":
		return float64(sp.Alive)
	case "TotalATP":
		return sp.TotalATP
	case "EnvGlucose":
		return sp.EnvGlucose
	}
	if strings.HasSuffix(name, "_mRNA") {
		return sp.MRNA[strings.TrimSuffix(name, "_mRNA")]
	}
	return sp.Proteins[name]
}

func (s *Sim) PhaseData(xname, yname string) [][2]float64 {
	out := make([][2]float64, 0, len(s.SpeciesHistory))
	for _, sp := range s.SpeciesHistory {
		out = append(out, [2]float64{s.speciesValue(sp, xname), s.speciesValue(sp, yname)})
	}
	return out
}

func renderPhase(pts [][2]float64, w, h int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	bg := color.RGBA{13, 13, 16, 255}
	for i := 0; i < w*h*4; i += 4 {
		img.Pix[i] = bg.R
		img.Pix[i+1] = bg.G
		img.Pix[i+2] = bg.B
		img.Pix[i+3] = bg.A
	}
	if len(pts) < 2 { return img }
	maxX, minX := pts[0][0], pts[0][0]
	maxY, minY := pts[0][1], pts[0][1]
	for _, p := range pts {
		if p[0] > maxX { maxX = p[0] }
		if p[0] < minX { minX = p[0] }
		if p[1] > maxY { maxY = p[1] }
		if p[1] < minY { minY = p[1] }
	}
	if maxX == minX { maxX = minX + 1 }
	if maxY == minY { maxY = minY + 1 }
	margin := 20
	toPx := func(x, y float64) (int, int) {
		return margin + int((x-minX)/(maxX-minX)*float64(w-2*margin)),
			h - margin - int((y-minY)/(maxY-minY)*float64(h-2*margin))
	}
	axis := color.RGBA{60, 60, 80, 255}
	for i := margin; i < h-margin; i++ { img.Set(margin, i, axis) }
	for i := margin; i < w-margin; i++ { img.Set(i, h-margin, axis) }
	for i := 1; i < len(pts); i++ {
		x0, y0 := toPx(pts[i-1][0], pts[i-1][1])
		x1, y1 := toPx(pts[i][0], pts[i][1])
		age := float64(i) / float64(len(pts))
		c := color.RGBA{uint8(80 + age*175), uint8(120 + age*80), 200, 255}
		drawLinePix(img, x0, y0, x1, y1, c)
	}
	x0, y0 := toPx(pts[0][0], pts[0][1])
	xN, yN := toPx(pts[len(pts)-1][0], pts[len(pts)-1][1])
	for dx := -3; dx <= 3; dx++ {
		for dy := -3; dy <= 3; dy++ {
			if dx*dx+dy*dy <= 9 {
				if x0+dx >= 0 && x0+dx < w && y0+dy >= 0 && y0+dy < h {
					img.Set(x0+dx, y0+dy, color.RGBA{80, 220, 120, 255})
				}
				if xN+dx >= 0 && xN+dx < w && yN+dy >= 0 && yN+dy < h {
					img.Set(xN+dx, yN+dy, color.RGBA{240, 100, 100, 255})
				}
			}
		}
	}
	return img
}

// ==================== Dose response ====================

type DosePoint struct {
	X, Y float64
}

func fitHill(pts []DosePoint) (K, n, ymin, ymax, ssr float64) {
	if len(pts) < 3 { return 1, 1, 0, 0, 1e18 }
	ys := make([]float64, len(pts))
	for i, p := range pts { ys[i] = p.Y }
	sort.Float64s(ys)
	ymin, ymax = ys[0], ys[len(ys)-1]
	if ymax == ymin { ymax = ymin + 1 }
	bestSSR := 1e18
	K, n = 1, 1
	for ki := 0; ki < 40; ki++ {
		Kc := 0.1 * math.Pow(100, float64(ki)/39.0)
		for ni := 0; ni < 30; ni++ {
			nc := 0.5 + float64(ni)*4.5/29.0
			var s float64
			for _, p := range pts {
				pred := ymin + (ymax-ymin)*math.Pow(p.X, nc)/(math.Pow(Kc, nc)+math.Pow(p.X, nc))
				d := p.Y - pred
				s += d * d
			}
			if s < bestSSR { bestSSR, K, n = s, Kc, nc }
		}
	}
	return K, n, ymin, ymax, bestSSR
}

func renderDose(pts []DosePoint, K, n, ymin, ymax float64, w, h int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	bg := color.RGBA{13, 13, 16, 255}
	for i := 0; i < w*h*4; i += 4 {
		img.Pix[i] = bg.R
		img.Pix[i+1] = bg.G
		img.Pix[i+2] = bg.B
		img.Pix[i+3] = bg.A
	}
	if len(pts) < 2 { return img }
	maxX, minX := pts[0].X, pts[0].X
	for _, p := range pts {
		if p.X > maxX { maxX = p.X }
		if p.X < minX { minX = p.X }
	}
	if maxX == minX { maxX = minX + 1 }
	if ymax == ymin { ymax = ymin + 1 }
	margin := 20
	toPx := func(x, y float64) (int, int) {
		return margin + int((x-minX)/(maxX-minX)*float64(w-2*margin)),
			h - margin - int((y-ymin)/(ymax-ymin)*float64(h-2*margin))
	}
	axis := color.RGBA{60, 60, 80, 255}
	for i := margin; i < h-margin; i++ { img.Set(margin, i, axis) }
	for i := margin; i < w-margin; i++ { img.Set(i, h-margin, axis) }
	if n > 0 && K > 0 {
		for i := 0; i < 100; i++ {
			x0 := minX + (maxX-minX)*float64(i)/100.0
			x1 := minX + (maxX-minX)*float64(i+1)/100.0
			y0 := ymin + (ymax-ymin)*math.Pow(x0, n)/(math.Pow(K, n)+math.Pow(x0, n))
			y1 := ymin + (ymax-ymin)*math.Pow(x1, n)/(math.Pow(K, n)+math.Pow(x1, n))
			px0, py0 := toPx(x0, y0)
			px1, py1 := toPx(x1, y1)
			drawLinePix(img, px0, py0, px1, py1, color.RGBA{180, 140, 255, 255})
		}
	}
	for _, p := range pts {
		px, py := toPx(p.X, p.Y)
		for dx := -3; dx <= 3; dx++ {
			for dy := -3; dy <= 3; dy++ {
				if dx*dx+dy*dy <= 9 && px+dx >= 0 && px+dx < w && py+dy >= 0 && py+dy < h {
					img.Set(px+dx, py+dy, color.RGBA{255, 220, 80, 255})
				}
			}
		}
	}
	return img
}

func (u *UI) speciesOptions() []string {
	opts := []string{"Alive", "TotalATP", "EnvGlucose"}
	if len(u.sim.SpeciesHistory) == 0 {
		return opts
	}
	last := u.sim.SpeciesHistory[len(u.sim.SpeciesHistory)-1]
	pk := make([]string, 0, len(last.Proteins))
	for k := range last.Proteins { pk = append(pk, k) }
	sort.Strings(pk)
	opts = append(opts, pk...)
	mk := make([]string, 0, len(last.MRNA))
	for k := range last.MRNA { mk = append(mk, k+"_mRNA") }
	sort.Strings(mk)
	opts = append(opts, mk...)
	return opts
}

func (u *UI) runDoseResponse() {
	if u.doseParam.Selected == "" || u.doseOut.Selected == "" {
		return
	}
	param := u.doseParam.Selected
	out := u.doseOut.Selected
	dialog.ShowInformation("Sweep running", "11 runs x 150 ticks in background...", u.window)
	go func() {
		base := u.sim.Params
		baseInd := map[string]float64{}
		for k, v := range u.sim.Inducers {
			baseInd[k] = v
		}
		N, ticks := 11, 150
		pts := make([]DosePoint, 0, N)
		for i := 0; i < N; i++ {
			val := float64(i) * 10.0 / float64(N-1)
			p := base
			ind := map[string]float64{"IPTG": baseInd["IPTG"], "aTc": baseInd["aTc"], "Arabinose": baseInd["Arabinose"]}
			switch param {
			case "IPTG", "aTc", "Arabinose":
				ind[param] = val
			case "DivisionATP":
				p.DivisionATP = 2 + val
			case "TranslationK":
				p.TranslationK = 0.5 + val
			case "BaseDeathRate":
				p.BaseDeathRate = val * 0.002
			case "EnvGlucose":
				p.EnvGlucose = val
			}
			sim := NewSim(p, time.Now().UnixNano()+int64(i))
			for k, v := range ind { sim.Inducers[k] = v }
			for t := 0; t < ticks; t++ {
				sim.Step()
			}
			var yv float64
			switch out {
			case "Alive":
				yv = float64(sim.AliveCount())
			case "TotalATP":
				for _, c := range sim.Cells {
					if c.Alive { yv += c.ATP }
				}
			case "EnvGlucose":
				yv = sim.EnvPool
			default:
				if strings.HasSuffix(out, "_mRNA") {
					yv = sim.GlobalMRNALevels()[strings.TrimSuffix(out, "_mRNA")]
				} else {
					yv = sim.GlobalProteinLevels()[out]
				}
			}
			pts = append(pts, DosePoint{X: val, Y: yv})
		}
		K, n, ymin, ymax, _ := fitHill(pts)
		fyne.Do(func() {
			u.dosePoints = pts
			u.doseFitK = K
			u.doseFitN = n
			u.doseYMin = ymin
			u.doseYMax = ymax
			if u.doseFit != nil {
				u.doseFit.SetText(fmt.Sprintf("Hill fit: K = %.3f, n = %.2f", K, n))
			}
			if u.doseCv != nil {
				u.doseCv.Refresh()
			}
		})
	}()
}

func hslToRGB(h, s, l float64) color.RGBA {
	h = h / 360.0
	var r, g, b float64
	if s == 0 {
		r, g, b = l, l, l
	} else {
		var q float64
		if l < 0.5 {
			q = l * (1 + s)
		} else {
			q = l + s - l*s
		}
		p := 2*l - q
		r = hueToRGB(p, q, h+1.0/3.0)
		g = hueToRGB(p, q, h)
		b = hueToRGB(p, q, h-1.0/3.0)
	}
	return color.RGBA{uint8(r * 255), uint8(g * 255), uint8(b * 255), 255}
}

func hueToRGB(p, q, t float64) float64 {
	if t < 0 {
		t += 1
	}
	if t > 1 {
		t -= 1
	}
	if t < 1.0/6.0 {
		return p + (q-p)*6*t
	}
	if t < 1.0/2.0 {
		return q
	}
	if t < 2.0/3.0 {
		return p + (q-p)*(2.0/3.0-t)*6
	}
	return p
}


func renderBifurcation(points []BifurcationPoint, w, h int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	bg := color.RGBA{14, 18, 22, 255}
	for i := 0; i < w*h*4; i += 4 {
		img.Pix[i] = bg.R
		img.Pix[i+1] = bg.G
		img.Pix[i+2] = bg.B
		img.Pix[i+3] = bg.A
	}
	if len(points) < 2 { return img }
	maxX, minX := points[0].ParamValue, points[0].ParamValue
	maxY := 1
	for _, p := range points {
		if p.ParamValue > maxX { maxX = p.ParamValue }
		if p.ParamValue < minX { minX = p.ParamValue }
		if p.Alive > maxY { maxY = p.Alive }
	}
	if maxX == minX { maxX = minX + 1 }
	margin := 20
	toPx := func(x float64, y int) (int, int) {
		return margin + int((x-minX)/(maxX-minX)*float64(w-2*margin)),
			h - margin - int(float64(y)/float64(maxY)*float64(h-2*margin))
	}
	axis := color.RGBA{60, 60, 80, 255}
	for i := margin; i < h-margin; i++ { img.Set(margin, i, axis) }
	for i := margin; i < w-margin; i++ { img.Set(i, h-margin, axis) }
	for i := 1; i < len(points); i++ {
		p0, p1 := points[i-1], points[i]
		x0, y0 := toPx(p0.ParamValue, p0.Alive)
		x1, y1 := toPx(p1.ParamValue, p1.Alive)
		c := color.RGBA{64, 196, 180, 255}
		switch p1.Regime {
		case "extinct": c = color.RGBA{200, 80, 80, 255}
		case "oscillating": c = color.RGBA{240, 190, 70, 255}
		case "stable": c = color.RGBA{90, 210, 130, 255}
		}
		drawLinePix(img, x0, y0, x1, y1, c)
	}
	return img
}

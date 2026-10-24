package main

import (
	"fmt"
	"image"
	"image/color"
	"math"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// geneColorAt returns a stable color for gene index i.
func geneColorAt(i int) color.RGBA {
	hue := float64((i * 55) % 360)
	return hslToRGB(hue, 0.7, 0.55)
}

// drawCircleRing draws a 1-pixel ring at given radius.
func drawCircleRing(img *image.RGBA, cx, cy int, radius float64, c color.Color) {
	steps := int(radius * 6.28 * 2)
	if steps < 200 {
		steps = 200
	}
	for i := 0; i < steps; i++ {
		angle := 2 * math.Pi * float64(i) / float64(steps)
		x := cx + int(radius*math.Sin(angle))
		y := cy - int(radius*math.Cos(angle))
		if x >= 0 && x < img.Bounds().Dx() && y >= 0 && y < img.Bounds().Dy() {
			img.Set(x, y, c)
		}
	}
}

// drawArc draws a thick arc between two angles.
func drawArc(img *image.RGBA, cx, cy int, radius, startAngle, endAngle float64, thickness int, c color.Color) {
	if endAngle < startAngle {
		endAngle += 2 * math.Pi
	}
	span := endAngle - startAngle
	steps := int(span*radius) * 2
	if steps < 30 {
		steps = 30
	}
	w, h := img.Bounds().Dx(), img.Bounds().Dy()
	for i := 0; i <= steps; i++ {
		t := float64(i) / float64(steps)
		angle := startAngle + t*span
		for d := -thickness / 2; d <= thickness/2; d++ {
			r := radius + float64(d)
			x := cx + int(r*math.Sin(angle))
			y := cy - int(r*math.Cos(angle))
			if x >= 0 && x < w && y >= 0 && y < h {
				img.Set(x, y, c)
			}
		}
	}
}

// drawRadialTick draws a short radial line at the given angle.
func drawRadialTick(img *image.RGBA, cx, cy int, radius, angle float64, length int, c color.Color) {
	w, h := img.Bounds().Dx(), img.Bounds().Dy()
	for d := -length; d <= length; d++ {
		r := radius + float64(d)
		x := cx + int(r*math.Sin(angle))
		y := cy - int(r*math.Cos(angle))
		if x >= 0 && x < w && y >= 0 && y < h {
			img.Set(x, y, c)
		}
	}
}

// drawFilledDot draws a small filled circle.
func drawFilledDot(img *image.RGBA, cx, cy, r int, c color.Color) {
	w, h := img.Bounds().Dx(), img.Bounds().Dy()
	for dx := -r; dx <= r; dx++ {
		for dy := -r; dy <= r; dy++ {
			if dx*dx+dy*dy <= r*r {
				x, y := cx+dx, cy+dy
				if x >= 0 && x < w && y >= 0 && y < h {
					img.Set(x, y, c)
				}
			}
		}
	}
}

// renderPlasmidMap draws the circular genome: backbone, gene arcs, restriction sites, origin.
func renderPlasmidMap(g *Genome, w, h int, circular bool) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	bg := color.RGBA{14, 18, 22, 255}
	for i := 0; i < w*h*4; i += 4 {
		img.Pix[i] = bg.R
		img.Pix[i+1] = bg.G
		img.Pix[i+2] = bg.B
		img.Pix[i+3] = bg.A
	}
	if g == nil || len(g.DNA) == 0 {
		return img
	}
	cx, cy := w/2, h/2
	radius := float64(min(w, h))/2 - 40
	if radius < 40 {
		return img
	}

	// backbone ring
	backboneColor := color.RGBA{70, 80, 95, 255}
	if circular {
		drawCircleRing(img, cx, cy, radius, backboneColor)
	} else {
		// linear: half-ring shape (just a full ring for simplicity, but dotted)
		drawCircleRing(img, cx, cy, radius, backboneColor)
	}

	n := len(g.DNA)
	if n == 0 {
		return img
	}

	// gene arcs
	for i, gene := range g.Genes {
		startAngle := 2 * math.Pi * float64(gene.PromoterStart) / float64(n)
		endAngle := 2 * math.Pi * float64(gene.CDSEnd) / float64(n)
		col := geneColorAt(i)
		if gene.Disabled || gene.Knockdown >= 1.0 {
			col = color.RGBA{70, 70, 80, 255}
		}
		drawArc(img, cx, cy, radius, startAngle, endAngle, 8, col)
		// small start marker (promoter)
		drawRadialTick(img, cx, cy, radius, startAngle, 4, color.RGBA{255, 220, 100, 255})
	}

	// origin marker (angle 0, top)
	drawRadialTick(img, cx, cy, radius, 0, 10, color.RGBA{240, 190, 70, 255})
	drawFilledDot(img, cx+int(radius*math.Sin(0)), cy-int(radius*math.Cos(0)), 3, color.RGBA{240, 190, 70, 255})

	// restriction sites from all enzymes, rendered as small dots around outer edge
	for _, e := range EnzymeTable {
		cuts := findSites(g.DNA, e, circular)
		for _, c := range cuts {
			angle := 2 * math.Pi * float64(c) / float64(n)
			x := cx + int((radius+14)*math.Sin(angle))
			y := cy - int((radius+14)*math.Cos(angle))
			drawFilledDot(img, x, y, 2, color.RGBA{100, 140, 200, 255})
		}
	}

	return img
}

// buildPlasmidTab returns the Plasmid tab UI: raster + legend.
func (u *UI) buildPlasmidTab() fyne.CanvasObject {
	u.plasmidCv = canvas.NewRaster(func(w, h int) image.Image {
		c := u.getSelectedCell()
		if c == nil {
			blank := image.NewRGBA(image.Rect(0, 0, w, h))
			for i := 0; i < w*h*4; i += 4 {
				blank.Pix[i+3] = 255
			}
			return blank
		}
		return renderPlasmidMap(c.Genome, w, h, u.digestCircular != nil && u.digestCircular.Checked)
	})
	u.plasmidCv.SetMinSize(fyne.NewSize(500, 500))

	u.plasmidLegend = container.NewVBox(mono("(no cell selected)"))
	legendScroll := container.NewVScroll(u.plasmidLegend)

	header := container.NewHBox(
		widget.NewLabelWithStyle("Genes", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
	)

	left := container.NewBorder(
		widget.NewLabelWithStyle("Plasmid map", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		nil, nil, nil, u.plasmidCv,
	)
	right := container.NewBorder(header, nil, nil, nil, legendScroll)

	split := container.NewHSplit(left, right)
	split.SetOffset(0.7)
	return split
}

// refreshPlasmidLegend rebuilds the gene color legend.
func (u *UI) refreshPlasmidLegend() {
	c := u.getSelectedCell()
	if c == nil || u.plasmidLegend == nil {
		return
	}
	var objs []fyne.CanvasObject
	for i, gene := range c.Genome.Genes {
		rect := canvas.NewRectangle(geneColorAt(i))
		rect.SetMinSize(fyne.NewSize(14, 14))
		status := ""
		if gene.Disabled {
			status = " [disabled]"
		} else if gene.Knockdown >= 0.99 {
			status = " [KO]"
		} else if gene.Knockdown > 0.01 {
			status = fmt.Sprintf(" [KD %d%%]", int(gene.Knockdown*100))
		}
		label := mono(fmt.Sprintf("%s [%d..%d]%s", gene.Name, gene.PromoterStart, gene.CDSEnd, status))
		objs = append(objs, container.NewHBox(rect, label))
	}
	if len(objs) == 0 {
		objs = append(objs, mono("(no genes)"))
	}
	objs = append(objs, widget.NewSeparator())
	objs = append(objs, container.NewHBox(
		canvas.NewRectangle(color.RGBA{240, 190, 70, 255}),
		mono("origin"),
	))
	objs = append(objs, container.NewHBox(
		canvas.NewRectangle(color.RGBA{100, 140, 200, 255}),
		mono("restriction site"),
	))
	u.plasmidLegend.Objects = objs
	u.plasmidLegend.Refresh()
}

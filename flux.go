package main

import (
	"math"
	"math/rand"
	"sort"
)

// ==================== Flux Balance Analysis ====================

type FluxResult struct {
	Fluxes     map[string]float64
	Objective  float64
	Bottleneck string
	Consistent bool
}

// FBA computes steady-state flux through pathway steps given enzyme
// concentrations and substrate supply. Uses iterative Jacobi refinement
// on the mass-balance v = kcat*E*S/(Km+S), capped by upstream flux.
func (s *Sim) RunFBA(c *Cell, maxIter int) FluxResult {
	fluxes := map[string]float64{}
	substrates := map[string]float64{}
	for name, v := range c.Proteins {
		substrates[name] = v
	}
	for k, v := range c.Metabolites {
		substrates[k] = v
	}
	if _, ok := substrates["Glucose"]; !ok {
		substrates["Glucose"] = c.Glucose
	}
	if _, ok := substrates["ATP"]; !ok {
		substrates["ATP"] = c.ATP
	}

	bottleneck := ""
	minFlux := math.Inf(1)

	for iter := 0; iter < maxIter; iter++ {
		changed := false
		for _, step := range c.Genome.Pathways {
			enz := c.Proteins[step.Enzyme]
			sub := substrates[step.Substrate]
			if enz <= 0 || sub <= 0 {
				continue
			}
			km := step.Km
			if km <= 0 {
				km = 5.0
			}
			kcat := step.Kcat
			if kcat <= 0 {
				kcat = 0.5
			}
			flux := kcat * enz * sub / (km + sub)
			key := step.Enzyme + ":" + step.Substrate + ">" + step.Product
			if flux > fluxes[key] {
				fluxes[key] = flux
				changed = true
			}
			if flux < minFlux {
				minFlux = flux
				bottleneck = key
			}
			substrates[step.Product] += flux * 0.5
		}
		if !changed {
			break
		}
	}

	objective := 0.0
	if v, ok := substrates["ATP"]; ok {
		objective = v
	}

	return FluxResult{
		Fluxes:     fluxes,
		Objective:  objective,
		Bottleneck: bottleneck,
		Consistent: len(c.Genome.Pathways) > 0,
	}
}

// ==================== Gillespie SSA ====================

type SSARecord struct {
	Time    float64
	Species string
	Delta   int
}

// GillespieStochastic runs a direct SSA over mRNA transcription events
// for one cell. Returns per-gene trajectories.
func (s *Sim) GillespieStochastic(c *Cell, tMax float64, rng *rand.Rand) map[string][]float64 {
	trajectories := map[string][]float64{}
	state := map[string]float64{}
	for _, gene := range c.Genome.Genes {
		trajectories[gene.Name] = []float64{0}
		state[gene.Name] = 0
	}
	t := 0.0
	for t < tMax {
		// total propensity
		totalProp := 0.0
		rates := map[string]float64{}
		for _, gene := range c.Genome.Genes {
			if gene.Disabled || gene.Knockdown >= 1.0 {
				continue
			}
			occ := c.PromoterOccupancy(gene, s)
			if gene.Knockdown > 0 {
				occ *= (1 - gene.Knockdown)
			}
			r := gene.BasalTranscription * occ
			rates[gene.Name] = r
			totalProp += r
		}
		if totalProp <= 0 {
			break
		}
		// time to next event
		dt := -math.Log(rng.Float64()) / totalProp
		t += dt
		// pick event
		pick := rng.Float64() * totalProp
		cum := 0.0
		for name, r := range rates {
			cum += r
			if pick <= cum {
				state[name] += 1
				trajectories[name] = append(trajectories[name], state[name])
				break
			}
		}
	}
	return trajectories
}

// ==================== Parameter Estimation ====================

type ParameterEstimate struct {
	Param      string
	BestValue  float64
	BestError  float64
	DataCurve  []float64
	ModelCurve []float64
}

// EstimateParam searches for a parameter value that minimizes SSE against
// observed trajectory. 'param' selects which parameter (translation_k,
// basal_transcription, division_atp, base_death_rate).
func (s *Sim) EstimateParam(param string, observed []float64, searchMin, searchMax float64, steps int) ParameterEstimate {
	type candidate struct {
		value float64
		err   float64
		curve []float64
	}
	var best candidate
	best.err = math.Inf(1)

	for i := 0; i < steps; i++ {
		val := searchMin + (searchMax-searchMin)*float64(i)/float64(steps-1)
		p := s.Params
		switch param {
		case "translation_k":
			p.TranslationK = val
		case "division_atp":
			p.DivisionATP = val
		case "base_death_rate":
			p.BaseDeathRate = val
		case "mutation_rate":
			p.MutationRate = val
		}
		sim := NewSim(p, s.Rng.Int63())
		n := len(observed)
		if n == 0 {
			n = 1
		}
		curve := make([]float64, 0, n)
		for t := 0; t < n; t++ {
			sim.Step()
			curve = append(curve, float64(sim.AliveCount()))
		}
		var sse float64
		for j := 0; j < n; j++ {
			d := observed[j] - curve[j]
			sse += d * d
		}
		if sse < best.err {
			best.err = sse
			best.value = val
			best.curve = curve
		}
	}

	pe := ParameterEstimate{
		Param:      param,
		BestValue:  best.value,
		BestError:  best.err,
		DataCurve:  observed,
		ModelCurve: best.curve,
	}
	return pe
}

// ==================== Bifurcation Scan ====================

type BifurcationPoint struct {
	ParamValue float64
	Alive      int
	MaxGen     int
	Regime     string
}

// BifurcationScan runs a series of simulations varying one parameter and
// reports which "regime" the population lands in: extinct, stable, oscillating,
// or explosive. Uses CV of alive count over the last 20 ticks to classify.
func (s *Sim) BifurcationScan(param string, searchMin, searchMax float64, steps, ticks int) []BifurcationPoint {
	results := make([]BifurcationPoint, 0, steps)
	for i := 0; i < steps; i++ {
		val := searchMin + (searchMax-searchMin)*float64(i)/float64(steps-1)
		p := s.Params
		switch param {
		case "division_atp":
			p.DivisionATP = val
		case "translation_k":
			p.TranslationK = val
		case "base_death_rate":
			p.BaseDeathRate = val
		case "env_glucose":
			p.EnvGlucose = val
		}
		sim := NewSim(p, s.Rng.Int63())
		for t := 0; t < ticks; t++ {
			sim.Step()
		}
		window := ticks / 5
		if window < 20 {
			window = 20
		}
		if window > 100 {
			window = 100
		}
		if window > len(sim.History) {
			window = len(sim.History)
		}
		counts := []float64{}
		if window > 0 {
			start := len(sim.History) - window
			for j := 0; j < window; j++ {
				h := sim.History[start+j]
				counts = append(counts, float64(h.Alive))
			}
		}
		regime := classifyRegime(counts, sim.AliveCount())
		results = append(results, BifurcationPoint{
			ParamValue: val,
			Alive:      sim.AliveCount(),
			MaxGen:     sim.MaxGen(),
			Regime:     regime,
		})
	}
	return results
}

func classifyRegime(counts []float64, alive int) string {
	if alive == 0 {
		return "extinct"
	}
	if len(counts) < 3 {
		return "unknown"
	}
	mean := 0.0
	for _, c := range counts {
		mean += c
	}
	mean /= float64(len(counts))
	if mean < 1 {
		return "extinct"
	}
	var ss float64
	for _, c := range counts {
		ss += (c - mean) * (c - mean)
	}
	prev := counts[len(counts)-2]
	cur := counts[len(counts)-1]
	trend := (cur - prev) / mean
	cv := math.Sqrt(ss/float64(len(counts))) / mean
	if cv < 0.02 && math.Abs(trend) < 0.05 {
		return "stable"
	}
	if cv > 0.15 {
		if mean > 50 {
			return "oscillating"
		}
		return "unstable"
	}
	if trend > 0.1 {
		return "growing"
	}
	return "declining"
}

// ==================== Hardy-Weinberg ====================

type HWPrediction struct {
	AA, Aa, aa          float64
	Freq_A              float64
	Freq_a              float64
	ObsAA, ObsAa, Obsaa float64
	ChiSq               float64
}

// HardyWeinberg computes expected genotype frequencies from observed counts
// of one trait. In the sim, 'AA' = no mutations in a gene, 'Aa' = 1-2 mutations,
// 'aa' = 3+ mutations. Provides chi-square deviation from H-W equilibrium.
func (s *Sim) HardyWeinberg() HWPrediction {
	var aa, aA, AA int
	for _, c := range s.Cells {
		if !c.Alive {
			continue
		}
		mutCount := 0
		for _, g := range c.Genome.Genes {
			if g.Knockdown > 0.5 {
				mutCount++
			}
		}
		switch {
		case mutCount == 0:
			AA++
		case mutCount <= 2:
			aA++
		default:
			aa++
		}
	}
	total := float64(AA + aA + aa)
	if total == 0 {
		return HWPrediction{}
	}
	p := (2*float64(AA) + float64(aA)) / (2 * total)
	q := 1 - p
	expAA := p * p * total
	expAa := 2 * p * q * total
	expaa := q * q * total
	chi := 0.0
	if expAA > 0 {
		chi += math.Pow(float64(AA)-expAA, 2) / expAA
	}
	if expAa > 0 {
		chi += math.Pow(float64(aA)-expAa, 2) / expAa
	}
	if expaa > 0 {
		chi += math.Pow(float64(aa)-expaa, 2) / expaa
	}
	return HWPrediction{
		AA: expAA, Aa: expAa, aa: expaa,
		Freq_A: p, Freq_a: q,
		ObsAA: float64(AA), ObsAa: float64(aA), Obsaa: float64(aa),
		ChiSq: chi,
	}
}

// ==================== Plasmid Simulation ====================

type PlasmidState struct {
	CopyNumber int
	Gene       string
	Expression float64
}

// PlasmidExpression models a high-copy plasmid carrying one gene, producing
// its protein at copy-number-scaled rate. Useful for overexpression studies.
func PlasmidExpression(copyNumber int, basalRate float64, growthRate float64) float64 {
	if copyNumber <= 0 {
		return 0
	}
	return float64(copyNumber) * basalRate * (1 + growthRate)
}

// SortBifurcationForDisplay sorts bifurcation results for consistent plotting
func SortBifurcationForDisplay(points []BifurcationPoint) []BifurcationPoint {
	out := append([]BifurcationPoint{}, points...)
	sort.Slice(out, func(i, j int) bool { return out[i].ParamValue < out[j].ParamValue })
	return out
}

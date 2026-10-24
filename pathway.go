package main

// PathwayStep models one enzyme-catalyzed reaction:
//   Substrate --[Enzyme]--> Product
// Governed by Michaelis-Menten kinetics (Kcat, Km).
type PathwayStep struct {
	Enzyme    string
	Substrate string
	Product   string
	Kcat      float64
	Km        float64
}

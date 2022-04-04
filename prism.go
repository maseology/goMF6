package gomf6

import "github.com/maseology/mmaths"

type mf6prism struct {
	mmaths.Prism
	Q       []float64
	Conn    []int
	Por, H0 float64 // Bn, Por, Tn float64
}

// New prism constructor
func (p *mf6prism) new(z []complex128, q []float64, conn []int, top, bot, h0, tn, porosity float64) {
	var pp mmaths.Prism
	pp.New(z, top, bot)
	p.Z = z          // complex (planform) coordinates
	p.Top = top      // cell top
	p.Bot = bot      // cell bottom
	p.Q = q          // flux array
	p.Conn = conn    // flux connectivity
	p.A = pp.A       // planform area
	p.Por = porosity // cell porosity
	p.H0 = h0        // cell head
	// p.Bn = bn        // saturated thickness at time step tn
	// p.Tn = tn        // initial time step (both bn and tn will adjust in transient cases)
}

func (p *mf6prism) Saturation() float64 {
	if p.H0 <= p.Bot {
		return 0.
	}
	return (p.H0 - p.Bot) / (p.Top - p.Bot)
}

func (p *mf6prism) Qcentroid() [3]float64 {
	return [3]float64{-1., -1., -1.}
}

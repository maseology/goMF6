package gomf6

import (
	"log"

	"github.com/maseology/goMF6/readers"
	"github.com/maseology/mmaths"
)

func readCBB(fp string, jaxr []readers.JAxr, prims []*mmaths.Prism) (pflx [][]float64, pqw map[int]float64) {
	dat1D, dat2D := readers.ReadCBB(fp)

	pflx = make([][]float64, len(prims))
	if vals, ok := dat1D["FLOW-JA-FACE"]; ok {
		for _, ja := range jaxr {
			pflx[ja.From] = append(pflx[ja.From], vals[ja.ID])
		}

		// for i, ja := range jaxr {
		// 	pflx[ja.From] = append(pflx[ja.From], vals[i])
		// }

		// for i := range jaxr { // initialize
		// 	pflx[i] = []float64{0., 0., 0., 0., 0., 0.} // left-up-right-down-bottom-top
		// }
		// for _, ja := range jaxr {
		// 	// fmt.Printf("from %d to %d flux %v\n", ja.f, ja.t, vals[ja.i])
		// 	pflx[ja.f][ja.p] = vals[ja.i]
		// }
	}
	// pflx = make([][]float64, len(jaxr))
	// if vals, ok := dat1D["FLOW-JA-FACE"]; ok {
	// 	// fmt.Printf("\nFLOW-JA-FACE data (%d):\n", len(vals))
	// 	for i := range jaxr { // initialize
	// 		pflx[i] = []float64{0., 0., 0., 0., 0., 0.} // left-up-right-down-bottom-top
	// 	}
	// 	for _, ja := range jaxr {
	// 		// fmt.Printf("from %d to %d flux %v\n", ja.f, ja.t, vals[ja.i])
	// 		pflx[ja.f][ja.p] = vals[ja.i]
	// 	}
	// }
	if vals, ok := dat2D["RCH"]; ok {
		for i, v := range vals {
			if len(v) > 1 {
				log.Fatalln("MODFLOW CBC read error: RCH given with greater than 1 NDAT")
			}
			pflx[i] = append(pflx[i], v[0])
		}
	}

	pqw = make(map[int]float64)
	if vals, ok := dat2D["WEL"]; ok {
		for i, v := range vals {
			if len(v) > 1 {
				log.Fatalln("MODFLOW CBC read error: WEL given with greater than 1 NDAT")
			}
			if v[0] != 0. {
				pqw[i] += v[0]
			}
		}
	}
	if vals, ok := dat2D["CHD"]; ok {
		for i, v := range vals {
			if len(v) > 1 {
				log.Fatalln("MODFLOW CBC read error: CHD given with greater than 1 NDAT")
			}
			if v[0] != 0. {
				pqw[i] += v[0]
			}
		}
	}

	// fmt.Println("\nflux summary [left-up-right-down-bottom-top] well")
	// for i := 0; i < len(pflx); i++ {
	// 	fmt.Println(i, *pflx[i])
	// }

	return
}

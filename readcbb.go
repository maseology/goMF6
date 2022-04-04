package gomf6

import (
	"log"

	"github.com/maseology/goMF6/readers"
)

func readCBB(fp string, jaxr []readers.JAxr, nprims int) (pflx [][]float64, pqw map[int]float64) {
	dat1D, dat2D := readers.ReadCBB(fp)

	if val, ok := dat1D["FLOW-JA-FACE"]; ok {
		pflx = make([][]float64, nprims)
		for i, ja := range jaxr {
			pflx[ja.From] = append(pflx[ja.From], val[i])
		}

		// for i := range jaxr { // initialize
		// 	pflx[i] = []float64{0., 0., 0., 0., 0., 0.} // left-up-right-down-bottom-top
		// }
		// for _, ja := range jaxr {
		// 	// fmt.Printf("from %d to %d flux %v\n", ja.f, ja.t, val[ja.i])
		// 	pflx[ja.f][ja.p] = val[ja.i]
		// }
	}
	// pflx = make([][]float64, len(jaxr))
	// if val, ok := dat1D["FLOW-JA-FACE"]; ok {
	// 	// fmt.Printf("\nFLOW-JA-FACE data (%d):\n", len(val))
	// 	for i := range jaxr { // initialize
	// 		pflx[i] = []float64{0., 0., 0., 0., 0., 0.} // left-up-right-down-bottom-top
	// 	}
	// 	for _, ja := range jaxr {
	// 		// fmt.Printf("from %d to %d flux %v\n", ja.f, ja.t, val[ja.i])
	// 		pflx[ja.f][ja.p] = val[ja.i]
	// 	}
	// }
	if val, ok := dat2D["RCH"]; ok {
		panic("to fix, harcoded to left-up-right-down-bottom-top scheme")
		for i, v := range val {
			if len(v) > 1 {
				log.Fatalln("MODFLOW CBC read error: RCH given with greater than 1 NDAT")
			}
			pflx[i][5] = v[0]
		}
	}

	pqw = make(map[int]float64)
	if val, ok := dat2D["WEL"]; ok {
		for i, v := range val {
			if len(v) > 1 {
				log.Fatalln("MODFLOW CBC read error: WEL given with greater than 1 NDAT")
			}
			if v[0] != 0. {
				pqw[i] += v[0]
			}
		}
	}
	if val, ok := dat2D["CHD"]; ok {
		for i, v := range val {
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

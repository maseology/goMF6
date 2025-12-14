package readers

import (
	"bytes"
	"encoding/binary"
	"io"
	"log"

	"github.com/maseology/mmaths"
	"github.com/maseology/mmio"
)

type grbGridHreader struct {
	NCELLS, NLAY, NROW, NCOL, NJA int32
	XORIGIN, YORIGIN, ANGROT      float64
}

func (g *grbGridHreader) read(b *bytes.Reader) bool {
	err := binary.Read(b, binary.LittleEndian, g)
	// fmt.Println(*g)
	if err != nil {
		if err == io.EOF {
			return true
		}
		log.Fatalln("Fatal error: grbGridHread failed: ", err)
	}
	return false
}

func ReadGRBgrid(buf *bytes.Reader) ([]*mmaths.Prism, [][]int, []JAxr, []int, []int) {
	g := grbGridHreader{}
	g.read(buf)

	delr, delc := make(map[int]float64), make(map[int]float64)
	for j := 0; j < int(g.NCOL); j++ {
		delr[j] = mmio.ReadFloat64(buf) // cell width
	}
	for i := 0; i < int(g.NROW); i++ {
		delc[i] = mmio.ReadFloat64(buf) // cell height
		g.YORIGIN += delc[i]            // adjusting origin from lower-left to upper-left
	}

	top, botm := make([]float64, int(g.NROW*g.NCOL)), make([]float64, int(g.NCELLS))
	for i := 0; i < int(g.NROW*g.NCOL); i++ {
		top[i] = mmio.ReadFloat64(buf)
	}
	for i := 0; i < int(g.NCELLS); i++ {
		botm[i] = mmio.ReadFloat64(buf)
	}

	ia, ja := make([]int, int(g.NCELLS)+1), make([]int, int(g.NJA))
	for i := 0; i <= int(g.NCELLS); i++ {
		ia[i] = int(mmio.ReadInt32(buf)) - 1
	}
	for j := 0; j < int(g.NJA); j++ {
		ja[j] = int(mmio.ReadInt32(buf)) - 1
	}

	idomain, icelltype := make([]int, int(g.NCELLS)), make([]int, int(g.NCELLS))
	for i := 0; i < int(g.NCELLS); i++ {
		idomain[i] = int(mmio.ReadInt32(buf))
	}
	for i := 0; i < int(g.NCELLS); i++ {
		icelltype[i] = int(mmio.ReadInt32(buf)) //  specifies how saturated thickness is treated
	}

	if !mmio.ReachedEOF(buf) {
		log.Fatalln("Fatal error: ReadGRBgrid read 003 failed: have not reached EOF")
	}

	// fmt.Printf("  nl,nr,nc: %v,%v,%v; UL-origin: (%v, %v)\n", g.NLAY, g.NROW, g.NCOL, g.XORIGIN, g.YORIGIN)
	prsms, c := make([]*mmaths.Prism, 0, g.NCELLS), 0
	isunstructured := g.NROW*g.NCOL == g.NCELLS
	for k := 0; k < int(g.NLAY); k++ {
		cl, o := 0, complex(g.XORIGIN, g.YORIGIN) // converted to upper-left (above)
		for i := 0; i < int(g.NROW); i++ {
			dy := -delc[i]
			for j := 0; j < int(g.NCOL); j++ {
				dx := delr[j]
				// p1---p2   y       0---nc
				//  | c |    |       |       clockwise, left-top-right-bottom
				// p0---p3   0---x   nr
				z := []complex128{o + complex(0., dy), o, o + complex(dx, 0.), o + complex(dx, dy)}

				var p mmaths.Prism
				if isunstructured {
					p.New(z, top[c], botm[c])
				} else {
					if k == 0 {
						p.New(z, top[cl], botm[c])
					} else {
						p.New(z, botm[c-int(g.NROW*g.NCOL)], botm[c])
					}
				}
				prsms = append(prsms, &p)
				// fmt.Println(c, k, i, j, p.Z, p.Top, p.Bot)

				o += complex(dx, 0.)
				c++
				cl++
			}
			o = complex(g.XORIGIN, imag(o)+dy)
		}
	}

	conn := g.buildTopology(idomain) // make(map[int][]int)

	// check connections
	jaxrOut, jaxrcnt := make([]JAxr, g.NJA), 0
	for i := range int(g.NCELLS) {
		if idomain[i] > 0 { // For a cell (N) that has been eliminated with IDOMAIN, the value for IA(N) and IA(N+1) will be equal, indicating that there are no connections or flows for that cell.
			i1, c1 := make([]int, ia[i+1]-ia[i]), make([]int, ia[i+1]-ia[i]) // MF6 order (looks to be) above-up-left-right-down-below
			for j := ia[i]; j < ia[i+1]; j++ {
				c1[j-ia[i]] = ja[j]
				i1[j-ia[i]] = j
			}

			connkey := make(map[int]bool) // temporary map for list checking
			for _, v := range conn[i] {
				if v >= 0 {
					connkey[v] = true
				}
			}
			if c1[0] != i {
				log.Fatalf("Fatal error: ReadGRBgrid cell id check 004 failed:\nCreated: %v\nFound: %v\n", i, c1[0])
			}
			if len(c1)-1 != len(connkey) {
				log.Fatalf("Fatal error: ReadGRBgrid connectivity check 005 failed, cell %d\n:\nCreated: %v\nFound: %v\n", i, conn[i], c1[1:])
			}
			for _, c := range c1[1:] {
				if !connkey[c] {
					log.Fatalf("Fatal error: ReadGRBgrid connectivity check 006 failed, cell %d\n:\nCreated: %v\nFound: %v\n", i, conn[i], c1[1:])
				}
			}

			// add internal connection
			jaxrOut[jaxrcnt] = JAxr{
				From:     c1[0],
				To:       c1[0],
				Position: -1,
				ID:       i1[0],
			} // JA to prsm/conn cross-reference
			jaxrcnt++

			// fmt.Println(conn[i], c1[1:])
			for j, v := range conn[i] {
				if v < 0 {
					continue
				}
				for j2, v2 := range c1[1:] {
					if v == v2 {
						jaxrOut[jaxrcnt] = JAxr{
							From:     c1[0],
							To:       v,
							Position: j,
							ID:       i1[j2+1],
						} // JA to prsm.conn cross-reference
						jaxrcnt++
					}
				}
			}
		}
	}

	if jaxrcnt != int(g.NJA) {
		log.Fatalf("Fatal error: ReadGRBgrid connectivity check 007 failed, number of connections created (%d) not equal to NJA (less number of cells)\n", len(jaxrOut))
	}

	// fmt.Println("left-up-right-down-bottom-top")
	// fmt.Println("\nJAXR [from to pos ia]:")
	// for _, v := range jaxrOut {
	// 	fmt.Println(v)
	// }

	return prsms, conn, jaxrOut, idomain, icelltype
}

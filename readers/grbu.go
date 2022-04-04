package readers

import (
	"bytes"
	"encoding/binary"
	"io"
	"log"

	"github.com/maseology/mmaths"
	"github.com/maseology/mmio"
)

type grbuHreader struct {
	NODES, NJA               int32
	XORIGIN, YORIGIN, ANGROT float64
}

func (g *grbuHreader) read(b *bytes.Reader) bool {
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

func ReadGRBU(buf *bytes.Reader) ([]*mmaths.Prism, [][]int, []JAxr) {

	// 1. READ DATA
	g := grbuHreader{}
	g.read(buf)

	nc, nja := int(g.NODES), int(g.NJA)
	nvert := 4 * nc //////////////////////////////  HARD-CODED: for rectilinear cells only //////////////////////////////
	top, botm := make([]float64, nc), make([]float64, nc)
	if err := binary.Read(buf, binary.LittleEndian, top); err != nil {
		panic(err)
	}
	if err := binary.Read(buf, binary.LittleEndian, botm); err != nil {
		panic(err)
	}

	ia, ja, icelltype := make([]int32, nc+1), make([]int32, nja), make([]int32, nc)
	if err := binary.Read(buf, binary.LittleEndian, ia); err != nil { // For each cell n, the number of cell connections plus one.
		panic(err)
	}
	if err := binary.Read(buf, binary.LittleEndian, ja); err != nil { // For each cell n a list of connected m cells.
		panic(err)
	}
	if err := binary.Read(buf, binary.LittleEndian, icelltype); err != nil { // integer variable that defines if a cell is convertible or confined -- A value of zero indicates that the cell is confined. A nonzero value indicates that the cell is convertible.
		panic(err)
	}

	vert := make([][2]float64, nvert)
	cellx := make([]float64, nc)
	celly := make([]float64, nc)
	iavert := make([]int32, nc)
	if err := binary.Read(buf, binary.LittleEndian, vert); err != nil {
		panic(err)
	}
	if err := binary.Read(buf, binary.LittleEndian, cellx); err != nil {
		panic(err)
	}
	if err := binary.Read(buf, binary.LittleEndian, celly); err != nil {
		panic(err)
	}
	if err := binary.Read(buf, binary.LittleEndian, iavert); err != nil {
		panic(err)
	}

	var njavert int32
	if err := binary.Read(buf, binary.LittleEndian, &njavert); err != nil {
		panic(err)
	}
	javert := make([]int32, njavert-1)
	if err := binary.Read(buf, binary.LittleEndian, javert); err != nil {
		panic(err)
	}

	if !mmio.ReachedEOF(buf) {
		log.Fatalln("Fatal error: readGRB read 003 failed: have not reached EOF")
	}

	// 2. CREATE PRISMS AND CONNECTIVITY
	prsms := func() []*mmaths.Prism {
		prsms, njavert32 := make([]*mmaths.Prism, nc), int32(njavert)-1
		for i := 0; i < nc; i++ {
			i0, i1 := iavert[i]-1, njavert32
			if i < nc-1 {
				i1 = iavert[i+1] - 1
			}
			vids := javert[i0:i1]
			if vids[0] != vids[len(vids)-1] {
				panic("")
			}
			z := make([]complex128, len(vids)-1)
			for j := 0; j < len(vids)-1; j++ {
				vertex := vert[vids[j]-1]
				z[j] = complex(vertex[0], vertex[1])
			}
			prsms[i] = &mmaths.Prism{
				Z:   z,
				Top: top[i],
				Bot: botm[i],
			}
		}
		return prsms
	}()

	conn, jaxr := func() ([][]int, []JAxr) {
		nc := int(g.NODES)
		tp := make([][]int, nc)
		jaxr, cja := make([]JAxr, nja), 0
		for i := 0; i < nc; i++ {
			a := ja[ia[i]-1 : ia[i+1]-1]
			tpt := make([]int, len(a)-1) ////////////////////////  HARD-CODED: NOT DOING: //[]int{-1, -1, -1, -1, -1, -1} // initialize, [CW laterals]-bottom-top  (ex. left-up-right-down-bottom-top)
			for j := 0; j < len(a); j++ {
				if j > 0 {
					tpt[j-1] = int(a[j]) - 1
				}
				jaxr[cja] = JAxr{
					From:     int(a[0]) - 1,
					To:       int(a[j]) - 1,
					Position: j,
				}
				cja++
			}
			tp[a[0]-1] = tpt
		}
		return tp, jaxr
	}()

	// // // fmt.Printf("  nl,nr,nc: %v,%v,%v; UL-origin: (%v, %v)\n", g.NLAY, g.NROW, g.NCOL, g.XORIGIN, g.YORIGIN)
	// // c, cpl, prsms := 0, nc, make(map[int]*Prism)
	// // for k := 0; k < int(g.NLAY); k++ {
	// // 	cl, o := 0, complex(g.XORIGIN, g.YORIGIN) // converted to upper-left (above)
	// // 	for i := 0; i < int(g.NROW); i++ {
	// // 		dy := -delc[i]
	// // 		for j := 0; j < int(g.NCOL); j++ {
	// // 			dx := delr[j]
	// // 			// p1---p2   y       0---nc
	// // 			//  | c |    |       |       clockwise, left-top-right-bottom
	// // 			// p0---p3   0---x   nr
	// // 			z := []complex128{o + complex(0., dy), o, o + complex(dx, 0.), o + complex(dx, dy)}
	// // 			if idomain[c] >= 0 {
	// // 				var p Prism
	// // 				if k == 0 {
	// // 					p.New(z, top[c], botm[c], top[cl], 0., defaultPorosity)
	// // 				} else {
	// // 					for kk := k - 1; kk >= 0; kk-- {
	// // 						c0 := kk*cpl + cl
	// // 						if idomain[c0] > 0 {
	// // 							p.New(z, botm[c0], botm[c], top[cl], 0., defaultPorosity)
	// // 							break
	// // 						} else if idomain[c0] == 0 {
	// // 							p.New(z, botm[c0], botm[c], botm[c0], 0., defaultPorosity)
	// // 							break
	// // 						} else if kk == 0 {
	// // 							p.New(z, top[cl], botm[c], top[cl], 0., defaultPorosity)
	// // 						}
	// // 					}
	// // 				}
	// // 				prsms[c] = &p
	// // 				// fmt.Println(c, k, i, j, p.Z, p.Top, p.Bot)
	// // 			}
	// // 			o += complex(dx, 0.)
	// // 			c++
	// // 			cl++
	// // 		}
	// // 		o = complex(g.XORIGIN, imag(o)+dy)
	// // 	}
	// // }

	// // conn := g.buildTopology() // make(map[int][]int)
	// // // for i := 0; i < int(g.NCELLS); i++ {
	// // // 	c1 := make([]int, ia[i+1]-ia[i])
	// // // 	for j := ia[i]; j < ia[i+1]; j++ {
	// // // 		c1[j-ia[i]] = ja[j]
	// // // 	}
	// // // 	conn[i] = c1
	// // // }

	// // check connections
	// jaxrOut, jaxrcnt := make(map[int]JAxr), 0
	// for i := 0; i < nc; i++ {
	// 	i1, c1 := make([]int, ia[i+1]-ia[i]), make([]int, ia[i+1]-ia[i]) // MF6 order (looks to be) above-up-left-right-down-below
	// 	for j := ia[i]; j < ia[i+1]; j++ {
	// 		c1[j-ia[i]] = int(ja[j]) - 1
	// 		i1[j-ia[i]] = int(j) - 1
	// 	}

	// 	connkey := make(map[int]bool) // temporary map for list checking
	// 	for _, v := range conn[i] {
	// 		if v >= 0 {
	// 			connkey[v] = true
	// 		}
	// 	}
	// 	if c1[0] != i {
	// 		log.Fatalf("Fatal error: readGRB cell id check 004 failed:\nCreated: %v\nFound: %v\n", i, c1[0])
	// 	}
	// 	if len(c1)-1 != len(connkey) {
	// 		log.Fatalf("Fatal error: readGRB connectivity check 005 failed, cell %d\n:\nCreated: %v\nFound: %v\n", i, conn[i], c1[1:])
	// 	}
	// 	for _, c := range c1[1:] {
	// 		if !connkey[c] {
	// 			log.Fatalf("Fatal error: readGRB connectivity check 006 failed, cell %d\n:\nCreated: %v\nFound: %v\n", i, conn[i], c1[1:])
	// 		}
	// 	}

	// 	// fmt.Println(conn[i], c1[1:])
	// 	for j, v := range conn[i] {
	// 		for j2, v2 := range c1[1:] {
	// 			if v == v2 {
	// 				jaxrOut[jaxrcnt] = JAxr{
	// 					f: c1[0],
	// 					t: v,
	// 					p: j,
	// 					i: i1[j2+1],
	// 				} // JA to prsm.conn cross-reference
	// 				jaxrcnt++
	// 			}
	// 		}
	// 	}
	// }

	// if len(jaxrOut) != int(g.NJA)-nc {
	// 	log.Fatalf("Fatal error: readGRB connectivity check 007 failed, number of connections created (%d) not equal to NJA (less number of cells)\n", len(jaxrOut))
	// }

	return prsms, conn, jaxr
}

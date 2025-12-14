package readers

import (
	"encoding/binary"
	"log"

	"github.com/maseology/mmaths"
	"github.com/maseology/mmio"
)

func GetExternalUgridBinary(fp string) (prsms []*mmaths.Prism, conn [][]int, jaxr []JAxr) {
	buf := mmio.OpenBinary(fp)

	var nc, nvert, nja, njavert int32
	func() {
		if err := binary.Read(buf, binary.LittleEndian, &nc); err != nil {
			panic(err)
		}
		if err := binary.Read(buf, binary.LittleEndian, &nvert); err != nil {
			panic(err)
		}
		if err := binary.Read(buf, binary.LittleEndian, &nja); err != nil {
			panic(err)
		}
		if err := binary.Read(buf, binary.LittleEndian, &njavert); err != nil {
			panic(err)
		}
	}()

	top := make([]float32, nc)
	botm := make([]float32, nc)
	ia, ja := make([]int32, nc+1), make([]int32, nja)
	vert := make([][2]float32, nvert)
	iavert := make([]int32, nc+1)
	javert := make([]int32, njavert)
	func() {
		if err := binary.Read(buf, binary.LittleEndian, top); err != nil {
			panic(err)
		}
		if err := binary.Read(buf, binary.LittleEndian, botm); err != nil {
			panic(err)
		}
		if err := binary.Read(buf, binary.LittleEndian, ia); err != nil {
			panic(err)
		}
		if err := binary.Read(buf, binary.LittleEndian, ja); err != nil {
			panic(err)
		}
		if err := binary.Read(buf, binary.LittleEndian, vert); err != nil {
			panic(err)
		}
		if err := binary.Read(buf, binary.LittleEndian, iavert); err != nil {
			panic(err)
		}
		if err := binary.Read(buf, binary.LittleEndian, javert); err != nil {
			panic(err)
		}

		if !mmio.ReachedEOF(buf) {
			log.Fatalln("Fatal error: getExternalUgridBinary read failed: have not reached EOF")
		}
	}()

	prsms = func() []*mmaths.Prism {
		prsms, njavert32 := make([]*mmaths.Prism, nc), int32(njavert)-1
		for i := range nc {
			i0, i1 := iavert[i], njavert32
			if i < nc {
				i1 = iavert[i+1]
			}
			vids := javert[i0:i1]
			if vids[0] != vids[len(vids)-1] {
				panic("GetExternalUgridBinary error, not a closed polygon")
			}
			z := make([]complex128, len(vids)-1)
			for j := 0; j < len(vids)-1; j++ {
				vertex := vert[vids[j]]
				z[j] = complex(float64(vertex[0]), float64(vertex[1]))
			}

			// a := 0.
			// nfaces := len(z)
			// for j := range z {
			// 	jj := (j + 1) % nfaces
			// 	a += real(z[j])*imag(z[jj]) - real(z[jj])*imag(z[j])
			// }
			// a /= -2. // negative used here because vertices are entered in clockwise order
			// if a <= 0. {
			// 	panic("GetExternalUgridBinary area calculation error, may be given in counter-clockwise order")
			// }

			prsms[i] = &mmaths.Prism{
				Z:   z,
				Top: float64(top[i]),
				Bot: float64(botm[i]),
			}
		}
		return prsms
	}()

	conn, jaxr = func() ([][]int, []JAxr) {
		tp := make([][]int, nc)
		jaxr, cja := make([]JAxr, nja), 0
		for i := range nc {
			a := ja[ia[i]:ia[i+1]]
			tpt := make([]int, len(a)-1)
			for j := range len(a) {
				if j > 0 {
					tpt[j-1] = int(a[j])
				}
				jaxr[cja] = JAxr{
					From:     int(a[0]),
					To:       int(a[j]),
					Position: j,
					ID:       cja,
				}
				cja++
			}
			tp[a[0]] = tpt
		}
		return tp, jaxr
	}()

	return prsms, conn, jaxr
}

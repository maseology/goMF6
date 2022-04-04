package readers

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/maseology/mmio"
)

type cbcHreader struct {
	KSTP, KPER          int32
	TEXT                [16]byte
	NDIM1               int32 // NCOL (DIS); NCPL (DISV); NODES (DISU); NJA (FLOW-JA-FACE, IMETH=1)
	NDIM2               int32 // NROW (DIS);    1 (DISV);     1 (DISU);   1 (FLOW-JA-FACE, IMETH=1)
	NDIM3               int32 // NLAY (DIS); NLAY (DISV);     1 (DISU);   1 (FLOW-JA-FACE, IMETH=1)
	IMETH               int32
	DELT, PERTIM, TOTIM float64
}

func (h *cbcHreader) cbcHread(b *bytes.Reader) bool {
	err := binary.Read(b, binary.LittleEndian, h)
	if err != nil {
		if err == io.EOF {
			return true
		}
		log.Fatalln("Fatal error: cbcHread failed: ", err)
	}
	return false
}

type cbcAuxReader struct {
	TXT1ID1, TXT2ID1, TXT1ID2, TXT2ID2 [16]byte
	NDAT                               int32
}

func (a *cbcAuxReader) cbcAuxRead(b *bytes.Reader) {
	err := binary.Read(b, binary.LittleEndian, a)
	if err != nil {
		log.Fatalln("Fatal error: cbcAuxRead failed: ", err)
	}
}

func ReadCBB(fp string) (dat1D map[string][]float64, dat2D map[string][][]float64) {
	bflx := mmio.OpenBinary(fp)
	dat1D = make(map[string][]float64)
	dat2D = make(map[string][][]float64)

	for {
		h := cbcHreader{}
		if h.cbcHread(bflx) {
			break // EOF
		}

		txt := strings.TrimSpace(string(h.TEXT[:]))
		// fmt.Printf("KSTP %d; KPER %d: %s\n", h.KPER, h.KSTP, txt)
		switch h.IMETH {
		case 1: // Read 1D array of size NDIM1*NDIM2*NDIM3
			n := int(-h.NDIM1 * h.NDIM2 * h.NDIM3)
			m1 := make([]float64, n)
			// for i := 0; i < n; i++ {
			// 	m1[i] = mmio.ReadFloat64(bflx)
			// }
			if err := binary.Read(bflx, binary.LittleEndian, m1); err != nil {
				panic(err)
			}
			dat1D[txt] = m1
		case 6: // Read text identifiers, auxiliary text labels, and list of information.
			a := cbcAuxReader{}
			a.cbcAuxRead(bflx)
			nd := int(a.NDAT)
			auxtext := make([]string, nd)
			for i := 0; i < nd-1; i++ {
				var b1 [16]byte
				if err := binary.Read(bflx, binary.LittleEndian, &b1); err != nil {
					log.Fatalln("Fatal error: AUXTEXT read failed: ", err)
				}
				auxtext[i] = string(b1[:])
			}
			var nlist int32
			if err := binary.Read(bflx, binary.LittleEndian, &nlist); err != nil {
				log.Fatalln("Fatal error: NLIST read failed: ", err)
			}
			d2D := make([][]float64, nlist)
			for i := 0; i < int(nlist); i++ {
				var id1, id2 int32
				if err := binary.Read(bflx, binary.LittleEndian, &id1); err != nil {
					log.Fatalln("Fatal error: ID1 read failed: ", err)
				}
				if err := binary.Read(bflx, binary.LittleEndian, &id2); err != nil {
					log.Fatalln("Fatal error: ID2 read failed: ", err)
				}
				m1 := make([]float64, nd)
				for j := 0; j < nd; j++ {
					m1[j] = mmio.ReadFloat64(bflx)
				}
				d2D[int(id1)-1] = m1
			}
			dat2D[txt] = d2D
		default:
			log.Fatalf("MODFLOW CBB read error: IMETH=%d not supported", h.IMETH)
		}
	}

	// print available outputs
	fmt.Println("  CBB: 2D")
	for i := range dat2D {
		fmt.Printf("      %s\n", i)
	}
	fmt.Println("  CBB: 1D")
	for i := range dat1D {
		fmt.Printf("      %s\n", i)
	}
	return
}

package readers

import (
	"bytes"
	"encoding/binary"
	"io"
	"log"
	"strings"

	"github.com/maseology/mmio"
)

type dvarHreader struct {
	KSTP, KPER    int32
	PERTIM, TOTIM float64
	TEXT          [16]byte
	NCOL          int32 // NCOL (DIS); NCPL (DISV); NODES (DISU)
	NROW          int32 // NROW (DIS);    1 (DISV);     1 (DISU)
	ILAY          int32 // NLAY (DIS); NLAY (DISV);     1 (DISU)
}

func (h *dvarHreader) dvarHread(b *bytes.Reader) bool {
	err := binary.Read(b, binary.LittleEndian, h)
	if err != nil {
		if err == io.EOF {
			return true
		}
		log.Fatalln("Fatal error: dvarHread failed: ", err)
	}
	return false
}

func ReadDependentVariable(fp string) map[string][]float64 {
	bflx, m1 := mmio.OpenBinary(fp), make(map[string][]float64)
	for {
		h := dvarHreader{}
		if h.dvarHread(bflx) {
			break // EOF
		}

		txt := strings.TrimSpace(string(h.TEXT[:]))
		nc := int(h.NROW * h.NCOL)
		m2 := make([]float64, nc)
		if err := binary.Read(bflx, binary.LittleEndian, m2); err != nil {
			panic(err)
		}
		m1[txt] = append(m1[txt], m2...)
		// // fmt.Printf("Layer %d; KSTP %d; KPER %d: %s\n", h.ILAY, h.KPER, h.KSTP, txt)
		// m2, c := make(map[int]float64), int(h.ILAY-1)*int(h.NROW*h.NCOL)
		// for i := 0; i < int(h.NROW); i++ {
		// 	for j := 0; j < int(h.NCOL); j++ {
		// 		m2[c] = mmio.ReadFloat64(bflx)
		// 		c++
		// 	}
		// }
		// if m1[txt] == nil {
		// 	m1[txt] = make(map[int]float64)
		// }
		// for i, v := range m2 {
		// 	m1[txt][i] = v
		// }
	}
	return m1

	// 	Using br As New BinaryReader(New FileStream(_filepath, FileMode.Open), System.Text.Encoding.Default)
	// 	Dim cnt As Integer = 1
	// 100:            Dim KSTP = br.ReadInt32
	// 	Dim KPER = br.ReadInt32
	// 	Dim PERTIM = br.ReadDouble
	// 	Dim TOTIM = br.ReadDouble
	// 	Dim TEXT = mmIO.HexToString(BitConverter.ToString(br.ReadBytes(16))).Trim
	// 	Dim NCOL = br.ReadInt32 ' NCPL (DISV); NODES (DISU)
	// 	Dim NROW = br.ReadInt32 ' if DISV or DISU, NROW=1
	// 	Dim ILAY = br.ReadInt32 ' if DISU, ILAY=1
	// 	Dim dic1 As New Dictionary(Of Integer, Double), cnt2 = 0
	// 	For i = 1 To NROW
	// 		For j = 1 To NCOL
	// 			For k = 1 To ILAY
	// 				dic1.Add(cnt2, br.ReadDouble)
	// 				cnt2 += 1
	// 			Next
	// 		Next
	// 	Next
	// 	_t.Add(cnt, dic1)
	// 	_tnam.Add(cnt, String.Format("{0}_SP{1:00000}_TS{2:00000}", TEXT, KPER, KSTP))
	// 	cnt += 1
	// 	If br.PeekChar <> -1 Then GoTo 100
	// End Using
}

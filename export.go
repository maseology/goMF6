package gomf6

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"log"
	"sort"
	"time"
)

// ExportVTK saves model domain as a *.vtk file for visualization.
func (mf6 *MF6) ExportToVTK(filepath string, vertExag float64) {
	// collect cell ids, building flow field
	fmt.Println(" exporting VTK flow field..")
	nprsm, cids := func() (int, []int) {
		cids, ii := make([]int, len(mf6.prsms)), 0
		for i := range mf6.prsms {
			cids[ii] = i
			ii++
		}
		sort.Ints(cids)
		nprsm := len(cids)
		return nprsm, cids
	}()
	vtkReorder := func(s []int) []int {
		l := len(s) / 2
		f := func(s []int) []int {
			for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
				s[i], s[j] = s[j], s[i]
			}
			return s
		}
		return append(f(s[:l]), f(s[l:])...)
	}

	// collect vertices
	v, vxr, nvert := func() (map[int][]float64, map[int][]int, int) {
		v, vxr, cnt := make(map[int][]float64), make(map[int][]int), 0
		for _, i := range cids {
			p, s1 := mf6.prsms[i], make([]int, 0)
			for _, c := range p.Z {
				v[cnt] = []float64{real(c), imag(c), p.Top * vertExag}
				s1 = append(s1, cnt)
				cnt++
			}
			for _, c := range p.Z {
				v[cnt] = []float64{real(c), imag(c), p.Bot * vertExag}
				s1 = append(s1, cnt)
				cnt++
			}
			vxr[i] = vtkReorder(s1)
		}
		nvert := cnt
		return v, vxr, nvert
	}()

	// write to data buffer
	buf, endi := new(bytes.Buffer), binary.BigEndian

	binary.Write(buf, endi, []byte("# vtk DataFile Version 3.0\n"))
	binary.Write(buf, endi, []byte(fmt.Sprintf("Unstructured prism domain: %d Prisms, %d vertices, %s\n", nprsm, nvert, time.Now().Format("2006-01-02 15:04:05"))))
	binary.Write(buf, endi, []byte("BINARY\n"))
	binary.Write(buf, endi, []byte("DATASET UNSTRUCTURED_GRID\n"))

	binary.Write(buf, endi, []byte(fmt.Sprintf("POINTS %d float\n", nvert)))
	for i := 0; i < nvert; i++ {
		binary.Write(buf, endi, float32(v[i][0]))
		binary.Write(buf, endi, float32(v[i][1]))
		binary.Write(buf, endi, float32(v[i][2]))
	}

	binary.Write(buf, endi, []byte(fmt.Sprintf("\nCELLS %d %d\n", nprsm, nprsm+nvert)))
	for _, i := range cids {
		binary.Write(buf, endi, int32(len(mf6.prsms[i].Z)*2))
		for _, nid := range vxr[i] {
			binary.Write(buf, endi, int32(nid))
		}
	}

	binary.Write(buf, endi, []byte(fmt.Sprintf("\nCELL_TYPES %d\n", nprsm)))
	for _, i := range cids {
		switch len(mf6.prsms[i].Z) {
		case 0, 1, 2:
			log.Fatalf("ExportVTK error: invalid prism shape")
		case 3:
			binary.Write(buf, endi, int32(13)) // VTK_WEDGE
		case 4:
			binary.Write(buf, endi, int32(12)) // VTK_HEXAHEDRON
		case 5:
			binary.Write(buf, endi, int32(15)) // VTK_PENTAGONAL_PRISM
		case 6:
			binary.Write(buf, endi, int32(16)) // VTK_HEXAGONAL_PRISM
		default:
			log.Fatalf("ExportVTK todo: >6 sided polyhedron")
		}
	}

	// cell index
	binary.Write(buf, endi, []byte(fmt.Sprintf("\nCELL_DATA %d\n", nprsm)))
	binary.Write(buf, endi, []byte(fmt.Sprintf("SCALARS cellID int32\n")))
	binary.Write(buf, endi, []byte(fmt.Sprintf("LOOKUP_TABLE default\n")))
	for _, i := range cids {
		binary.Write(buf, endi, int32(i))
	}

	// saturation
	binary.Write(buf, endi, []byte(fmt.Sprintf("SCALARS saturation float\n")))
	binary.Write(buf, endi, []byte(fmt.Sprintf("LOOKUP_TABLE default\n")))
	for _, i := range cids {
		binary.Write(buf, endi, float32(mf6.prsms[i].Saturation()))
	}

	// // mean flux
	// binary.Write(buf, endi, []byte(fmt.Sprintf("\nVECTORS Vcentroid double\n")))
	// for _, i := range cids {
	// 	binary.Write(buf, endi, mf6.prsms[i].Qcentroid())
	// }

	// write to file
	if err := ioutil.WriteFile(filepath, buf.Bytes(), 0644); err != nil {
		log.Fatalf("ioutil.WriteFile failed: %v", err)
	}
}

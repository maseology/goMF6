package gomf6

import (
	"fmt"

	"github.com/maseology/goMF6/readers"
	"github.com/maseology/mmio"
)

func ReadMF6(fprfx string) MF6 {
	grbfp := fmt.Sprintf("%s.disu.grb", fprfx)
	if _, ok := mmio.FileExists(grbfp); !ok {
		grbfp = fmt.Sprintf("%s.dis.grb", fprfx)
		if _, ok := mmio.FileExists(grbfp); !ok {
			panic("ReadMODFLOW: no grb found")
		}
	}
	pset, conn, jaxr := readGRB(grbfp)

	// collect fluxes
	fpcbc := fmt.Sprintf("%s.cbb", fprfx)
	if _, ok := mmio.FileExists(fpcbc); !ok {
		fpcbc = fmt.Sprintf("%s.cbc", fprfx)
		if _, ok := mmio.FileExists(fpcbc); !ok {
			fpcbc = fmt.Sprintf("%s.flx", fprfx)
			if _, ok := mmio.FileExists(fpcbc); !ok {
				panic("ReadMODFLOW: not all required MF6 files found")
			}
		}
	}
	pflx, pqw := readCBB(fpcbc, jaxr, len(pset))

	// convert prism type
	mfpset := make([]*mf6prism, len(pset))
	for i, p := range pset {
		pp := mf6prism{}
		pp.new(p.Z, pflx[i], conn[i], p.Top, p.Bot, -9999., 0., defaultPorosity)
		mfpset[i] = &pp
	}

	// collect heads
	func() {
		for m, v := range readers.ReadDependentVariable(fmt.Sprintf("%s.hds", fprfx)) {
			fmt.Printf("  DV: %s\n", m)
			if m == "HEAD" {
				for i, vv := range v {
					if vv < mfpset[i].Top {
						mfpset[i].H0 = vv
					} else {
						mfpset[i].H0 = mfpset[i].Top
					}
				}
			}
		}
	}()

	return MF6{
		Prsms: mfpset,
		Qw:    pqw,
		Zw: func() map[int]complex128 {
			zc := make(map[int]complex128, len(pqw))
			for k := range pqw {
				zc[k] = mfpset[k].Centroid()
			}
			return zc
		}(),
	}
}

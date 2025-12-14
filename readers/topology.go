package readers

import "log"

func (g *grbGridHreader) buildTopology(idomain []int) [][]int {
	cid, nl, nr, nc := 0, int(g.NLAY), int(g.NROW), int(g.NCOL)
	// fmt.Println(nl, nr, nc)
	tp := make([][]int, int(g.NCELLS))
	for k := range nl {
		for i := range nr {
			for j := range nc {
				c1 := []int{-1, -1, -1, -1, -1, -1} // initialize, left-up-right-down-bottom-top
				if cid == 110404 {
					print("")
				}
				if idomain[cid] > 0 {

					// left
					if j > 0 {
						if idomain[cid-1] > 0 {
							c1[0] = cid - 1
						}
					}

					// up
					if i > 0 {
						if idomain[cid-nc] > 0 {
							c1[1] = cid - nc
						}
					}

					// right
					if j < nc-1 {
						if idomain[cid+1] > 0 {
							c1[2] = cid + 1
						}
					}

					// down
					if i < nr-1 {
						if idomain[cid+nc] > 0 {
							c1[3] = cid + nc
						}
					}

					// bottom/below
					if k < nl-1 {
						idom1 := idomain[cid+nc*nr]
						if idom1 > 0 {
							c1[4] = cid + nc*nr
						} else if idom1 < 0 {
							if cid+2*nc*nr < len(idomain) && idomain[cid+2*nc*nr] > 0 {
								c1[4] = cid + nc*nr*2
							}
						} else {
							log.Fatalln("Fatal error: grbGridHreader.buildTopology idomain DOWN issue")
						}
					}

					// top/above
					if k > 0 {
						idom1 := idomain[cid-nc*nr]
						if idom1 > 0 {
							c1[5] = cid - nc*nr
						} else if idom1 < 0 {
							if cid-2*nc*nr >= 0 && idomain[cid-2*nc*nr] > 0 {
								c1[5] = cid - nc*nr*2
							}
						} else {
							log.Fatalln("Fatal error: grbGridHreader.buildTopology idomain UP issue")
						}
					}

				}

				tp[cid] = c1
				cid++
			}
		}
	}
	return tp
}

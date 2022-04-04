package readers

func (g *grbGridHreader) buildTopology() [][]int {
	cid, nl, nr, nc := 0, int(g.NLAY), int(g.NROW), int(g.NCOL)
	// fmt.Println(nl, nr, nc)
	tp := make([][]int, 0, int(g.NCELLS))
	for k := 0; k < nl; k++ {
		for i := 0; i < nr; i++ {
			for j := 0; j < nc; j++ {
				c1 := []int{-1, -1, -1, -1, -1, -1} // initialize, left-up-right-down-bottom-top

				// left
				if j > 0 {
					c1[0] = cid - 1
				}

				// up
				if i > 0 {
					c1[1] = cid - nc
				}

				// right
				if j < nc-1 {
					c1[2] = cid + 1
				}

				// down
				if i < nr-1 {
					c1[3] = cid + nc
				}

				// bottom/below
				if k < nl-1 {
					c1[4] = cid + nc*nr
				}

				// top/above
				if k > 0 {
					c1[5] = cid - nc*nr
				}

				tp[cid] = c1
				cid++
			}
		}
	}
	return tp
}

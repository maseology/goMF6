package gomf6

type MF6 struct {
	prsms []*mf6prism        // prism dimensions
	zw    map[int]complex128 // well(/point) flux/sink-source coordinate
	qw    map[int]float64    // well(/point) flux/sink-source
}

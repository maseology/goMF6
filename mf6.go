package gomf6

type MF6 struct {
	Prsms []*mf6prism        // prism dimensions
	Zw    map[int]complex128 // well(/point) flux/sink-source coordinate
	Qw    map[int]float64    // well(/point) flux/sink-source
}

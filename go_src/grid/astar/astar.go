package astar

import(
	"math"
)


motion := [8][3]float64{{1,0,1},{0,1,1},{-1,0,1},{0,-1,1},{-1,-1,math.Sqrt(3)},{-1,1,math.Sqrt(3)},{1,-1,math.Sqrt(3)},{1,1,math.Sqrt(3)}}

type Astar struct {
	Reso   float64
	Rr     float64
	Motion [8][3]float64
	ObjMap[][]bool
	
	minX float64
	minY float64
	maxX float64
	maxY float64

	xWidth float64
	yWidth float64
}

func NewAstar(ox, oy []float64, reso, rr float64) *Astar {
	a := new(Astar)
	a.Reso = reso
	a.Rr = rr
	a.Motion = [8][3]float64{{1,0,1},{0,1,1},{-1,0,1},{0,-1,1},{-1,-1,math.Sqrt(3)},{-1,1,math.Sqrt(3)},{1,-1,math.Sqrt(3)},{1,1,math.Sqrt(3)}}
	return a
}



package grid

import (
	"log"
	"math"
)

func heuristic(n1, n2 *AstarNode) float64 {
	w := 1.0
	d := w * math.Hypot(float64(n1.Ix)-float64(n2.Ix), float64(n1.Iy)-float64(n2.Iy))
	return d
}

// nodeを受取idを返す
func gridIndex(g *Grid, n *AstarNode) int {
	return n.Iy*g.XWidth + n.Ix
}

// x,yどちらかのindexを受取座標を返す
func gridPos(g *Grid, index, minP int) float64 {
	pos := float64(index)*g.Reso + float64(minP)
	return pos
}

func AstarPlan(g *Grid, sx, sy, gx, gy float64, hidden []int) (rx, ry []float64, Notfail bool) {
	nstart := NewNodeG(g.xyIndex(sx, g.MinX), g.xyIndex(sy, g.MinY), 0.0, -1)
	ngoal := NewNodeG(g.xyIndex(gx, g.MinX), g.xyIndex(gy, g.MinY), 0.0, -1)

	if !g.VerifyGridP(gridIndex(g, nstart), hidden) {
		log.Print("start point is not verified")
		return rx, ry, false
	}

	open_set := make(map[int]*AstarNode)
	close_set := make(map[int]*AstarNode)
	open_set[gridIndex(g, nstart)] = nstart
	count := 0

	for {
		count += 1
		if len(open_set) == 0 {
			log.Print("open set is empty.. count is ", count)
			return rx, ry, false
		}

		minCost := 9999999.9
		minKey := -1
		for key, val := range open_set {
			calCost := val.Cost + heuristic(ngoal, val)
			if calCost < minCost {
				minCost = calCost
				minKey = key
			}
		}
		cId := minKey
		current := open_set[cId]

		if current.Ix == ngoal.Ix && current.Iy == ngoal.Iy {
			log.Print("find goal")
			ngoal.Pind = current.Pind
			ngoal.Cost = current.Cost
			rx, ry = finalPath(g, ngoal, close_set)
			return rx, ry, true
		}

		delete(open_set, cId)

		close_set[cId] = current

		var nId int
		var node *AstarNode
		motion := [8][3]float64{{1.0, 0, 1.0}, {0, 1.0, 1.0}, {-1.0, 0, 1.0}, {0, -1.0, 1.0}, {-1.0, -1.0, math.Sqrt(2)}, {-1.0, 1.0, math.Sqrt(2)}, {1.0, -1.0, math.Sqrt(2)}, {1.0, 1.0, math.Sqrt(2)}}
		for _, v := range motion {
			node = NewNodeG(current.Ix+int(v[0]), current.Iy+int(v[1]), current.Cost+v[2], cId)
			nId = gridIndex(g, node)

			if !g.VerifyGridP(gridIndex(g, node), hidden) {
				continue
			}

			if _, ok := close_set[nId]; ok {
				continue
			}

			if _, ok := open_set[nId]; !ok {
				open_set[nId] = node
			}
		}
	}
}

// 最後に経路の順番にする
func finalPath(g *Grid, ngoal *AstarNode, closeSet map[int]*AstarNode) (rx, ry []float64) {
	log.Print("calculating final path... ")
	rx = append(rx, gridPos(g, ngoal.Ix, g.MinX))
	ry = append(ry, gridPos(g, ngoal.Iy, g.MinY))

	pind := ngoal.Pind
	for pind != -1 {
		n := closeSet[pind]
		rx = append(rx, gridPos(g, n.Ix, g.MinX))
		ry = append(ry, gridPos(g, n.Iy, g.MinY))
		pind = n.Pind
	}
	return rx, ry
}

// each grid for astar planning
type AstarNode struct {
	Index int
	Ix    int
	Iy    int

	Cost float64
	Pind int

	Obj bool //障害物ならtrue
}

// for astar constructor
func NewNodeG(ix, iy int, cost float64, pind int) *AstarNode {
	n := new(AstarNode)
	n.Ix = ix
	n.Iy = iy
	n.Cost = cost
	n.Pind = pind
	return n
}

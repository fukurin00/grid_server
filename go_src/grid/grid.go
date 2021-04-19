package grid

import (
	"image"
	"image/color"
	"log"
	"math"
	"os"

	msg "github.com/fukurin00/grid_server/msg"
	tools "github.com/fukurin00/grid_server/tools"
	_ "github.com/jbuchbinder/gopnm"
)

// motion := [8][3]float64{{1.0, 0, 1.0},{0,1.0,1.0},{-1.0,0,1.0},{0,-1.0,1.0},{-1.0,-1.0,math.Sqrt(3)},{-1.0,1.0,math.Sqrt(3)},{1.0,-1.0,math.Sqrt(3)},{1.0,1.0,math.Sqrt(3)}}

type Grid struct {
	Reso float64

	MinX int
	MaxX int
	MinY int
	MaxY int

	XWidth int
	YWidth int

	OList []Point
	Ox    []float64
	Oy    []float64

	ObjMap [][]bool

	Nodes map[int]*Node
}

type Point struct {
	X float64
	Y float64
}

func NewGridNo() *Grid {
	g := new(Grid)
	return g
}

func NewGrid(reso float64) *Grid {
	g := new(Grid)

	g.Reso = reso

	return g
}

// Calculating Object Map on Grid
func (g *Grid) CalcObjMap(rr float64) {
	g.MaxX = int(math.Round(tools.MaxFloat(g.Ox)))
	g.MaxY = int(math.Round(tools.MaxFloat(g.Oy)))
	g.MinX = int(math.Round(tools.MinFloat(g.Ox)))
	g.MinY = int(math.Round(tools.MinFloat(g.Oy)))

	g.XWidth = int(math.Round(float64(g.MaxX) - float64(g.MinX)))
	g.YWidth = int(math.Round(float64(g.MaxY) - float64(g.MinY)))

	g.ObjMap = make([][]bool, g.XWidth)
	for i := 0; i < g.XWidth; i++ {
		g.ObjMap[i] = make([]bool, g.YWidth)
	}
	g.Nodes = make(map[int]*Node)

	for ix := 0; ix < g.XWidth; ix++ {
		x := g.CalcXYPosition(ix, g.MinX)
		for iy := 0; iy < g.YWidth; iy++ {
			y := g.CalcXYPosition(iy, g.MinY)

			ind := iy*g.XWidth + ix
			g.Nodes[ind] = NewNode(ind, ix, iy, x, y)
			for _, ip := range g.OList {
				d := math.Hypot(ip.X-float64(x), ip.Y-float64(y))
				if d < rr {
					g.ObjMap[ix][iy] = true
					g.Nodes[ind].Obj = true
					break
				}
			}
		}
	}
	log.Print("complete calculate objmap")
	log.Print(`minX, maxX, minY, maxY
		xwidth, ywidth`)
	log.Print(g.MinX, g.MaxX, g.MinY, g.MaxY)
	log.Print(g.XWidth, g.YWidth)
}

func (g Grid) CalcXYPosition(index, minP int) float64 {
	pos := float64(index)*g.Reso + float64(minP)
	return pos
}

func (g Grid) CalcPosition(index int) (float64, float64) {
	// px := g.MinX + math.Round(float64(index%g.XWidth))*g.Reso
	// py := g.MinY + math.Round(float64(index/g.XWidth))*g.Reso
	px := g.Nodes[index].X
	py := g.Nodes[index].Y
	return px, py
}

// だめなグリッドを任意で追加する
func (g Grid) VerifyGridP(index int, hidden []int) bool {
	if tools.CheckSameCom(index, hidden) {
		return false
	}
	return g.VerifyGrid(index)
}

func (g Grid) VerifyGrid(index int) bool {
	if index > g.XWidth*g.YWidth-1 {
		return false
	}
	px, py := g.CalcPosition(index)

	if px < float64(g.MinX) {
		return false
	} else if py < float64(g.MinY) {
		return false
	} else if px >= float64(g.MaxX) {
		return false
	} else if py >= float64(g.MaxY) {
		return false
	}

	if g.ObjMap[g.Nodes[index].Ix][g.Nodes[index].Iy] {
		return false
	}
	return true
}

func (g Grid) PosToGrid(x, y float64) int {
	return int(math.Round(y)) - g.MinY*g.XWidth + int(math.Round(x)) - g.MinX
}

func (g Grid) CalcRobotGrid(x, y, rr float64) []int {
	var overs []int
	center := g.PosToGrid(x, y)
	overs = append(overs, center)

	around := [8]int{-1, 1, -1 * g.XWidth, g.XWidth, -1*g.XWidth - 1, -1*g.XWidth + 1, g.XWidth - 1, g.XWidth + 1}
	aspect := int(math.Round(rr / (2 * g.Reso)))
	for i := 1; i < aspect+1; i++ {
		for _, v := range around {
			target := center + i*v
			if g.VerifyGrid(target) {
				tX, tY := g.CalcPosition(target)
				d := math.Hypot(float64(tX)-x, float64(tY)-y)
				if d <= rr {
					overs = append(overs, target)
				}
			}
		}
	}
	return overs
}

func (g *Grid) ReadMapImage(yamlFile, mapFile string) error {
	mapConfig := msg.ReadImageYaml(yamlFile)
	reso := mapConfig.Resolution
	origins := mapConfig.Origin
	log.Print("resolution:", reso, " origins:", origins)

	file, err := os.Open(mapFile)
	if err != nil {
		return err
	}
	defer file.Close()

	imageData, _, err := image.Decode(file)
	if err != nil {
		return err
	}

	g.OList = nil
	g.Ox = nil
	g.Oy = nil

	bound := imageData.Bounds()
	imgSet := image.NewGray(bound)

	maxX := bound.Max.X
	maxY := bound.Max.Y

	for i := 0; i < maxX; i++ {
		for j := 0; j < maxY; j++ {
			oldPix := imageData.At(i, j)
			pixel := color.GrayModel.Convert(oldPix)
			imgSet.Set(i, j, pixel)
		}
	}
	//rImg := tools.ReverseSlice(imgSet.Pix)
	stride := imgSet.Stride

	insideWall := false
	for i, pixel := range imgSet.Pix {
		if i%2 != 0 {
			continue
		}
		if pixel == 0 {
			if insideWall {
				continue
			}
			x := float64(i%stride)*(reso) + (origins[0])
			y := float64(i/stride)*(reso) + (origins[1])
			point := Point{
				X: x,
				Y: y,
			}
			g.Ox = append(g.Ox, x)
			g.Oy = append(g.Oy, y)
			g.OList = append(g.OList, point)
		} else {
			insideWall = false
		}
	}
	log.Print("complete loading map", mapFile)
	return nil
}

// xyどちらかの座標を受取、idを返す
func (g Grid) xyIndex(p float64, minp int) int {
	return int(math.Round((p - float64(minp)) / g.Reso))
}

func heuristic(n1, n2 *Node) float64 {
	w := 1.0
	d := w * math.Hypot(float64(n1.Ix)-float64(n2.Ix), float64(n1.Iy)-float64(n2.Iy))
	return d
}

// nodeを受取idを返す
func (g Grid) gridIndex(n *Node) int {
	return (n.Iy-g.MinY)*g.XWidth + (n.Ix - g.MinX)
}

// x,yどちらかのindexを受取座標を返す
func (g Grid) gridPos(index, minP int) float64 {
	pos := float64(index)*g.Reso + float64(minP)
	return pos
}

func (g Grid) AstarPlan(sx, sy, gx, gy float64, hidden []int) (rx, ry []float64, Notfail bool) {
	nstart := NewNodeG(g.xyIndex(sx, g.MinX), g.xyIndex(sy, g.MinY), 0.0, -1)
	ngoal := NewNodeG(g.xyIndex(gx, g.MinX), g.xyIndex(gy, g.MinY), 0.0, -1)

	open_set := make(map[int]*Node)
	close_set := make(map[int]*Node)
	open_set[g.gridIndex(nstart)] = nstart

	for {
		if len(open_set) == 0 {
			log.Print("open set is empty..")
			var failX []float64
			var failY []float64
			return failX, failY, false
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
			break
		}

		delete(open_set, cId)

		close_set[cId] = current

		var nId int
		var node *Node
		motion := [8][3]float64{{1.0, 0, 1.0}, {0, 1.0, 1.0}, {-1.0, 0, 1.0}, {0, -1.0, 1.0}, {-1.0, -1.0, math.Sqrt(3)}, {-1.0, 1.0, math.Sqrt(3)}, {1.0, -1.0, math.Sqrt(3)}, {1.0, 1.0, math.Sqrt(3)}}
		for _, v := range motion {
			node = NewNodeG(current.Ix+int(v[0]), current.Iy+int(v[1]), current.Cost+v[2], cId)
			nId = g.gridIndex(node)
		}

		if !g.VerifyGridP(g.gridIndex(node), hidden) {
			continue
		}

		if _, ok := close_set[nId]; ok {
			continue
		}

		if _, ok := open_set[nId]; !ok {
			open_set[nId] = node
		}
	}
	rx, ry = g.finalPath(ngoal, close_set)
	return rx, ry, true
}

// 最後に経路の順番にする
func (g Grid) finalPath(ngoal *Node, closeSet map[int]*Node) (rx, ry []float64) {
	rx = append(rx, g.gridPos(ngoal.Ix, g.MinX))
	ry = append(ry, g.gridPos(ngoal.Iy, g.MinY))

	pind := ngoal.Pind
	for pind != -1 {
		n := closeSet[pind]
		rx = append(rx, g.gridPos(n.Ix, g.MinX))
		ry = append(ry, g.gridPos(n.Iy, g.MinY))
	}
	return rx, ry
}

type Node struct {
	Index int
	Ix    int
	Iy    int

	X    float64
	Y    float64
	Cost float64
	Pind int

	Obj bool //障害物ならtrue
}

func NewNode(index, ix, iy int, x, y float64) *Node {
	n := new(Node)
	n.Index = index
	n.Ix = ix
	n.Iy = iy

	n.X = x
	n.Y = y

	return n
}

func NewNodeG(x, y int, cost float64, pind int) *Node {
	n := new(Node)
	n.Ix = x
	n.Iy = y
	n.Cost = cost
	n.Pind = pind
	return n
}

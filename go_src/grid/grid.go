package grid

import (
	"fmt"
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

type GridState int

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

	ObjMap [][]bool //if object, it is true

	Nodes map[int]*Node
}

type Point struct {
	X float64
	Y float64
}

// for testing
func NewGridNo() *Grid {
	g := new(Grid)
	return g
}

// Grid Constructor
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

	g.XWidth = int(math.Round(float64(g.MaxX) - float64(g.MinX)/g.Reso))
	g.YWidth = int(math.Round(float64(g.MaxY) - float64(g.MinY)/g.Reso))

	g.ObjMap = make([][]bool, g.XWidth)
	for i := 0; i < g.XWidth; i++ {
		g.ObjMap[i] = make([]bool, g.YWidth)
	}
	g.Nodes = make(map[int]*Node)

	var ind int
	for iy := 0; iy < g.YWidth; iy++ {
		y := g.CalcXYPosition(iy, g.MinY)
		for ix := 0; ix < g.XWidth; ix++ {
			x := g.CalcXYPosition(ix, g.MinX)

			ind = iy*g.XWidth + ix
			g.Nodes[ind] = NewNode(ind, ix, iy, x, y)
			for _, ip := range g.OList {
				d := math.Hypot(ip.X-float64(x), ip.Y-float64(y))
				if d <= rr {
					g.ObjMap[ix][iy] = true
					g.Nodes[ind].Obj = true
					break
				}
			}
			if g.Nodes[ind].Obj {
				fmt.Print("*")
			} else {
				fmt.Print(".")
			}
			if ix == g.XWidth-1 {
				fmt.Println()
			}
		}
	}
	log.Print("complete calculate objmap")
	log.Print(`minX, maxX, minY, maxY
		xwidth, ywidth`)
	log.Print(g.MinX, g.MaxX, g.MinY, g.MaxY)
	log.Print(g.XWidth, g.YWidth)
	log.Print("max index is ", ind)
}

// xyどちらかの座標を受取、idを返す
func (g Grid) xyIndex(p float64, minp int) int {
	return int(math.Round((p - float64(minp)) / g.Reso))
}

// get x or y position from thats index
func (g Grid) CalcXYPosition(index, minP int) float64 {
	pos := float64(index)*g.Reso + float64(minP)
	return pos
}

//get positoin from node index
func (g Grid) CalcPosition(index int) (float64, float64) {
	// px := g.MinX + math.Round(float64(index%g.XWidth))*g.Reso
	// py := g.MinY + math.Round(float64(index/g.XWidth))*g.Reso
	px := g.Nodes[index].X
	py := g.Nodes[index].Y
	return px, py
}

// だめなグリッドを任意で追加site verify grid
func (g Grid) VerifyGridP(index int, hidden []int) bool {
	if tools.CheckSameCom(index, hidden) {
		return false
	}
	return g.VerifyGrid(index)
}

// return true if index grid is passable
func (g Grid) VerifyGrid(index int) bool {
	if index > g.XWidth*g.YWidth-1 {
		// log.Print("index ", index, " is overflow")
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

	if g.Nodes[index].Obj {
		return false
	}
	return true
}

//from position to return node index
func (g Grid) PosToGrid(x, y float64) int {
	ix := g.xyIndex(x, g.MinX)
	iy := g.xyIndex(y, g.MinY)
	return iy*g.XWidth + ix
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

// read image file of ROS format
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

// each Grid
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

// grid constructor
func NewNode(index, ix, iy int, x, y float64) *Node {
	n := new(Node)
	n.Index = index
	n.Ix = ix
	n.Iy = iy

	n.X = x
	n.Y = y

	return n
}

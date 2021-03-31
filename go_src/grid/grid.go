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

	g.ObjMap = make([][]bool, g.XWidth*g.YWidth)
	for ix := 0; ix < g.XWidth; ix++ {
		x := g.CalcXYPosition(ix, g.MinX)
		for iy := 0; iy < g.YWidth; iy++ {
			y := g.CalcXYPosition(iy, g.MinY)
			for _, ip := range g.OList {
				d := math.Hypot(ip.X-float64(x), ip.Y-float64(y))
				if d < rr {
					g.ObjMap[ix][iy] = true
					break
				}
			}
		}
	}
	log.Print("complete calculate objmap")
}

func (g Grid) CalcXYPosition(index, minP int) int {
	pos := index*int(g.Reso) + minP
	return pos
}

func (g Grid) CalcPosition(index int) (int, int) {
	px := g.MinX + int(math.Round(float64(index%g.XWidth))*g.Reso)
	py := g.MinY + int(math.Round(float64(index/g.XWidth))*g.Reso)
	return px, py
}

func (g Grid) VerifyGrid(index int) bool {
	px, py := g.CalcPosition(index)

	if px < g.MinX {
		return false
	} else if py < g.MinY {
		return false
	} else if px >= g.MaxX {
		return false
	} else if py >= g.MaxY {
		return false
	}

	if g.ObjMap[px][py] {
		return false
	}
	return true
}

func (g Grid) PosToGrid(x, y float64) int {
	return (int(math.Round(y))-g.MinY)*g.XWidth + (int(math.Round(x)) - g.MinX)
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

	//log.Print(imageData.Bounds())

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

type Node struct {
	Index uint32
	X     uint32
	Y     uint32
	Cost  float64
	Pind  float64
}

package grid

import (
	"image"
	"image/color"
	"log"
	"os"

	msg "example.com/msg"
	_ "github.com/jbuchbinder/gopnm"
)

type Grid struct {
	Reso float64

	MinX int
	MaxX int
	MinY int
	MaxY int

	XWidth uint32
	YWidth uint32

	ObjMap []Point
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

func (g *Grid) ReadMapImage(yamlFile, mapFile string) {
	mapConfig := msg.ReadImageYaml(yamlFile)
	reso := mapConfig.Resolution
	origins := mapConfig.Origin
	log.Print(reso, origins)

	file, err := os.Open(mapFile)
	if err != nil {
		log.Print(err)
	}
	defer file.Close()

	imageData, imageType, err := image.Decode(file)
	if err != nil {
		log.Print(err)
	}

	log.Print(imageData.Bounds())
	log.Print(imageType)

	g.ObjMap = nil
	// g.MaxX = imageData.Bounds().Max.X
	// g.MaxY = imageData.Bounds().Max.Y
	// g.MinX = imageData.Bounds().Min.X
	// g.MinY = imageData.Bounds().Min.Y

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
			point := Point{
				X: float64(i%stride)*(reso) + (origins[0]),
				Y: float64(i/stride)*(reso) + (origins[1]),
			}
			g.ObjMap = append(g.ObjMap, point)
		} else {
			insideWall = false
		}
	}
	log.Print(len(g.ObjMap))
	log.Print("complete loading map")

}

type Node struct {
	Index uint32
	X     uint32
	Y     uint32
	Cost  float64
	Pind  float64
}

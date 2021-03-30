package main

import (
	gridMod "example.com/grid"
)

func main() {
	g := gridMod.NewGridNo()
	yamlFile := "../../map/trusco_map_edited.yaml"
	mapFile := "../../map/trusco_map_edited.pgm"
	g.ReadMapImage(yamlFile, mapFile)
}

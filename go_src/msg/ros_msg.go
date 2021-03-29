package msg

import (
	"math"
)

type ROS_header struct {
	Seq      uint32    `json:"seq"`
	Stamp    TimeStamp `json:"stamp"`
	Frame_id string    `json:"frame_id"`
}

type TimeStamp struct {
	Secs  uint32 `json:"secs"`
	Nsecs uint32 `json:"nsecs"`
}

func (t TimeStamp) CalcUnix() float64 {
	return float64(t.Secs) + float64(t.Nsecs*uint32(math.Pow10(-9)))
}

type Point struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
	Z float64 `json:"z"`
}

func (p Point) Distance(o Point) float64 {
	return math.Hypot(p.X-o.X, p.Y-o.Y)
}

type Quaternion struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
	Z float64 `json:"z"`
	W float64 `json:"w"`
}

type Pose struct {
	Position    Point      `json:"position"`
	Orientation Quaternion `json:"orientation"`
}

type PoseStamp struct {
	Header ROS_header `json:"header"`
	Pose   Pose       `json:"pose"`
}

type Path struct {
	Header ROS_header  `json:"header"`
	Poses  []PoseStamp `json:"poses"`
}

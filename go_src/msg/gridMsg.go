package msg

type Stop struct {
	Header ROS_header `json:"header"`
	From   TimeStamp  `json:"from"`
	To     TimeStamp  `json:"to"`
}

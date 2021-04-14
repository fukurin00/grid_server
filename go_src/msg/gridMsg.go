package msg

type Stop struct {
	Header ROS_header `json:"header"`
	From   TimeStamp  `json:"fromT"`
	To     TimeStamp  `json:"toT"`
}

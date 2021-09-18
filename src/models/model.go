package models

type ErrorMessage struct {
	Message string `json:"message"`
}

type Rect struct {
	TopX    float64
	TopY    float64
	BottomX float64
	BottomY float64
}

type Point struct {
	X float64
	Y float64
}

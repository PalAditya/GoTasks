package utils

import (
	"InShorts/src/models"
)

var boundingMap map[string]models.Rect

func init() {
	// May not be 100% accurate. Mostly as a POC of caching
	//This should be created as a list of Rects
	boundingMap = make(map[string]models.Rect)
	boundingMap["West Bengal"] = models.Rect{23.7, 87.0, 22.3, 88.4}
	boundingMap["Maharashtra"] = models.Rect{21.1, 74.0, 19.0, 77.0}
}

func IsInBoundingBox(p models.Point) string {
	for state, bb := range boundingMap {
		if bb.TopX <= p.X && p.X <= bb.BottomX && bb.BottomY <= p.Y && p.Y <= bb.TopY {
			return state
		}
	}
	return ""
}

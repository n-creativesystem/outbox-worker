package utils

import (
	"time"

	"github.com/doug-martin/goqu/v9"
)

var JST = time.FixedZone("Asia/Tokyo", 9*60*60)

func init() {
	time.Local = JST
	goqu.SetTimeLocation(JST)
}

func NowInJST() time.Time {
	return time.Now().In(JST)
}

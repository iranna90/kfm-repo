package dailymilk

import "time"

type Balance struct {
	Id        int64
	Amount    int64
	Modified  time.Time
	PersonRef int64
}

package payment

import "time"

type Payment struct {
	Id        int64
	amount    int64
	PaidTo    string
	Day       time.Time
	PersonRef int64
}

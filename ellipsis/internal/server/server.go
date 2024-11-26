package server

import (
	"encoding/csv"
	"os"
	"strconv"
	"time"

	dec "github.com/shopspring/decimal"
)

const (
	dateFmt        = "2006-01-02 15:04:05"
	approxSecToDay = 0.00001 // 1s is approximately 0.00001 (1.15741e-5) days
)

type Fill struct {
	Time           time.Time
	Direction      int
	Price          dec.Decimal
	Quantity       dec.Decimal
	SequenceNumber uint64
}

type Server struct {
	fills []*Fill
}

func NewServer() *Server {
	file := must(os.Open("./trades.csv"))
	defer file.Close()
	records := must(csv.NewReader(file).ReadAll())

	// successful init is required, so we panic if anything goes wrong
	var fills []*Fill
	for _, record := range records[1:] {
		fill := &Fill{
			Time:           must(time.Parse(dateFmt, record[0])).UTC(),
			Direction:      must(strconv.Atoi(record[1])),
			Price:          must(dec.NewFromString(record[2])),
			Quantity:       must(dec.NewFromString(record[3])),
			SequenceNumber: must(strconv.ParseUint(record[4], 10, 64)),
		}
		fills = append(fills, fill)
	}

	return &Server{
		fills: fills,
	}
}

func (s *Server) GetFillsAPI(startTsInSeconds int64, endTsInSeconds int64) []*Fill {
	startTime := time.Unix(startTsInSeconds, 0).UTC()
	endTime := time.Unix(endTsInSeconds, 0).UTC()

	// fetching 1 day's worth of data should take around 1s (0.864s)
	intervalLen := float64(endTsInSeconds - startTsInSeconds)
	sleepTime := time.Duration(intervalLen * float64(time.Second) * approxSecToDay)
	time.Sleep(sleepTime)

	var result []*Fill
	for _, fill := range s.fills {
		if fill.Time.After(startTime) && fill.Time.Compare(endTime) <= 0 {
			result = append(result, fill)
		}
	}

	return result
}

func must[T any](ret T, err error) T {
	if err != nil {
		panic(err)
	}
	return ret
}

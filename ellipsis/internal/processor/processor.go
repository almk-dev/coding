package processor

import (
	"ellipsis/internal/server"
	"fmt"
	"strconv"
	"strings"

	dec "github.com/shopspring/decimal"
)

type QueryType int

const (
	Count QueryType = iota
	Buys
	Sells
	Vol
)

var queryTypeMap = map[rune]QueryType{
	'C': Count,
	'B': Buys,
	'S': Sells,
	'V': Vol,
}

// we use an empty struct instead of bool for memory efficiency
type Set map[uint64]Exists
type Exists struct{}

type Processor struct {
	server *server.Server
}

type Opts struct {
	Server *server.Server
}

func NewProcessor(opts Opts) *Processor {
	return &Processor{
		server: opts.Server,
	}
}

func (p *Processor) ProcessQuery(query string) error {
	queryType, startTsInSeconds, endTsInSeconds, err := p.parseQuery(query)
	if err != nil {
		return fmt.Errorf("failed to parse query: %w", err)
	}
	fills := p.server.GetFillsAPI(*startTsInSeconds, *endTsInSeconds)

	switch *queryType {
	case Count:
		p.ProcessCount(fills)
	case Buys:
		p.ProcessBuys(fills)
	case Sells:
		p.ProcessSells(fills)
	case Vol:
		p.ProcessVol(fills)
	}

	return nil
}

func (p *Processor) ProcessCount(fills []*server.Fill) {
	var totalCount int
	seen := make(map[uint64]struct{})
	for _, fill := range fills {
		if _, ok := seen[fill.SequenceNumber]; !ok {
			seen[fill.SequenceNumber] = Exists{}
			totalCount++
		}
	}

	fmt.Println(totalCount)
}

func (p *Processor) ProcessBuys(fills []*server.Fill) {
	var totalBuys int
	seen := make(Set)
	for _, fill := range fills {
		if _, ok := seen[fill.SequenceNumber]; !ok && fill.Direction > 0 {
			seen[fill.SequenceNumber] = Exists{}
			totalBuys++
		}
	}

	fmt.Println(totalBuys)
}

func (p *Processor) ProcessSells(fills []*server.Fill) {
	var totalSells int
	seen := make(map[uint64]struct{})
	for _, fill := range fills {
		if _, ok := seen[fill.SequenceNumber]; !ok && fill.Direction < 0 {
			seen[fill.SequenceNumber] = Exists{}
			totalSells++
		}
	}

	fmt.Println(totalSells)
}

func (p *Processor) ProcessVol(fills []*server.Fill) {
	var totalVol dec.Decimal
	for _, fill := range fills {
		totalVol = totalVol.Add(fill.Price.Mul(fill.Quantity))
	}

	fmt.Println(totalVol)
}

func (p *Processor) parseQuery(query string) (*QueryType, *int64, *int64, error) {
	fields := strings.Fields(query)
	if len(fields) != 3 {
		return nil, nil, nil, fmt.Errorf("invalid query witn %d fields: %s", len(fields), query)
	}

	if len(fields[0]) != 1 {
		return nil, nil, nil, fmt.Errorf("invalid query type of len %d: %s", len(fields[0]), fields[0])
	}
	queryType, ok := queryTypeMap[rune(fields[0][0])]
	if !ok {
		return nil, nil, nil, fmt.Errorf("invalid query type: %s", fields[0])
	}
	startTsInSeconds, err := strconv.ParseInt(fields[1], 10, 64)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("invalid start timestamp: %s", fields[1])
	}
	endTsInSeconds, err := strconv.ParseInt(fields[2], 10, 64)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("invalid end timestamp: %s", fields[2])
	}

	return &queryType, &startTsInSeconds, &endTsInSeconds, nil
}

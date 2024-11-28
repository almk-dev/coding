package processor

import (
	"ellipsis/internal/server"
	"fmt"
	"sort"
	"strconv"
	"strings"

	dec "github.com/shopspring/decimal"
)

// we use an empty struct instead of bool for memory efficiency
type set map[uint64]setFlag
type setFlag struct{}

var exists = setFlag{}

type queryType int

const (
	count queryType = iota
	buys
	sells
	vol
)

var queryTypeMap = map[rune]queryType{
	'C': count,
	'B': buys,
	'S': sells,
	'V': vol,
}

type interval struct {
	start int64
	end   int64
}

type data struct {
	count int
	buys  int
	sells int
	vol   dec.Decimal
}

type Processor struct {
	server *server.Server
	cache  map[interval]data
}

type Opts struct {
	Server *server.Server
}

func NewProcessor(opts Opts) *Processor {
	return &Processor{
		server: opts.Server,
		cache:  make(map[interval]data),
	}
}

func (p *Processor) ProcessQuery(query string) error {
	queryType, startTsInSeconds, endTsInSeconds, err := p.parseQuery(query)
	if err != nil {
		return fmt.Errorf("failed to parse query: %w", err)
	}

	cacheHitIntervals, queryIntervals, staleIntervals := p.processIntervals(*startTsInSeconds, *endTsInSeconds)
	p.processAll(*queryType, cacheHitIntervals, queryIntervals, staleIntervals)

	return nil
}

func (p *Processor) processIntervals(
	startTsInSeconds int64,
	endTsInSeconds int64,
) ([]interval, []interval, []interval) {
	var cacheIntervals, queryIntervals []interval
	for k := range p.cache {
		cacheIntervals = append(cacheIntervals, k)
	}

	var cacheHitIntervals []interval
	for _, cacheInterval := range cacheIntervals {
		if (cacheInterval.start < startTsInSeconds && cacheInterval.end <= startTsInSeconds) ||
			(cacheInterval.start >= endTsInSeconds && cacheInterval.end > endTsInSeconds) {
			continue
		}
		cacheHitIntervals = append(cacheHitIntervals, cacheInterval)
	}

	sort.Slice(cacheHitIntervals, func(i, j int) bool {
		return cacheHitIntervals[i].start < cacheHitIntervals[j].start
	})

	var last int64 = startTsInSeconds
	var staleIntervals []interval
	for _, cacheHit := range cacheHitIntervals {
		if cacheHit.start < startTsInSeconds && cacheHit.end > endTsInSeconds {
			staleIntervals = append(staleIntervals, cacheHit)
			cacheHitIntervals = cacheHitIntervals[1:]
			continue
		}
		if cacheHit.start < startTsInSeconds {
			staleIntervals = append(staleIntervals, cacheHit)
			queryIntervals = append(queryIntervals, interval{start: startTsInSeconds, end: cacheHit.end})
			cacheHitIntervals = cacheHitIntervals[1:]
			last = cacheHit.end
			continue
		}
		if cacheHit.end > endTsInSeconds {
			staleIntervals = append(staleIntervals, cacheHit)
			if cacheHit.start > last {
				queryIntervals = append(queryIntervals, interval{start: last, end: cacheHit.start})
			}
			cacheHitIntervals = cacheHitIntervals[:len(cacheHitIntervals)-1]
			last = cacheHit.start
			continue
		}
		if cacheHit.start > last {
			queryIntervals = append(queryIntervals, interval{start: last, end: cacheHit.start})
		}
		last = cacheHit.end
	}

	if last < endTsInSeconds {
		queryIntervals = append(queryIntervals, interval{start: last, end: endTsInSeconds})
	}
	if len(queryIntervals) == 0 && len(cacheHitIntervals) == 0 {
		queryIntervals = append(queryIntervals, interval{start: startTsInSeconds, end: endTsInSeconds})
	}

	return cacheHitIntervals, queryIntervals, staleIntervals
}

func (p *Processor) processAll(qt queryType, cacheHitIntervals, queryIntervals, staleIntervals []interval) {
	totalData := data{}
	for _, interval := range cacheHitIntervals {
		totalData.count += p.cache[interval].count
		totalData.buys += p.cache[interval].buys
		totalData.sells += p.cache[interval].sells
		totalData.vol = totalData.vol.Add(p.cache[interval].vol)
	}

	for _, interval := range queryIntervals {
		queryData := data{}
		fills := p.server.GetFillsAPI(interval.start, interval.end)
		seenCount, seenBuys, seenSells := make(set), make(set), make(set)

		for _, fill := range fills {
			if _, ok := seenCount[fill.SequenceNumber]; !ok {
				seenCount[fill.SequenceNumber] = exists
				queryData.count++
			}
			if _, ok := seenBuys[fill.SequenceNumber]; !ok && fill.Direction > 0 {
				seenBuys[fill.SequenceNumber] = exists
				queryData.buys++
			}
			if _, ok := seenSells[fill.SequenceNumber]; !ok && fill.Direction < 0 {
				seenSells[fill.SequenceNumber] = exists
				queryData.sells++
			}
			queryData.vol = queryData.vol.Add(fill.Price.Mul(fill.Quantity))
		}

		totalData.count += queryData.count
		totalData.buys += queryData.buys
		totalData.sells += queryData.sells
		totalData.vol = totalData.vol.Add(queryData.vol)

		p.updateCache(interval, queryData, staleIntervals)
	}

	p.writeOutput(qt, totalData)
}

func (p *Processor) parseQuery(query string) (*queryType, *int64, *int64, error) {
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

func (p *Processor) updateCache(
	newInterval interval,
	newData data,
	staleIntervals []interval,
) {
	for _, stale := range staleIntervals {
		var refreshInterval interval
		if stale.start < newInterval.start && stale.end > newInterval.end {
			delete(p.cache, stale)
			break
		} else if stale.start < newInterval.start && stale.end == newInterval.end {
			refreshInterval = interval{start: stale.start, end: newInterval.start}
		} else if stale.start == newInterval.start && stale.end > newInterval.end {
			refreshInterval = interval{start: newInterval.end, end: stale.end}
		} else {
			continue
		}

		p.cache[refreshInterval] = data{
			count: p.cache[stale].count - newData.count,
			buys:  p.cache[stale].buys - newData.buys,
			sells: p.cache[stale].sells - newData.sells,
			vol:   p.cache[stale].vol.Sub(newData.vol),
		}
		delete(p.cache, stale)
	}

	p.cache[newInterval] = newData
}

func (p *Processor) writeOutput(qt queryType, totalData data) {
	switch qt {
	case count:
		fmt.Println(totalData.count)
	case buys:
		fmt.Println(totalData.buys)
	case sells:
		fmt.Println(totalData.sells)
	case vol:
		fmt.Println(totalData.vol)
	}
}

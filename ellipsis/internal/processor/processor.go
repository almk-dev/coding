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

type cache struct {
	countCache map[interval]int
	buysCache  map[interval]int
	sellsCache map[interval]int
	volCache   map[interval]dec.Decimal
}

type Processor struct {
	server *server.Server
	cache  *cache
}

type Opts struct {
	Server *server.Server
}

func NewProcessor(opts Opts) *Processor {
	return &Processor{
		server: opts.Server,
		cache: &cache{
			countCache: make(map[interval]int),
			buysCache:  make(map[interval]int),
			sellsCache: make(map[interval]int),
			volCache:   make(map[interval]dec.Decimal),
		},
	}
}

func (p *Processor) ProcessQuery(query string) error {
	queryType, startTsInSeconds, endTsInSeconds, err := p.parseQuery(query)
	if err != nil {
		return fmt.Errorf("failed to parse query: %w", err)
	}
	cacheHitIntervals, queryIntervals, staleIntervals := p.processIntervals(*queryType, *startTsInSeconds, *endTsInSeconds)
	fmt.Println("type:", query[:1], "start:", *startTsInSeconds, "end:", *endTsInSeconds)
	fmt.Println("hit:", cacheHitIntervals, "query:", queryIntervals, "stale:", staleIntervals)

	switch *queryType {
	case count:
		p.processCount(cacheHitIntervals, queryIntervals, staleIntervals)
	case buys:
		p.processBuys(cacheHitIntervals, queryIntervals, staleIntervals)
	case sells:
		p.processSells(cacheHitIntervals, queryIntervals, staleIntervals)
	case vol:
		p.processVol(cacheHitIntervals, queryIntervals, staleIntervals)
	}

	return nil
}

func (p *Processor) processIntervals(
	qt queryType,
	startTsInSeconds int64,
	endTsInSeconds int64,
) ([]interval, []interval, []interval) {
	var cacheIntervals, queryIntervals []interval
	switch qt {
	case count:
		for k := range p.cache.countCache {
			cacheIntervals = append(cacheIntervals, k)
		}
	case buys:
		for k := range p.cache.buysCache {
			cacheIntervals = append(cacheIntervals, k)
		}
	case sells:
		for k := range p.cache.sellsCache {
			cacheIntervals = append(cacheIntervals, k)
		}
	case vol:
		for k := range p.cache.volCache {
			cacheIntervals = append(cacheIntervals, k)
		}
	}

	var cacheHitIntervals []interval
	for _, cacheInterval := range cacheIntervals {
		if (cacheInterval.start < startTsInSeconds && cacheInterval.end <= endTsInSeconds) ||
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
			break
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

func (p *Processor) processCount(cacheHitIntervals, queryIntervals, staleIntervals []interval) {
	var totalCount int
	for _, interval := range cacheHitIntervals {
		totalCount += p.cache.countCache[interval]
	}

	for _, interval := range queryIntervals {
		fills := p.server.GetFillsAPI(interval.start, interval.end)
		seen := make(map[uint64]struct{})
		for _, fill := range fills {
			if _, ok := seen[fill.SequenceNumber]; !ok {
				seen[fill.SequenceNumber] = exists
				totalCount++
			}
		}
		p.updateCache(count, interval, &totalCount, nil, staleIntervals)
	}

	fmt.Println(totalCount)
}

func (p *Processor) processBuys(cacheHitIntervals, queryIntervals, staleIntervals []interval) {
	var totalBuys int
	for _, interval := range cacheHitIntervals {
		totalBuys += p.cache.buysCache[interval]
	}

	fmt.Println(totalBuys)

	for _, interval := range queryIntervals {
		fills := p.server.GetFillsAPI(interval.start, interval.end)
		seen := make(set)
		for _, fill := range fills {
			if _, ok := seen[fill.SequenceNumber]; !ok && fill.Direction > 0 {
				seen[fill.SequenceNumber] = exists
				totalBuys++
			}
		}
		p.updateCache(buys, interval, &totalBuys, nil, staleIntervals)
		fmt.Println(totalBuys)
	}

	fmt.Println(totalBuys)
}

func (p *Processor) processSells(cacheHitIntervals, queryIntervals, staleIntervals []interval) {
	var totalSells int
	for _, interval := range cacheHitIntervals {
		totalSells += p.cache.sellsCache[interval]
	}

	for _, interval := range queryIntervals {
		fills := p.server.GetFillsAPI(interval.start, interval.end)
		seen := make(map[uint64]struct{})
		for _, fill := range fills {
			if _, ok := seen[fill.SequenceNumber]; !ok && fill.Direction < 0 {
				seen[fill.SequenceNumber] = exists
				totalSells++
			}
		}
		p.updateCache(sells, interval, &totalSells, nil, staleIntervals)
	}

	fmt.Println(totalSells)
}

func (p *Processor) processVol(cacheHitIntervals, queryIntervals, staleIntervals []interval) {
	var totalVol dec.Decimal
	for _, interval := range cacheHitIntervals {
		totalVol = totalVol.Add(p.cache.volCache[interval])
	}

	for _, interval := range queryIntervals {
		fills := p.server.GetFillsAPI(interval.start, interval.end)
		for _, fill := range fills {
			totalVol = totalVol.Add(fill.Price.Mul(fill.Quantity))
		}
		p.updateCache(vol, interval, nil, &totalVol, staleIntervals)
	}

	fmt.Println(totalVol)
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
	qt queryType,
	newInterval interval,
	totalInt *int,
	totalDec *dec.Decimal,
	staleIntervals []interval,
) {
	for _, stale := range staleIntervals {
		var refreshInterval interval
		if stale.start < newInterval.start && stale.end > newInterval.end {
			delete(p.cache.countCache, stale)
			break
		} else if stale.start < newInterval.start && stale.end == newInterval.end {
			refreshInterval = interval{start: stale.start, end: newInterval.start}
		} else if stale.start == newInterval.start && stale.end > newInterval.end {
			refreshInterval = interval{start: newInterval.end, end: stale.end}
		} else {
			continue
		}

		switch qt {
		case count:
			p.cache.countCache[refreshInterval] = p.cache.countCache[stale] - *totalInt
			delete(p.cache.countCache, stale)
		case buys:
			p.cache.buysCache[refreshInterval] = p.cache.buysCache[stale] - *totalInt
			delete(p.cache.buysCache, stale)
		case sells:
			p.cache.sellsCache[refreshInterval] = p.cache.sellsCache[stale] - *totalInt
			delete(p.cache.sellsCache, stale)
		case vol:
			p.cache.volCache[refreshInterval] = p.cache.volCache[stale].Sub(*totalDec)
			delete(p.cache.volCache, stale)
		}
	}

	switch qt {
	case count:
		p.cache.countCache[newInterval] = *totalInt
	case buys:
		p.cache.buysCache[newInterval] = *totalInt
	case sells:
		p.cache.sellsCache[newInterval] = *totalInt
	case vol:
		p.cache.volCache[newInterval] = *totalDec
	}
}

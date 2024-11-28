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
	cacheHitIntervals, queryIntervals := p.processIntervals(*queryType, *startTsInSeconds, *endTsInSeconds)

	switch *queryType {
	case count:
		p.processCount(cacheHitIntervals, queryIntervals)
	case buys:
		p.processBuys(cacheHitIntervals, queryIntervals)
	case sells:
		p.processSells(cacheHitIntervals, queryIntervals)
	case vol:
		p.processVol(cacheHitIntervals, queryIntervals)
	}

	return nil
}

func (p *Processor) processIntervals(
	qt queryType,
	startTsInSeconds int64,
	endTsInSeconds int64,
) ([]interval, []interval) {
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

	var cacheHits []interval
	for _, cacheInterval := range cacheIntervals {
		if cacheInterval.start >= startTsInSeconds && cacheInterval.end <= endTsInSeconds {
			cacheHits = append(cacheHits, cacheInterval)
		} else if startTsInSeconds < cacheInterval.start && endTsInSeconds > cacheInterval.start && endTsInSeconds <= cacheInterval.end {
			cacheHits = append(cacheHits, cacheInterval)
		} else if startTsInSeconds >= cacheInterval.start && startTsInSeconds < cacheInterval.end && endTsInSeconds > cacheInterval.end {
			queryIntervals = append(queryIntervals, interval{start: cacheInterval.start, end: endTsInSeconds})
		}
	}

	sort.Slice(cacheHits, func(i, j int) bool {
		return cacheHits[i].start < cacheHits[j].start
	})

	var last int64
	for _, cacheHit := range cacheHits {
		if cacheHit.start < startTsInSeconds {
			queryIntervals = append(queryIntervals, interval{start: startTsInSeconds, end: cacheHit.end})
			cacheHits = cacheHits[1:]
			last = cacheHit.end
			continue
		}
		if cacheHit.end > endTsInSeconds {
			queryIntervals = append(queryIntervals, interval{start: last, end: endTsInSeconds})
			cacheHits = cacheHits[:len(cacheHits)-1]
			continue
		}
		if cacheHit.start > last {
			queryIntervals = append(queryIntervals, interval{start: last, end: cacheHit.start})
		}
		last = cacheHit.end
	}

	if len(queryIntervals) == 0 && len(cacheHits) == 0 {
		queryIntervals = append(queryIntervals, interval{start: startTsInSeconds, end: endTsInSeconds})
	}

	return cacheHits, queryIntervals
}

func (p *Processor) processCount(cacheHitIntervals, queryIntervals []interval) {
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
		p.updateCache(count, interval, &totalCount, nil)
	}

	fmt.Println(totalCount)
}

func (p *Processor) processBuys(cacheHitIntervals, queryIntervals []interval) {
	var totalBuys int
	for _, interval := range cacheHitIntervals {
		totalBuys += p.cache.buysCache[interval]
	}

	for _, interval := range queryIntervals {
		fills := p.server.GetFillsAPI(interval.start, interval.end)
		seen := make(set)
		for _, fill := range fills {
			if _, ok := seen[fill.SequenceNumber]; !ok && fill.Direction > 0 {
				seen[fill.SequenceNumber] = exists
				totalBuys++
			}
		}
		p.updateCache(buys, interval, &totalBuys, nil)
	}

	fmt.Println(totalBuys)
}

func (p *Processor) processSells(cacheHitIntervals, queryIntervals []interval) {
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
		p.updateCache(sells, interval, &totalSells, nil)
	}

	fmt.Println(totalSells)
}

func (p *Processor) processVol(cacheHitIntervals, queryIntervals []interval) {
	var totalVol dec.Decimal
	for _, interval := range cacheHitIntervals {
		totalVol = totalVol.Add(p.cache.volCache[interval])
	}

	for _, interval := range queryIntervals {
		fills := p.server.GetFillsAPI(interval.start, interval.end)
		for _, fill := range fills {
			totalVol = totalVol.Add(fill.Price.Mul(fill.Quantity))
		}
		p.updateCache(vol, interval, nil, &totalVol)
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
	interval interval,
	totalInt *int,
	totalDec *dec.Decimal,
) {
	switch qt {
	case count:
		p.cache.countCache[interval] = *totalInt
	case buys:
		p.cache.buysCache[interval] = *totalInt
	case sells:
		p.cache.sellsCache[interval] = *totalInt
	case vol:
		p.cache.volCache[interval] = *totalDec
	}
}

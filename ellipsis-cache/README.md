### Running
Assuming that `go` `1.23.1` is installed, the program can be run like so:
```sh
$ cat input.txt | go run .
```
However, any `go` version after `1.18` (post-generics) should be compatible–just update the version in the `go.mod` file accordingly. 


### Language choice
The program is implemented in `go`, since that's what I have been using the most in recent months. The provided `rust` server code was ported over as closely as possible while still being true to idiomatic `go`.

I considered following the "DO NOT MODIFY THIS FILE" warning, but ultimately decided to port the code because:

- I was told the assignment was language agnostic
- The instructions emphasize "production-grade code," so it was better to stick to what I had more professional experience in
- A hybrid solution using a linked libary or FFI is brittle and too dependent on computer architecture, and thus is a risky choice

### Design
#### Porting to `go`
I decided to keep the `Processor` and `Server` design so the `go` port would be easy to recognize and follow. Some minor differences to note:

- The main entry file (`ellipsis.go`) was separated from the component "services," which were placed in `internal/` as is the norm in `go` projects
    - The distinction between `internal`, `external`, `pkg`, etc. is not very useful in this small codebase, however
- `Server` was made a dependency of `Processor`
    - This submission is single-threaded and single-server, but having a "pool" of servers as a dependency makes for an easy design pattern if we want the processor to distribute its API calls across available servers
- Initial loading of trades/fills makes use of `must`, which panics if anything goes wrong
    - In an actual deployment, we may want to check each step's error values explicitly, but since the CSV load is a necessary precursor to anything else, I felt it was simpler to use `must` instead 

#### Organization and minor details
Some helpers and structs were added to keep the top-level logic easy to follow:
- Both `Processor` and `Server` make use of `New*` and `Opts` to make it easy to inialize in `main`
- `parseQuery` handles breaking up each query string into more usable tokens, and keeps `ProcessQuery` less noisey
- We create an enum of `queryType` runes (single char strings in `go`) to easily route queries
- We define alias of `struct{}` to be `setFlag`/`exists` in favor of using `bool`, as `go` implements it as 0 bytes compared to the 1-byte `bool`, which provides marginal memory usage improvements
- We use the third-party `shopspring/decimal` library to avoid implementing our own decimal functionality

#### Cache design
The core premise is that whenever we have a cache miss, we compute all four queries and cache the final totals, allowing each query type to cross-reference previous results from a different query type.

Some other options/tradeoffs that were considered:
- Caching the entire fill data (whole rows) after each API call was a possibility, but that is very memory inefficient and is not much better than calling the entire CSV range and storing it into the cache in order to bypass the sleep calls after a one-time investment
- We could store the four query types in their own separate cache, but that complicates the logic and results in a lot of repetitive branching of the code
    - For the 1000-line input, the cache miss rate is quite high
    - This is the most memory efficient of all the options, and is potentially the most performant as the input size becomes **much** larger

A key component of this implementation is how we determine intervals for each query:
- We want to gather a list of cache hits for each query
- We want to gather a list of sub-queries that fill in the gaps between cache hits
- We want to update "stale" cache hits that are **partially** outside the initial query range
    - We want to split this cache range into two components: the portion outside the query range, and the portion inside
    - We can accomplish this by re-querying the inside portion and subtracting it from the original cache hit
    - We can then delete the original cache entry and replace it with its two component entries
- For cache ranges that fully encompass the query range, we delete it in favor of the smaller new query range, as we have no way of determining "how much" is before or after the query range without making more expensive API calls

After we have all the intervals figured out, we can actually process everything:
- We sum up all the cache hit values (for all query types)
- We make additional API calls to retrive fills for all gaps in our cache hits
    - We iterate through each fill line/row and calculate totals for each query type 
    - We use separate sets to keep track of "seen" sequence numbers
        - This is not needed for the volume calculation
    - We update the cache with each new API call, using the logic above
- We sum up the new totals from the API calls with the totals from the cache hits
- We only print to `stdout` the field corresponding to the original query type

### Shortcomings
Some observations can be made for future improvements, or optimizations that weren't pursued due to time constraints:
- We are not properly handling cache hits that fully encompass the query range
    - Instead of discarding that cache hit, there may be a way to reuse parts of it without more API calls
- Right now, the cache grows indefinitely
    - Given more time, a TTL or LRU attribute can be added to track cache entries
    - Alternatively a heap/pq could be implemented to keep track of the N most frequent cache hits, dropping the least used items
- We use three different sets to keep track of "seen" sequence numbers when computing count, buys, and sells
    - There may be a way to consolidate the sets so we only need to allocate one

### Benchmarks
On my machine (Apple M1 Pro), I average `~28.6s` for the full input **without caching** and `~22.6s` **with caching**, a `~21%` improvement. However, if we increase the input size greatly using the same range, we see that the time plateaus and the performance gains are more extreme. For example, we see that the time only increases `5.43%` for a 4x increase in input size:

```sh
% time ( cat input.txt | go run . )
( cat input.txt | go run .; )  8.26s user 0.64s system 39% cpu 22.607 total

% time ( cat input4x.txt | go run . )
( cat input4x.txt | go run .; )  9.30s user 0.73s system 42% cpu 23.834 total
```
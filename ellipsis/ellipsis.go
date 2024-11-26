package main

import (
	"bufio"
	"ellipsis/internal/processor"
	"ellipsis/internal/server"
	"fmt"
	"os"
)

func main() {
	processor := processor.NewProcessor(processor.Opts{
		Server: server.NewServer(),
	})
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		processor.ProcessQuery(scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "error reading stdin:", err)
	}
}

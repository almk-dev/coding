package main

import (
	"fmt"
	"os"
	"strings"
	//	"strconv"
	// "unicode"
)

func main() {
	content, _ := os.ReadFile("input.txt")
	lines := strings.Split(string(content), "\n")

	fmt.Println("part1:", part1(lines))
	fmt.Println("part2:", part2(lines))
}

func part1(lines []string) int {
	return 100
}

func part2(lines []string) int {
	return 100
}

package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

func main() {
	content, _ := os.ReadFile("input.txt")
	lines := strings.Split(strings.TrimSuffix(string(content), "\n"), "\n")

	fmt.Println("part1:", part1(lines))
	fmt.Println("part2:", part2(lines))
}

func part1(lines []string) int {
	rmax, gmax, bmax := 12, 13, 14
	total := 0

	for i, line := range lines {
		r, g, b := maxCubes(line)
		if r <= rmax && b <= bmax && g <= gmax {
			total += i + 1
		}
	}

	return total
}

func part2(lines []string) int {
	total := 0

	for _, line := range lines {
		r, g, b := maxCubes(line)
		total += r * g * b
	}

	return total
}

func maxCubes(line string) (r, g, b int) {
	maxes := make(map[string]int)
	draws := strings.Split(strings.Split(line, ": ")[1], "; ")

	for _, draw := range draws {
		counts := strings.Split(draw, ", ")
		for _, count := range counts {
			info := strings.Split(count, " ")
			num, _ := strconv.Atoi(info[0])
			color := info[1]
			maxes[color] = max(maxes[color], num)
		}
	}

	return maxes["red"], maxes["green"], maxes["blue"]
}

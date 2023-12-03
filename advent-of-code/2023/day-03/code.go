package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"unicode"
)

func main() {
	content, _ := os.ReadFile("input.txt")
	lines := strings.Split(strings.TrimSuffix(string(content), "\n"), "\n")

	fmt.Println("part1:", part1(lines))
	fmt.Println("part2:", part2(lines))
}

func part1(lines []string) int {
	nums, syms := parseSchematic(lines)
	total := 0
	for _, sym := range syms {
		sline, spos := sym[0], sym[1]
		for key, _ := range nums {
			kline, kstart, kend := key[0], key[1], key[2]
			if sline-1 <= kline && kline <= sline+1 {
				if (kstart <= spos-1 && spos-1 <= kend) ||
					(kstart <= spos && spos <= kend) ||
					(kstart <= spos+1 && spos+1 <= kend) {
					num, _ := strconv.Atoi(string(lines[kline][kstart:kend]))
					fmt.Println("line:", kline, "number:", num)
					total += num
					delete(nums, key)
				}
			}
		}
	}
	return total
}

func parseSchematic(lines []string) (nums map[[3]int]bool, syms [][2]int) {
	nums = make(map[[3]int]bool)
	syms = make([][2]int, 0)
	for lidx, line := range lines {
		start := -1
		for cidx, char := range line {
			if unicode.IsDigit(char) && start == -1 {
				start = cidx
				if cidx == len(line)-1 {
					nums[[3]int{lidx, start, cidx + 1}] = true
				}
			} else if !unicode.IsDigit(char) {
				if start != -1 {
					nums[[3]int{lidx, start, cidx}] = true
					start = -1
				}
				if char != '.' {
					syms = append(syms, [2]int{lidx, cidx})
				}
			}
		}
	}
	return nums, syms
}

func part2(lines []string) int {
	return 100
}

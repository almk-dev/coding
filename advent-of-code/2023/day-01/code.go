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
	lines := strings.Split(string(content), "\n")

	fmt.Println("part1:", part1(lines))
	fmt.Println("part2:", part2(lines))
}

func part1(lines []string) int {
	var total int = 0
	for _, line := range lines {
		first, last := findDigits(line)
		ni, _ := strconv.Atoi(first + last)
		total += ni
	}
	return total
}

func part2(lines []string) int {
	digits := map[string]int{
		"one":   1,
		"two":   2,
		"three": 3,
		"four":  4,
		"five":  5,
		"six":   6,
		"seven": 7,
		"eight": 8,
		"nine":  9,
	}

	total := 0
	for _, line := range lines {
		first, _ := findDigits(frontReplace(line, &digits))
		_, last := findDigits(backReplace(line, &digits))
		num, _ := strconv.Atoi(first + last)
		total += num
	}
	return total

}

func frontReplace(line string, digits *map[string]int) string {
	for i, _ := range line {
		for key, _ := range *digits {
			if len(key)+i > len(line) {
				continue
			}
			if line[i:i+len(key)] == key {
				return line[:i] + strconv.Itoa((*digits)[key]) + line[i+len(key):]
			}
		}
	}
	return line
}

func backReplace(line string, digits *map[string]int) string {
	rline := []rune(line)
	for i := len(rline); i >= 0; i-- {
		for key, _ := range *digits {
			if i-len(key) < 0 {
				continue
			}
			if string(rline[i-len(key):i]) == key {
				return string(rline[:i-len(key)]) + strconv.Itoa((*digits)[key]) + string(rline[i:])
			}
		}
	}
	return line
}

func findDigits(line string) (int, int) {
	var first, last string
	for _, char := range line {
		if unicode.IsDigit(char) {
			last = string(char)
			if len(first) == 0 {
				first = string(char)
			}
		}
	}
	return first, last
}

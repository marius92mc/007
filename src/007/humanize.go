package main

// This code is from https://github.com/dustin/go-humanize with specific customizations
// for my own use case.

import (
	"errors"
	"math"
	"regexp"
	"strconv"
)

var siPrefixTable = map[float64]string{
	-24: "y", // yocto
	-21: "z", // zepto
	-18: "a", // atto
	-15: "f", // femto
	-12: "p", // pico
	-9:  "n", // nano
	-6:  "µ", // micro
	-3:  "m", // milli
	0:   "",
	3:   "k", // kilo
	6:   "M", // mega
	9:   "G", // giga
	12:  "T", // tera
	15:  "P", // peta
	18:  "E", // exa
	21:  "Z", // zetta
	24:  "Y", // yotta
}

var revSIPrefixTable = revfmap(siPrefixTable)

// revfmap reverses the map and precomputes the power multiplier
func revfmap(in map[float64]string) map[string]float64 {
	rv := map[string]float64{}
	for k, v := range in {
		rv[v] = math.Pow(10, k)
	}
	return rv
}

var riParseRegex *regexp.Regexp

func init() {
	ri := `^([0-9.]+)([`
	for _, v := range siPrefixTable {
		ri += v
	}
	ri += `]?)(.*)`

	riParseRegex = regexp.MustCompile(ri)
}

// ComputeSI finds the most appropriate SI prefix for the given number
// and returns the prefix along with the value adjusted to be within
// that prefix.
//
// e.g. ComputeSI(2.2345e-12) -> (2.2345, "p")
func ComputeSI(input float64) (float64, string) {
	if input == 0 {
		return 0, ""
	}
	exponent := math.Floor(logn(input, 10))
	exponent = math.Floor(exponent/3) * 3

	value := input / math.Pow(10, exponent)

	// Handle special case where value is exactly 1000.0
	// Should return 1M instead of 1000k
	if value == 1000.0 {
		exponent += 3
		value = input / math.Pow(10, exponent)
	}

	prefix := siPrefixTable[exponent]
	return value, prefix
}

// SI returns a string with default formatting.
//
// SI uses Ftoa to format float value, removing trailing zeros.
//
// e.g. SI(1000000, B) -> 1MB
// e.g. SI(2.2345e-12, "F") -> 2.2345pF
func SI(input float64, prec int, sep string, unit string) string {
	value, prefix := ComputeSI(input)
	return Ftoa(value, prec) + sep + prefix + unit
}

var errInvalid = errors.New("invalid input")

// ParseSI parses an SI string back into the number and unit.
//
// e.g. ParseSI(2.2345pF) -> (2.2345e-12, "F", nil)
func ParseSI(input string) (float64, string, error) {
	found := riParseRegex.FindStringSubmatch(input)
	if len(found) != 4 {
		return 0, "", errInvalid
	}
	mag := revSIPrefixTable[found[2]]
	unit := found[3]

	base, err := strconv.ParseFloat(found[1], 64)
	return base * mag, unit, err
}

func stripTrailingZeros(s string) string {
	offset := len(s) - 1
	for offset > 0 {
		if s[offset] == '.' {
			offset--
			break
		}
		if s[offset] != '0' {
			break
		}
		offset--
	}
	return s[:offset+1]
}

// Ftoa converts a float to a string with no trailing zeros.
func Ftoa(num float64, prec int) string {
	return stripTrailingZeros(strconv.FormatFloat(num, 'f', prec, 64))
}

func logn(n, b float64) float64 {
	return math.Log(n) / math.Log(b)
}

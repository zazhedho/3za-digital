package money

import (
	"errors"
	"strconv"
	"strings"
)

var ErrInvalidAmount = errors.New("invalid amount")

func ParseCents(value string) (int64, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, nil
	}

	sign := int64(1)
	if strings.HasPrefix(value, "-") {
		sign = -1
		value = strings.TrimPrefix(value, "-")
	}
	value = strings.TrimPrefix(value, "+")
	if value == "" {
		return 0, ErrInvalidAmount
	}

	parts := strings.Split(value, ".")
	if len(parts) > 2 {
		return 0, ErrInvalidAmount
	}

	wholePart := parts[0]
	if wholePart == "" {
		wholePart = "0"
	}
	if !digitsOnly(wholePart) {
		return 0, ErrInvalidAmount
	}

	whole, err := strconv.ParseInt(wholePart, 10, 64)
	if err != nil {
		return 0, err
	}

	fraction := ""
	if len(parts) == 2 {
		fraction = parts[1]
	}
	if fraction != "" && !digitsOnly(fraction) {
		return 0, ErrInvalidAmount
	}

	centsPart := int64(0)
	if len(fraction) > 0 {
		firstTwo := fraction
		if len(firstTwo) > 2 {
			firstTwo = firstTwo[:2]
		}
		for len(firstTwo) < 2 {
			firstTwo += "0"
		}
		centsPart, err = strconv.ParseInt(firstTwo, 10, 64)
		if err != nil {
			return 0, err
		}
		if len(fraction) > 2 && fraction[2] >= '5' {
			centsPart++
			if centsPart == 100 {
				whole++
				centsPart = 0
			}
		}
	}

	return sign * (whole*100 + centsPart), nil
}

func FormatCents(cents int64) string {
	sign := ""
	if cents < 0 {
		sign = "-"
		cents = -cents
	}
	return sign + strconv.FormatInt(cents/100, 10) + "." + leftPad2(strconv.FormatInt(cents%100, 10))
}

func Normalize(value string) (string, error) {
	cents, err := ParseCents(value)
	if err != nil {
		return "", err
	}
	return FormatCents(cents), nil
}

func NormalizeOrZero(value string) string {
	normalized, err := Normalize(value)
	if err != nil {
		return "0.00"
	}
	return normalized
}

func NormalizeOrTrim(value string) string {
	normalized, err := Normalize(value)
	if err != nil {
		return strings.TrimSpace(value)
	}
	return normalized
}

func NormalizePositive(value string) (string, error) {
	cents, err := ParseCents(value)
	if err != nil || cents <= 0 {
		return "", ErrInvalidAmount
	}
	return FormatCents(cents), nil
}

func Add(left string, right string) (string, error) {
	leftCents, err := ParseCents(left)
	if err != nil {
		return "", err
	}
	rightCents, err := ParseCents(right)
	if err != nil {
		return "", err
	}
	return FormatCents(leftCents + rightCents), nil
}

func Sub(left string, right string) (string, error) {
	leftCents, err := ParseCents(left)
	if err != nil {
		return "", err
	}
	rightCents, err := ParseCents(right)
	if err != nil {
		return "", err
	}
	return FormatCents(leftCents - rightCents), nil
}

func SubOrZero(left string, right string) string {
	result, err := Sub(left, right)
	if err != nil {
		return "0.00"
	}
	return result
}

func MarkupAmount(providerCharge string, percent string) (string, string, string, error) {
	providerCents, err := ParseCents(providerCharge)
	if err != nil {
		return "", "", "", err
	}
	basisPoints, err := ParsePercentBasisPoints(percent)
	if err != nil {
		return "", "", "", err
	}
	markupCents := divRound(providerCents*basisPoints, 10000)
	amountCents := providerCents + markupCents
	return FormatCents(providerCents), FormatCents(amountCents), FormatCents(markupCents), nil
}

func ParsePercentBasisPoints(value string) (int64, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, nil
	}
	parts := strings.Split(value, ".")
	if len(parts) > 2 {
		return 0, ErrInvalidAmount
	}
	wholePart := parts[0]
	if wholePart == "" {
		wholePart = "0"
	}
	if !digitsOnly(wholePart) {
		return 0, ErrInvalidAmount
	}
	whole, err := strconv.ParseInt(wholePart, 10, 64)
	if err != nil {
		return 0, err
	}
	fraction := ""
	if len(parts) == 2 {
		fraction = parts[1]
	}
	if fraction != "" && !digitsOnly(fraction) {
		return 0, ErrInvalidAmount
	}
	firstTwo := fraction
	if len(firstTwo) > 2 {
		firstTwo = firstTwo[:2]
	}
	for len(firstTwo) < 2 {
		firstTwo += "0"
	}
	frac := int64(0)
	if firstTwo != "" {
		frac, err = strconv.ParseInt(firstTwo, 10, 64)
		if err != nil {
			return 0, err
		}
	}
	if len(fraction) > 2 && fraction[2] >= '5' {
		frac++
		if frac == 100 {
			whole++
			frac = 0
		}
	}
	return whole*100 + frac, nil
}

func digitsOnly(value string) bool {
	for _, ch := range value {
		if ch < '0' || ch > '9' {
			return false
		}
	}
	return true
}

func leftPad2(value string) string {
	if len(value) == 1 {
		return "0" + value
	}
	return value
}

func divRound(numerator int64, denominator int64) int64 {
	if denominator == 0 {
		return 0
	}
	if numerator >= 0 {
		return (numerator + denominator/2) / denominator
	}
	return (numerator - denominator/2) / denominator
}

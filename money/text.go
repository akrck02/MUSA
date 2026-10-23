package money

import (
	"regexp"
	"strconv"
	"strings"
)

const (
	centsInEuro    = 100
	thousandsGroup = 3
	euroSign       = "€"
)

// amountPattern is up to nine digits of euros and, after a comma or a dot, one or two of cents.
var amountPattern = regexp.MustCompile(`^(\d{1,9})(?:[.,](\d{1,2}))?$`)

// Format writes the amount with its currency, the Spanish way: "1.234,50 €"; whole amounts drop
// the decimals ("12 €").
func Format(amount Money) string { return FormatForInput(amount) + " " + euroSign }

// FormatForInput writes the amount without currency ("9,90"), to fill a field that edits it.
func FormatForInput(amount Money) string {
	sign := ""
	if amount < 0 {
		sign = "-"
		amount = -amount
	}
	euros := groupThousands(int64(amount / centsInEuro))
	cents := int64(amount % centsInEuro)
	if cents == 0 {
		return sign + euros
	}
	return sign + euros + "," + leadingZero(cents)
}

// Parse reads what the artist types in a price field ("12", "12,5", "12.50 €"); found is false for
// empty text, letters, negative amounts and more than two decimals.
func Parse(text string) (amount Money, found bool) {
	clean := strings.TrimSpace(strings.ReplaceAll(text, euroSign, ""))
	match := amountPattern.FindStringSubmatch(clean)
	if match == nil {
		return 0, false
	}
	decimals := match[2]
	for len(decimals) < 2 {
		decimals += "0" // "12,5" means 50 cents
	}
	euros, _ := strconv.ParseInt(match[1], 10, 64) // the pattern only lets digits through
	cents, _ := strconv.ParseInt(decimals, 10, 64) // the same
	return Money(euros*centsInEuro + cents), true
}

func groupThousands(value int64) string {
	digits := strconv.FormatInt(value, 10)
	var grouped strings.Builder
	for index, digit := range digits {
		if index > 0 && (len(digits)-index)%thousandsGroup == 0 {
			grouped.WriteByte('.')
		}
		grouped.WriteRune(digit)
	}
	return grouped.String()
}

func leadingZero(cents int64) string {
	if cents < 10 {
		return "0" + strconv.FormatInt(cents, 10)
	}
	return strconv.FormatInt(cents, 10)
}

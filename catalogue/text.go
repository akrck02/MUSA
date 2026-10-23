package catalogue

import (
	"regexp"
	"strconv"
	"strings"
	"unicode/utf16"
)

// accents maps the accented letters of Spanish and its neighbours to their plain ones.
var accents = strings.NewReplacer(
	"á", "a", "à", "a", "ä", "a", "â", "a", "ã", "a",
	"é", "e", "è", "e", "ë", "e", "ê", "e",
	"í", "i", "ì", "i", "ï", "i", "î", "i",
	"ó", "o", "ò", "o", "ö", "o", "ô", "o", "õ", "o",
	"ú", "u", "ù", "u", "ü", "u", "û", "u",
	"ñ", "n", "ç", "c",
)

// Fold lowercases text and takes its accents off, so "Lámina" and "lamina" compare equal.
func Fold(text string) string { return accents.Replace(strings.ToLower(text)) }

var spaces = regexp.MustCompile(`\s+`)

// TagDisplayName trims a tag's name and joins its inner spaces into one, keeping the casing typed.
func TagDisplayName(raw string) string { return spaces.ReplaceAllString(strings.TrimSpace(raw), " ") }

// TagKey is what makes two tags the same: "pokemon ", "Pokémon" and "POKEMON" share one key.
func TagKey(raw string) string { return Fold(TagDisplayName(raw)) }

// DistinctTagNames drops blank names and repeated tags, keeping the first spelling of each.
func DistinctTagNames(names []string) []string {
	seen := map[string]bool{}
	var distinct []string
	for _, name := range names {
		display := TagDisplayName(name)
		key := Fold(display)
		if display == "" || seen[key] {
			continue
		}
		seen[key] = true
		distinct = append(distinct, display)
	}
	return distinct
}

// FNV-1a's numbers, over 32 bits. A plain "times 31" hash barely changes modulo eight, so most tags
// would share a colour.
const (
	fnvOffset int32 = -2128831035
	fnvPrime  int32 = 16777619
)

// TagColorFor picks a colour from a tag's key, always the same one for the same key.
func TagColorFor(key string) TagColor {
	hash := fnvOffset
	for _, unit := range utf16.Encode([]rune(key)) {
		hash = (hash ^ int32(unit)) * fnvPrime
	}
	mixed := int(hash ^ int32(uint32(hash)>>16))
	count := len(TagColors)
	return TagColors[((mixed%count)+count)%count]
}

const millimetresInCentimetre = 10

var (
	countPattern       = regexp.MustCompile(`^\d{1,6}$`)
	centimetresPattern = regexp.MustCompile(`^(\d{1,5})(?:[.,](\d))?$`)
)

// ParseCount reads a whole number of units, such as stock; found is false for anything else.
func ParseCount(text string) (count int, found bool) {
	clean := strings.TrimSpace(text)
	if !countPattern.MatchString(clean) {
		return 0, false
	}
	count, _ = strconv.Atoi(clean) // the pattern only lets digits through
	return count, true
}

// ParseCentimetres reads centimetres with at most one decimal ("29,7") as millimetres (297).
func ParseCentimetres(text string) (millimetres int, found bool) {
	match := centimetresPattern.FindStringSubmatch(strings.TrimSpace(text))
	if match == nil {
		return 0, false
	}
	whole, _ := strconv.Atoi(match[1])   // the pattern only lets digits through
	decimal, _ := strconv.Atoi(match[2]) // empty when there is no decimal, which reads as zero
	return whole*millimetresInCentimetre + decimal, true
}

// FormatCentimetres writes millimetres back as centimetres ("29,7", "30").
func FormatCentimetres(millimetres int) string {
	whole := strconv.Itoa(millimetres / millimetresInCentimetre)
	if millimetres%millimetresInCentimetre == 0 {
		return whole
	}
	return whole + "," + strconv.Itoa(millimetres%millimetresInCentimetre)
}

// jsonList writes strings as a JSON list, to hand SQLite a list of ids in one argument.
func jsonList(values []string) string {
	quoted := make([]string, len(values))
	for index, value := range values {
		quoted[index] = strconv.Quote(value)
	}
	return "[" + strings.Join(quoted, ",") + "]"
}

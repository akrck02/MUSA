package catalogue

import "strings"

// stressedEndings are the accented last syllables that lose their accent in the plural:
// «camión», «camiones».
var stressedEndings = map[string]string{"án": "an", "én": "en", "ín": "in", "ón": "on", "ún": "un", "és": "es"}

// Plural is the Spanish plural of a tag's name, for the tab that groups it: «Lámina» gives
// «Láminas», «Llavero» «Llaveros», «Imán» «Imanes» and «Lápiz» «Lápices». A name of several words
// takes it on its first one: «Taza grande» gives «Tazas grande»… so keep tags to one word.
func Plural(name string) string {
	first, rest, hasRest := strings.Cut(name, " ")
	plural := pluralWord(first)
	if hasRest {
		return plural + " " + rest
	}
	return plural
}

func pluralWord(word string) string {
	lower := strings.ToLower(word)
	switch {
	case word == "":
		return word
	case strings.HasSuffix(lower, "s"), strings.HasSuffix(lower, "x"):
		return word
	case strings.ContainsAny(lower[len(lower)-1:], "aeiou"), strings.HasSuffix(lower, "é"), strings.HasSuffix(lower, "á"), strings.HasSuffix(lower, "ó"):
		return word + "s"
	case strings.HasSuffix(lower, "z"):
		return strings.TrimSuffix(word, word[len(word)-1:]) + "ces"
	}
	for stressed, plain := range stressedEndings {
		if strings.HasSuffix(word, stressed) {
			return strings.TrimSuffix(word, stressed) + plain + "es"
		}
	}
	return word + "es"
}

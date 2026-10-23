package catalogue

// TagColor is one of the fixed colours tags are painted with; the screens turn each into real
// colours for the light and the dark theme.
type TagColor string

// The tag colours. The words are the ones stored in the database.
const (
	Rose   TagColor = "ROSE"
	Amber  TagColor = "AMBER"
	Lime   TagColor = "LIME"
	Teal   TagColor = "TEAL"
	Sky    TagColor = "SKY"
	Violet TagColor = "VIOLET"
	Slate  TagColor = "SLATE"
	Coral  TagColor = "CORAL"
)

// TagColors lists every tag colour, in the order TagColorFor picks from.
var TagColors = []TagColor{Rose, Amber, Lime, Teal, Sky, Violet, Slate, Coral}

// Tag is a free label (a theme, a fandom, a fair) that groups and filters products.
type Tag struct {
	ID    string
	Name  string
	Color TagColor
}

// MaximumSuggestions is how many tags the editor suggests at once.
const MaximumSuggestions = 8

// SuggestTags picks the tags not chosen yet whose name holds what is typed; with nothing typed,
// any tags not chosen yet.
func SuggestTags(all []Tag, chosenNames []string, typed string) []Tag {
	chosen := map[string]bool{}
	for _, name := range chosenNames {
		chosen[TagKey(name)] = true
	}
	typedKey := TagKey(typed)
	var suggestions []Tag
	for _, tag := range all {
		key := TagKey(tag.Name)
		if chosen[key] || !containsText(key, typedKey) {
			continue
		}
		suggestions = append(suggestions, tag)
		if len(suggestions) == MaximumSuggestions {
			break
		}
	}
	return suggestions
}

package customers

var problemWords = map[Code]string{
	CodeBlankName:      "Ponle un nombre",
	CodeNetworkPattern: "El enlace tiene que llevar " + UserMark,
}

// ProblemWords is what the artist reads about a problem, in Spanish; the code when it has no words.
func ProblemWords(code Code) string {
	if words, found := problemWords[code]; found {
		return words
	}
	return string(code)
}

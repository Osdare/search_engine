package types

//term -> posting
type InvertedIndex map[string]Posting

type Posting struct {
	NormUrl       string
	TermFrequency int
}

package types

// term -> postings
type InvertedIndex map[string][]Posting

type Posting struct {
	NormUrl       string
	TermFrequency int
}

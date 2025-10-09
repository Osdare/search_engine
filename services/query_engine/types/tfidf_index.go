package types

type Index struct {
	Word string
	Postings []ScorePosting
}

type ScorePosting struct {
	NormUrl string
	Score float64
}
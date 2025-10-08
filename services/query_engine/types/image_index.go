package types

type ImageIndex struct {
	Word string
	Postings []ImagePosting
}

type ImagePosting struct {
	TermFrequency int
	ImageUrl      string
}
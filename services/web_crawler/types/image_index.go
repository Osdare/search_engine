package types

type ImageIndex map[string][]ImagePosting

type ImagePosting struct {
	TermFrequency int
	ImageUrl      string
}

package types

type Image struct {
	PageUrl  string
	ImageUrl string
	//all text associated with the image
	//this includes alt text and caption so far
	Text     string
}

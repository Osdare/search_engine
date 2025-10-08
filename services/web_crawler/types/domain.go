package types

//times are in nanoseconds
type Domain struct {
	Name        string
	CrawlDelay  int64
	LastCrawled int64
	Allowed     []string
	Disallowed  []string
}

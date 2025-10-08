package tfidf

import (
	"math"
)

func relativeFrequency(termFrequency int, docLength int) float64 {
	return float64(termFrequency) / float64(docLength)
}

func inverseDocumentFrequency(totalDocs int, docsWithTerm int) float64 {
	N := float64(totalDocs)
	dft := float64(docsWithTerm)
	idf := math.Log((1+N)/(1+dft)) + 1
	return idf
}

func Tfidf(termFrequency int, docLength int, totalDocs int, docsWithTerm int) float64 {
	tf := relativeFrequency(termFrequency, docLength)
	idf := inverseDocumentFrequency(totalDocs, docsWithTerm)
	return tf * idf
}

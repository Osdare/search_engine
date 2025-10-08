package pagerank

import (
	"indexer/types"
)

func GetPageRanks(pageNodes []types.PageNode) (pageRanks map[string]float64) {

	backLinks := make(map[string][]string)
	outLinks := make(map[string][]string)
	outLinkCounts := make(map[string]int)
	for _, node := range pageNodes {
		backLinks[node.Hash] = node.BackLinks
		outLinks[node.Hash] = node.OutLinks
		outLinkCounts[node.Hash] = len(node.OutLinks)
	}
	var initialRank float64 = 1 / float64(len(pageNodes))
	pageRanks = make(map[string]float64)
	for urlHash := range outLinks {
		pageRanks[urlHash] = initialRank
	}

	damping := 0.85
	iterations := 10
	for range make([]struct{}, iterations) {
		newRanks := make(map[string]float64)

		var danglingMass float64
		for hash, rank := range pageRanks {
			if outLinkCounts[hash] == 0 {
				danglingMass += rank
			}
		}
		massContribution := damping * (danglingMass / float64(len(pageNodes)))

		for hash := range pageRanks {
			var cumRank float64

			urlBacklinks, exists := backLinks[hash]
			if exists {
				for _, backlink := range urlBacklinks {
					outlinkCount, ok := outLinkCounts[backlink]
					if ok {
						backlinkRank, ok := pageRanks[backlink]
						if ok {
							cumRank += backlinkRank / float64(outlinkCount)
						}
					}
				}
			}
			newRanks[hash] = (1-damping)/float64(len(pageNodes)) + massContribution + damping*cumRank
		}

		pageRanks = newRanks
	}
	
	return pageRanks
}

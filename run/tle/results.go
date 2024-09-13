package tle

type Results struct {
	LatenciesNs []int64 `json:"latenciesNs"`
	Polls       []int   `json:"polls"`
	Wfts        []int   `json:"wfts"`
	QueryTimes  []int64 `json:"queryTimes"`
}

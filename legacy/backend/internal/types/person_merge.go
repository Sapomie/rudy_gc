package types

type PersonMergeCandidate struct {
	Person    *Person  `json:"person"`
	CastNames []string `json:"castNames"`
}

type PersonMergePreview struct {
	Keep               *PersonMergeCandidate   `json:"keep"`
	Sources            []*PersonMergeCandidate `json:"sources"`
	MoveCastNames      []string                `json:"moveCastNames"`
	RemovePersonIds    []int64                 `json:"removePersonIds"`
	AffectedMovieCount int                     `json:"affectedMovieCount"`
}

type PersonMergeResult struct {
	KeepPersonId       int64    `json:"keepPersonId"`
	RemovedPersonIds   []int64  `json:"removedPersonIds"`
	MoveCastNames      []string `json:"moveCastNames"`
	AffectedMovieCount int      `json:"affectedMovieCount"`
}

package types

type Directory struct {
	Id        int64  `json:"id"`
	ParentId  int64  `json:"parent_id"`
	Name      string `json:"name"`
	Depth     int64  `json:"depth"`
	Path      string `json:"path"`
	CreatedOn int64  `json:"created_on"`
	UpdatedOn int64  `json:"updated_on"`
}

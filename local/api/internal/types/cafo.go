package types

type CafoProfile struct {
	Id        int64  `json:"id"`
	PersonId  int64  `json:"personId"`
	Name      string `json:"name"`
	Alias     string `json:"alias"`
	Chinese   string `json:"chinese"`
	BirthDay  int64  `json:"birthDay"`
	Height    int64  `json:"height"`
	Cup       string `json:"cup"`
	Bwh       string `json:"bwh"`
	Avatar    string `json:"avatar"`
	Biography string `json:"biography"`
	CreatedOn int64  `json:"createdOn"`
	UpdatedOn int64  `json:"updatedOn"`
}

type CafoProfileWriteForm struct {
	Name      string `json:"name"`
	Alias     string `json:"alias"`
	Chinese   string `json:"chinese"`
	BirthDay  int64  `json:"birthDay"`
	Height    int64  `json:"height"`
	Cup       string `json:"cup"`
	Bwh       string `json:"bwh"`
	Avatar    string `json:"avatar"`
	Biography string `json:"biography"`
}

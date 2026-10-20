package types

// Detail 表示抓到的详情页原文
type Detail struct {
	Id        int64
	Name      string
	JavId     string
	Prefix    string
	QueryUrl  string
	Content   string
	CreatedOn int64
	UpdatedOn int64
}

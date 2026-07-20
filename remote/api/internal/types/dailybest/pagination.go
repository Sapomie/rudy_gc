package dailybest

const (
	DefaultPageSize int64 = 24
	MaxPageSize     int64 = 200
)

type PageQuery struct {
	Page     int64
	PageSize int64
	Enabled  bool
}

func (q PageQuery) Offset() int {
	return int((q.Page - 1) * q.PageSize)
}

func (q PageQuery) Limit() int {
	return int(q.PageSize)
}

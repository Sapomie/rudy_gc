package types

type DirSummary struct {
	Id        int64  `json:"id"`
	ParentId  int64  `json:"parent_id"`
	Name      string `json:"name"`
	Depth     int64  `json:"depth"`
	Path      string `json:"path"`
	UpdatedOn int64  `json:"updated_on"`

	// 可选聚合
	FilmCount     *int64  `json:"film_count,omitempty"`
	TotalSize     *int64  `json:"total_size,omitempty"`
	LastFilmBirth *int64  `json:"last_film_birth,omitempty"`
	LastUpdatedOn *int64  `json:"last_updated_on,omitempty"`
	CoverURL      *string `json:"cover_url,omitempty"` // 目录代表封面（可选）
}

type DirDetail struct {
	Directory   *Directory   `json:"directory"`
	Breadcrumbs []Breadcrumb `json:"breadcrumbs"`
	Stats       *DirStats    `json:"stats,omitempty"`
	MovieTypes  []*MovieType `json:"movie_types"`
}

type Breadcrumb struct {
	Id   int64  `json:"id"`
	Name string `json:"name"`
	Path string `json:"path"`
}

type DirStats struct {
	Recursive     bool         `json:"recursive"`
	FilmCount     int64        `json:"film_count"`
	TotalSize     int64        `json:"total_size"`
	LastFilmBirth int64        `json:"last_film_birth"`
	LastUpdatedOn int64        `json:"last_updated_on"`
	Buckets       []TimeBucket `json:"buckets,omitempty"` // 按月/年聚合时返回
}

type TimeBucket struct {
	Key   string `json:"key"` // "YYYY-MM" 或 "YYYY"
	Count int64  `json:"count"`
	Size  int64  `json:"size"`
}

type PageResult[T any] struct {
	Items      []T   `json:"items"`
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

type ListDirFilmsRequest struct {
	DirID     int64  `json:"dir_id"`       // 当前目录ID
	Page      int64  `json:"page"`         // 页码
	PageSize  int64  `json:"page_size"`    // 每页数量
	OrderBy   string `json:"od" form:"od"` // 排序字段名，如 "updated_on"、"name" 等 	// 是否升序
	Recursive bool   `json:"recursive"`    // 是否包含子目录
}

// DirPageRequest 统一封装目录详情页所需的入参
type DirPageRequest struct {
	UseRoot   bool  `form:"-"` // 从 path 判断，不来自 query
	DirID     int64 `form:"-"` // 从 path 判断，不来自 query
	Recursive bool  `form:"recursive"`

	Page      int64  `form:"p"`
	PageSize  int64  `form:"ps"`
	SortField string `form:"od"` // 对齐现有参数命名

	ChildrenPage int64 `form:"cp"`
	ChildrenSize int64 `form:"cs"`
}

// DirPageResult 统一返回页面渲染所需数据
type DirPageResult struct {
	Detail   *DirDetail    `json:"detail"`   // {Directory, Breadcrumbs, Stats, MovieTypes(Film)}
	Children []*DirSummary `json:"children"` // 子目录卡片
	Total    int64         `json:"total"`    // 影片总数（用于分页）
}

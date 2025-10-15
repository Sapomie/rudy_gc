package consts

// ---- 字幕状态 ----

// ---- 自制状态 ----
const (
	FilmNoSelfMake = iota + 1
	FilmSelfMake
)

// ---- 元数据扫描需求 ----
const (
	FilmMetaDataNeedScan = iota + 1
	FilmMetaDataNoNeedScan
)

// ---- 修复/去码状态 ----
const (
	FilmNotErased = iota + 1
	FilmErased
	FilmNoMosaic
)

// ---- 删除状态 ----
const (
	FilmIsNotRemoved = iota + 1
	FilmIsRemoved
)

const (
	FilmNoSub = iota + 1
	FilmHasSub
)

const (
	MovieAll              int64 = 1 // amovie里的所有数据，不再在仅限与vfilm里
	OwnedAll              int64 = 2 // vfilm里的所有,不加条件
	OwnedAllNotRemoved    int64 = 3 // is_removed=1
	OwnedHasSubNotRemoved int64 = 4 // has_sub=2, is_removed=1
	OwnedNoSubNotRemoved  int64 = 5 // has_sub=1, is_removed=1
	OwnedRemoved          int64 = 6 // is_removed=2
)

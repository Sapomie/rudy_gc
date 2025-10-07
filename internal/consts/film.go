package consts

// ---- 字幕状态 ----
const (
	FilmNoSub = iota + 1
	FilmHasSub
)

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

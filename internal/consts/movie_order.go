// internal/consts/movie_order.go
package consts

// OrderBy 常量值：前端请求中 orderby 参数可选项
const (
	OrderByDetailUpdateTime = "du"  // 按detail更新日期
	OrderByCastAgeDesc      = "cad" // 按演员年龄倒序
	OrderByCastAgeAsc       = "caa" // 按演员年龄正序
	OrderByViewerWatched    = "vw"  // 按观看人数

	OrderByReleasingDate = "rd" // 按上映日期

	OrderByRankDate    = "rk"  // 按第一次上榜日期倒序
	OrderByHighestRank = "hrk" // 按最高排名
	OrderByDaysInRank  = "drk" // 在榜天数

	OrderByBirthTime      = "bt"   // 按 vfilm 下载时间
	OrderByMediaBirthTime = "mbt"  // 按 w_media 下载时间
	OrderByScTimes        = "sc"   // 按评分次数
	OrderByComeTimes      = "co"   // 按出现次数
	OrderByLastScTime     = "lsct" // 按最后评分时间
)

// 允许的排序字段枚举（防止前端乱传）
const (
	SortByUpdatedOn = "updated_on"
	SortByName      = "name"
	SortByCreatedOn = "created_on"
)

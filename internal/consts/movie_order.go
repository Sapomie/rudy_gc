// internal/consts/movie_order.go
package consts

// OrderBy 常量值：前端请求中 orderby 参数可选项
const (
	OrderByBirthTime     = "bt"   // 按影片拍摄时间
	OrderByReleasingDate = "rd"   // 按上映日期
	OrderByRankDate      = "rk"   // 按排名起始日
	OrderByCastAgeDesc   = "cad"  // 按演员年龄倒序
	OrderByCastAgeAsc    = "caa"  // 按演员年龄正序
	OrderByViewerWatched = "vw"   // 按观看人数
	OrderByName          = "n"    // 按名称
	OrderByScTimes       = "sc"   // 按评分次数
	OrderByComeTimes     = "co"   // 按出现次数
	OrderByLastScTime    = "lsct" // 按最后评分时间
	OrderByHighestRank   = "hrk"  // 按最高排名
)

// internal/types/item_patch.go
package types

type ItemPatch struct {
	// 可选更新的字段（传 nil 表示“不改”）
	HasDownloadCover *int64
	HasChinese       *int64
	HasDetail        *int64
	DetailNeedScan   *int64
	DetailBirthTime  *int64
	DetailUpdateTime *int64
	UpdatedOn        *int64
}

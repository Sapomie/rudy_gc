package consts

type UpsertStatus int

const (
	UpsertInserted  UpsertStatus = iota + 1 // 新插入
	UpsertUpdated                           // 已存在并更新
	UpsertUnchanged                         // 已存在但无变化
)

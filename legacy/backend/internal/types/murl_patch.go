package types

// MurlPatch 表示对 bm_murl 的部分字段更新（nil 表示不改）
type MurlPatch struct {
	JacketImgLocal *string
	JacketImg      *string
	UpdatedOn      *int64
}

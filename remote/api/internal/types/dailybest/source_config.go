package dailybest

import "fmt"

// SourceConfig 的探测地址由当前域名生成，数据库只保存域名。
type SourceConfig struct {
	JavAddress     string `json:"jav_address"`
	TargetProbeURL string `json:"target_probe_url"`
	CreatedOn      int64  `json:"created_on"`
	UpdatedOn      int64  `json:"updated_on"`
}

type UpdateSourceConfigRequest struct {
	JavAddress string `json:"jav_address"`
}

func SourcePageURL(address string, page int64) string {
	return fmt.Sprintf("https://%s/cn/vl_bestrated.php?&mode=&page=%d", address, page)
}

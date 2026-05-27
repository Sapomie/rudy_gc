package sc

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

// 定义结构体
type scJsonData struct {
	DurationMinutes int64  `json:"duration"`
	Fg              string `json:"fg"`
	Vessel          string `json:"vessel"`
	Remarks         string `json:"remarks"`
}

func isDataJsonFile(name string) bool {
	return filepath.Base(name) == "data.json"
}

// 从文件读取 JSON 数据
func getJsonData(file string) (scJsonData, error) {
	var d scJsonData

	// 检查文件名是否符合要求
	if !isDataJsonFile(file) {
		return d, errors.New("文件名必须为 data.json")
	}

	// 读取文件内容
	bytes, err := os.ReadFile(file)
	if err != nil {
		return d, err
	}

	// 解析 JSON
	if err := json.Unmarshal(bytes, &d); err != nil {
		return d, err
	}

	return d, nil
}

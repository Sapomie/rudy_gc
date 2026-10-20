package fc2

import (
	"encoding/json"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
)

const ParserVersion = "v1"

var (
	codePattern       = regexp.MustCompile(`(?i)^FC2PPV-\s*(\d+)$`)
	keyPattern        = regexp.MustCompile(`(?i)^FC2PPV-\s*(\d+)-([1-9]\d*)$`)
	encodedKeyPattern = regexp.MustCompile(`(?i)^15122252531-(\d+)-([1-9]\d*)(?:_|$)`)
	fieldPattern      = regexp.MustCompile(`^【([^】]+)】\s*[：:]\s*(.*)$`)
	attachmentPattern = regexp.MustCompile(`(?i)^(.+?\.(?:jpe?g|png|gif|webp))\s*\(([0-9]+(?:\.[0-9]+)?)\s*(B|KB|MB|GB),\s*下载次数:\s*(\d+)\)$`)
	sizePattern       = regexp.MustCompile(`(?i)^([0-9]+(?:\.[0-9]+)?)\s*(B|K|KB|M|MB|G|GB)$`)
)

type Attachment struct {
	FileName       string `json:"file_name"`
	MediaType      string `json:"media_type"`
	SizeText       string `json:"size_text"`
	SizeBytes      int64  `json:"size_bytes"`
	DownloadCount  int64  `json:"download_count"`
	UploadTimeText string `json:"upload_time_text,omitempty"`
}

type Meta struct {
	ParserVersion string       `json:"parser_version"`
	Title         string       `json:"title"`
	PerformerText string       `json:"performer_text"`
	SizeText      string       `json:"size_text"`
	SizeBytes     int64        `json:"size_bytes"`
	MaskType      string       `json:"mask_type"`
	SeedTermText  string       `json:"seed_term_text"`
	Attachments   []Attachment `json:"attachments"`
}

func NormalizeCode(value string) (string, int64, bool) {
	match := codePattern.FindStringSubmatch(strings.TrimSpace(value))
	if len(match) != 2 {
		return "", 0, false
	}
	number, err := strconv.ParseInt(match[1], 10, 64)
	if err != nil || number <= 0 {
		return "", 0, false
	}
	return fmt.Sprintf("FC2PPV-%d", number), number, true
}

func NormalizeKey(value string) (string, string, int64, int64, bool) {
	match := keyPattern.FindStringSubmatch(strings.TrimSpace(value))
	if len(match) != 3 {
		return "", "", 0, 0, false
	}
	ppvNo, err := strconv.ParseInt(match[1], 10, 64)
	if err != nil || ppvNo <= 0 {
		return "", "", 0, 0, false
	}
	codeSeq, err := strconv.ParseInt(match[2], 10, 64)
	if err != nil || codeSeq <= 0 {
		return "", "", 0, 0, false
	}
	code := fmt.Sprintf("FC2PPV-%d", ppvNo)
	return BuildKey(code, codeSeq), code, ppvNo, codeSeq, true
}

func BuildKey(code string, codeSeq int64) string {
	normalized, _, ok := NormalizeCode(code)
	if !ok || codeSeq <= 0 {
		return ""
	}
	return fmt.Sprintf("%s-%d", normalized, codeSeq)
}

func NormalizeEncodedMediaKey(value string) (string, bool) {
	base := strings.ToUpper(strings.TrimSpace(value))
	match := encodedKeyPattern.FindStringSubmatch(base)
	if len(match) != 3 {
		return "", false
	}
	ppvNo, err := strconv.ParseInt(match[1], 10, 64)
	if err != nil || ppvNo <= 0 {
		return "", false
	}
	codeSeq, err := strconv.ParseInt(match[2], 10, 64)
	if err != nil || codeSeq <= 0 {
		return "", false
	}
	return BuildKey(fmt.Sprintf("FC2PPV-%d", ppvNo), codeSeq), true
}

func Parse(raw string) (Meta, error) {
	meta := Meta{ParserVersion: ParserVersion, Attachments: []Attachment{}}
	lines := strings.Split(strings.ReplaceAll(raw, "\r\n", "\n"), "\n")
	fields := make(map[string]string)
	for _, line := range lines {
		match := fieldPattern.FindStringSubmatch(strings.TrimSpace(line))
		if len(match) == 3 {
			fields[strings.TrimSpace(match[1])] = strings.TrimSpace(match[2])
		}
	}

	meta.Title = fields["影片名称"]
	meta.PerformerText = fields["出演女优"]
	meta.SizeText = fields["影片容量"]
	meta.MaskType = fields["是否有码"]
	meta.SeedTermText = fields["种子期限"]
	if meta.Title == "" {
		return Meta{}, fmt.Errorf("缺少影片名称")
	}
	if meta.SizeText != "" {
		sizeBytes, err := ParseSizeBytes(meta.SizeText)
		if err != nil {
			return Meta{}, fmt.Errorf("影片容量解析失败: %w", err)
		}
		meta.SizeBytes = sizeBytes
	}

	for index, line := range lines {
		match := attachmentPattern.FindStringSubmatch(strings.TrimSpace(line))
		if len(match) != 5 {
			continue
		}
		sizeText := strings.TrimSpace(match[2] + " " + strings.ToUpper(match[3]))
		sizeBytes, err := ParseSizeBytes(sizeText)
		if err != nil {
			return Meta{}, fmt.Errorf("附件容量解析失败: %w", err)
		}
		downloadCount, err := strconv.ParseInt(match[4], 10, 64)
		if err != nil {
			return Meta{}, fmt.Errorf("附件下载次数解析失败: %w", err)
		}
		fileName := strings.TrimSpace(match[1])
		mediaType := "image"
		if strings.EqualFold(strings.TrimPrefix(strings.ToLower(fileName[strings.LastIndex(fileName, "."):]), "."), "gif") {
			mediaType = "animation"
		}
		uploadTimeText := ""
		for next := index + 1; next < len(lines) && next <= index+3; next++ {
			candidate := strings.TrimSpace(lines[next])
			if strings.HasSuffix(candidate, " 上传") {
				uploadTimeText = strings.TrimSpace(strings.TrimSuffix(candidate, " 上传"))
				break
			}
		}
		meta.Attachments = append(meta.Attachments, Attachment{
			FileName:       fileName,
			MediaType:      mediaType,
			SizeText:       sizeText,
			SizeBytes:      sizeBytes,
			DownloadCount:  downloadCount,
			UploadTimeText: uploadTimeText,
		})
	}

	return meta, nil
}

func ParseSizeBytes(value string) (int64, error) {
	match := sizePattern.FindStringSubmatch(strings.TrimSpace(value))
	if len(match) != 3 {
		return 0, fmt.Errorf("无法识别容量 %q", value)
	}
	number, err := strconv.ParseFloat(match[1], 64)
	if err != nil || number < 0 {
		return 0, fmt.Errorf("容量数字无效 %q", value)
	}
	unit := strings.ToUpper(match[2])
	multiplier := float64(1)
	switch unit {
	case "K", "KB":
		multiplier = 1024
	case "M", "MB":
		multiplier = 1024 * 1024
	case "G", "GB":
		multiplier = 1024 * 1024 * 1024
	}
	bytes := number * multiplier
	if bytes > math.MaxInt64 {
		return 0, fmt.Errorf("容量超出范围 %q", value)
	}
	return int64(math.Round(bytes)), nil
}

func Marshal(meta Meta) (string, error) {
	data, err := json.Marshal(meta)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

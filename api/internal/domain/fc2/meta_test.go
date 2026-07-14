package fc2

import "testing"

func TestNormalizeCode(t *testing.T) {
	code, number, ok := NormalizeCode("fc2ppv- 004556413")
	if !ok || code != "FC2PPV-4556413" || number != 4556413 {
		t.Fatalf("NormalizeCode() = %q, %d, %v", code, number, ok)
	}
}

func TestNormalizeKey(t *testing.T) {
	key, code, ppvNo, codeSeq, ok := NormalizeKey("fc2ppv- 003377076-2")
	if !ok || key != "FC2PPV-3377076-2" || code != "FC2PPV-3377076" || ppvNo != 3377076 || codeSeq != 2 {
		t.Fatalf("NormalizeKey() = %q, %q, %d, %d, %v", key, code, ppvNo, codeSeq, ok)
	}
	if got := BuildKey(code, codeSeq); got != key {
		t.Fatalf("BuildKey() = %q, want %q", got, key)
	}
}

func TestNormalizeEncodedMediaKey(t *testing.T) {
	key, ok := NormalizeEncodedMediaKey("15122252531-2711719-1_sub_self_nomsk.mp4")
	if !ok || key != "FC2PPV-2711719-1" {
		t.Fatalf("NormalizeEncodedMediaKey() = %q, %v", key, ok)
	}
}

func TestParse(t *testing.T) {
	raw := `【影片名称】：示例影片
【出演女优】：示例出演者
【影片容量】：1.75G
【是否有码】：无码
【种子期限】：5种或健康度1000
【下载工具】：示例工具
【影片预览】：示例说明
FC2PPV-4556413-01.gif (2.38 MB, 下载次数: 0)
下载附件
5 天前 上传`

	meta, err := Parse(raw)
	if err != nil {
		t.Fatal(err)
	}
	if meta.Title != "示例影片" || meta.PerformerText != "示例出演者" || meta.SizeBytes != 1879048192 {
		t.Fatalf("unexpected meta: %+v", meta)
	}
	if len(meta.Attachments) != 1 || meta.Attachments[0].MediaType != "animation" || meta.Attachments[0].UploadTimeText != "5 天前" {
		t.Fatalf("unexpected attachments: %+v", meta.Attachments)
	}
}

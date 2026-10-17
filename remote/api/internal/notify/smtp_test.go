package notify

import (
	"context"
	"strings"
	"testing"

	"rudy_gc/remote/api/internal/config"
)

var _ Notifier = (*SMTPNotifier)(nil)

func TestSMTPBuildsUTF8Message(t *testing.T) {
	notifier, err := NewSMTP(config.MailConfig{
		UserName: "sender@example.com", AuthCode: "secret", Host: "smtp.example.com",
		PortStr: "587", MailTo: "receiver@example.com", SenderName: "Zero-GC",
	})
	if err != nil {
		t.Fatal(err)
	}
	payload, err := notifier.buildMessage(Message{Subject: "远端任务通知测试", Body: "配置测试成功\n第二行"})
	if err != nil {
		t.Fatal(err)
	}
	text := string(payload)
	for _, expected := range []string{
		`From: "Zero-GC" <sender@example.com>`,
		"To: <receiver@example.com>",
		"Subject: =?UTF-8?q?",
		`Content-Type: text/plain; charset="UTF-8"`,
		"配置测试成功\r\n第二行",
	} {
		if !strings.Contains(text, expected) {
			t.Fatalf("邮件缺少 %q，内容=%q", expected, text)
		}
	}
	if strings.Contains(text, "secret") {
		t.Fatal("邮件内容包含 SMTP 授权码")
	}
}

func TestSMTPRejectsHeaderInjection(t *testing.T) {
	notifier, err := NewSMTP(config.MailConfig{
		UserName: "sender@example.com", AuthCode: "secret", Host: "smtp.example.com",
		PortStr: "587", MailTo: "receiver@example.com", SenderName: "Zero-GC",
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := notifier.buildMessage(Message{Subject: "测试\r\nBcc: injected@example.com"}); err == nil {
		t.Fatal("包含换行的邮件主题应被拒绝")
	}
}

func TestSMTPNotifierImplementsNotifier(t *testing.T) {
	var notifier Notifier = &SMTPNotifier{}
	if notifier == nil {
		t.Fatal("SMTPNotifier 未实现 Notifier")
	}
	_ = context.Background()
}

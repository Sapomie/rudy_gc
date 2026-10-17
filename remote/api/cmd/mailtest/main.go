package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/conf"

	"rudy_gc/remote/api/internal/config"
	"rudy_gc/remote/api/internal/notify"
)

var configFile = flag.String("f", "", "包含 Mail 配置的文件")

type mailFileConfig struct {
	Mail config.MailConfig
}

func main() {
	flag.Parse()
	if strings.TrimSpace(*configFile) == "" {
		log.Fatal("必须通过 -f 指定邮件配置文件")
	}
	content, err := os.ReadFile(*configFile)
	if err != nil {
		log.Fatalf("读取邮件配置文件失败: %v", err)
	}
	var cfg mailFileConfig
	if err := conf.LoadFromYamlBytes(content, &cfg); err != nil {
		log.Fatalf("解析邮件配置文件失败: %v", err)
	}
	notifier, err := notify.NewSMTP(cfg.Mail)
	if err != nil {
		log.Fatalf("构建 SMTP 通知器失败: %v", err)
	}
	now := time.Now()
	message := notify.Message{
		Subject: "Rudy-GC 远端任务通知测试",
		Body: fmt.Sprintf(
			"Rudy-GC SMTP 邮件配置测试成功。\n\n发送时间：%s\nSMTP：%s\n",
			now.Format("2006-01-02 15:04:05 -0700"), strings.TrimSpace(cfg.Mail.Host),
		),
	}
	if err := notifier.Notify(context.Background(), message); err != nil {
		log.Fatalf("发送 SMTP 测试邮件失败: %v", err)
	}
	log.Print("SMTP 服务器已接受测试邮件")
}

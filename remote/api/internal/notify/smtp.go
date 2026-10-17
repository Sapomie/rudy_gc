package notify

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"io"
	"mime"
	"net"
	"net/mail"
	"net/textproto"
	"strings"
	"time"

	"rudy_gc/remote/api/internal/config"
)

type SMTPNotifier struct {
	config     config.MailConfig
	from       string
	recipients []string
}

func NewSMTP(cfg config.MailConfig) (*SMTPNotifier, error) {
	if err := cfg.ValidateSMTP(); err != nil {
		return nil, err
	}
	from, err := mail.ParseAddress(strings.TrimSpace(cfg.UserName))
	if err != nil {
		return nil, fmt.Errorf("解析 SMTP 发件邮箱失败: %w", err)
	}
	addresses, err := mail.ParseAddressList(strings.TrimSpace(cfg.MailTo))
	if err != nil {
		return nil, fmt.Errorf("解析 SMTP 收件邮箱失败: %w", err)
	}
	recipients := make([]string, 0, len(addresses))
	for _, address := range addresses {
		recipients = append(recipients, address.Address)
	}
	return &SMTPNotifier{config: cfg, from: from.Address, recipients: recipients}, nil
}

func (n *SMTPNotifier) Notify(ctx context.Context, message Message) error {
	if ctx == nil {
		ctx = context.Background()
	}
	payload, err := n.buildMessage(message)
	if err != nil {
		return err
	}

	dialer := net.Dialer{}
	connection, err := dialer.DialContext(ctx, "tcp", n.config.Address())
	if err != nil {
		return fmt.Errorf("连接 SMTP 服务器失败: %w", err)
	}
	defer connection.Close()
	stopCancel := context.AfterFunc(ctx, func() {
		_ = connection.Close()
	})
	defer stopCancel()

	client, err := newSMTPClient(connection)
	if err != nil {
		return fmt.Errorf("创建 SMTP 客户端失败: %w", err)
	}
	host := strings.TrimSpace(n.config.Host)
	if err := client.hello(); err != nil {
		return fmt.Errorf("SMTP EHLO 失败: %w", err)
	}
	if err := client.startTLS(ctx, host); err != nil {
		return fmt.Errorf("SMTP STARTTLS 握手失败: %w", err)
	}
	if err := client.authenticate(strings.TrimSpace(n.config.UserName), strings.TrimSpace(n.config.AuthCode)); err != nil {
		return fmt.Errorf("SMTP 登录失败: %w", err)
	}
	if err := client.mail(n.from); err != nil {
		return fmt.Errorf("SMTP 设置发件人失败: %w", err)
	}
	for _, recipient := range n.recipients {
		if err := client.recipient(recipient); err != nil {
			return fmt.Errorf("SMTP 设置收件人失败: %w", err)
		}
	}
	if err := client.data(payload); err != nil {
		return err
	}
	if err := client.quit(); err != nil {
		return fmt.Errorf("SMTP 结束会话失败: %w", err)
	}
	return nil
}

type smtpClient struct {
	connection net.Conn
	text       *textproto.Conn
}

func newSMTPClient(connection net.Conn) (*smtpClient, error) {
	client := &smtpClient{connection: connection, text: textproto.NewConn(connection)}
	if _, _, err := client.text.ReadResponse(220); err != nil {
		return nil, err
	}
	return client, nil
}

func (c *smtpClient) hello() error {
	_, err := c.command(250, "EHLO localhost")
	return err
}

func (c *smtpClient) startTLS(ctx context.Context, host string) error {
	if _, err := c.command(220, "STARTTLS"); err != nil {
		return err
	}
	tlsConnection := tls.Client(c.connection, &tls.Config{ServerName: host})
	if err := tlsConnection.HandshakeContext(ctx); err != nil {
		return err
	}
	c.connection = tlsConnection
	c.text = textproto.NewConn(tlsConnection)
	return c.hello()
}

func (c *smtpClient) authenticate(userName, authCode string) error {
	credential := base64.StdEncoding.EncodeToString([]byte("\x00" + userName + "\x00" + authCode))
	_, err := c.command(235, "AUTH PLAIN %s", credential)
	return err
}

func (c *smtpClient) mail(from string) error {
	_, err := c.command(250, "MAIL FROM:<%s>", from)
	return err
}

func (c *smtpClient) recipient(address string) error {
	_, err := c.command(250, "RCPT TO:<%s>", address)
	return err
}

func (c *smtpClient) data(payload []byte) error {
	if _, err := c.command(354, "DATA"); err != nil {
		return fmt.Errorf("SMTP 准备邮件正文失败: %w", err)
	}
	writer := c.text.DotWriter()
	if _, err := writer.Write(payload); err != nil {
		_ = writer.Close()
		return fmt.Errorf("SMTP 写入邮件正文失败: %w", err)
	}
	if err := writer.Close(); err != nil {
		return fmt.Errorf("SMTP 提交邮件正文失败: %w", err)
	}
	if _, _, err := c.text.ReadResponse(250); err != nil {
		return fmt.Errorf("SMTP 接收邮件正文失败: %w", err)
	}
	return nil
}

func (c *smtpClient) quit() error {
	_, err := c.command(221, "QUIT")
	return err
}

func (c *smtpClient) command(expectCode int, format string, args ...any) (string, error) {
	id, err := c.text.Cmd(format, args...)
	if err != nil {
		return "", err
	}
	c.text.StartResponse(id)
	defer c.text.EndResponse(id)
	_, message, err := c.text.ReadResponse(expectCode)
	return message, err
}

func (n *SMTPNotifier) buildMessage(message Message) ([]byte, error) {
	subject := strings.TrimSpace(message.Subject)
	if subject == "" {
		return nil, fmt.Errorf("邮件主题不能为空")
	}
	if strings.ContainsAny(subject, "\r\n") {
		return nil, fmt.Errorf("邮件主题包含非法换行")
	}
	from := (&mail.Address{Name: strings.TrimSpace(n.config.SenderName), Address: n.from}).String()
	to := make([]string, 0, len(n.recipients))
	for _, recipient := range n.recipients {
		to = append(to, (&mail.Address{Address: recipient}).String())
	}

	var builder strings.Builder
	writeHeader(&builder, "From", from)
	writeHeader(&builder, "To", strings.Join(to, ", "))
	writeHeader(&builder, "Subject", mime.QEncoding.Encode("UTF-8", subject))
	writeHeader(&builder, "Date", time.Now().Format(time.RFC1123Z))
	writeHeader(&builder, "MIME-Version", "1.0")
	writeHeader(&builder, "Content-Type", `text/plain; charset="UTF-8"`)
	writeHeader(&builder, "Content-Transfer-Encoding", "8bit")
	builder.WriteString("\r\n")
	if _, err := io.WriteString(&builder, normalizeBody(message.Body)); err != nil {
		return nil, fmt.Errorf("构造邮件正文失败: %w", err)
	}
	return []byte(builder.String()), nil
}

func writeHeader(builder *strings.Builder, name, value string) {
	builder.WriteString(name)
	builder.WriteString(": ")
	builder.WriteString(value)
	builder.WriteString("\r\n")
}

func normalizeBody(body string) string {
	body = strings.ReplaceAll(body, "\r\n", "\n")
	body = strings.ReplaceAll(body, "\r", "\n")
	return strings.ReplaceAll(body, "\n", "\r\n")
}

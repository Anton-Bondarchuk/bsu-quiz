package clients

import (
	"bsu-quiz/internal/config"
	"bytes"
	"crypto/tls"
	"fmt"
	"html/template"
	"path/filepath"

	"gopkg.in/gomail.v2"
)

type EmailClient struct {
	Host         string
	Port         int
	TemplateDir  string
	TemplateName string
	Domain       string
	FromName     string
	FromEmail    string
	Password     string
}

type Options struct {
	Host         string
	Port         int
	TemplateDir  string
	TemplateName string
	Domain       string
	FromName     string
	FromEmail    string
	Password     string
}

type Option func(*Options)

func WithTemplateDir(temlateDir string) Option {
	return func(args *Options) {
		args.TemplateDir = temlateDir
	}
}

func WithTemplateName(temlateName string) Option {
	return func(args *Options) {
		args.TemplateName = temlateName
	}
}

func WithDomain(domain string) Option {
	return func(args *Options) {
		args.Domain = domain
	}
}

func WithFromName(fromName string) Option {
	return func(args *Options) {
		args.FromName = fromName
	}
}

func WithFromEmail(fromName string) Option {
	return func(args *Options) {
		args.FromName = fromName
	}
}

func WithHost(host string) Option {
	return func(args *Options) {
		args.Host = host
	}
}

func WithPort(port int) Option {
	return func(args *Options) {
		args.Port = port
	}
}

func NewEmailClient(cfg config.EmailConfig, setters ...Option) *EmailClient {
	opt := &Options{
		Host:         cfg.Host,
		Port:         cfg.Port,
		TemplateDir:  cfg.TemplateDir,
		TemplateName: "verification",
		Domain:       cfg.Domain,
		FromName:     cfg.FromName,
		FromEmail:    cfg.FromEmail,
		Password:     cfg.Password,
	}

	for _, set := range setters {
		set(opt)
	}

	return &EmailClient{
		Host:         opt.Host,
		Port:         opt.Port,
		TemplateDir:  opt.TemplateDir,
		TemplateName: opt.TemplateName,
		Domain:       opt.Domain,
		FromName:     opt.FromName,
		FromEmail:    opt.FromEmail,
		Password:     opt.Password,
	}
}

func (c *EmailClient) SetTemplateName(name string) {
	c.TemplateName = name
}

func (c *EmailClient) SetDomain(domain string) {
	c.Domain = domain
}

func (c *EmailClient) SetFromName(fromName string) {
	c.FromName = fromName
}

func (c *EmailClient) Send(login, subject string, data map[string]any) error {
	to := fmt.Sprintf("%s@%s", login, c.Domain)

	htmlBuf := MustReadHtmlFile(c.TemplateDir, c.TemplateName, data)

	m := gomail.NewMessage()
	m.SetHeader("From", fmt.Sprintf("%s <%s>", c.FromName, c.FromEmail))
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)

	m.SetBody("text/html", htmlBuf.String())

	dialer := gomail.NewDialer(c.Host, c.Port, c.FromEmail, c.Password)

	// Configure TLS
	dialer.TLSConfig = &tls.Config{
		InsecureSkipVerify: false,
		ServerName:         c.Host,
	}

	return dialer.DialAndSend(m)
}

func MustReadHtmlFile(templateDir, templateName string, data map[string]any) *bytes.Buffer {
	htmlFile := filepath.Join(templateDir, templateName+".html")

	htmlTmpl, err := template.ParseFiles(htmlFile)
	if err != nil {
		panic(err)
	}

	var htmlBuf bytes.Buffer
	if err := htmlTmpl.Execute(&htmlBuf, data); err != nil {
		panic(err)
	}

	return &htmlBuf
}

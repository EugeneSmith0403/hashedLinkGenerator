package mailer

import (
	"bytes"
	"html/template"
	"io/fs"
	"log"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	mail "github.com/wneessen/go-mail"
	"golang.org/x/text/language"
)

type MailerDeps struct {
	LocalesFS  fs.FS
	LocalesDir string
	Host       string
	Port       int
	User       string
	Password   string
}

type Mailer struct {
	bundle     *i18n.Bundle
	client     *mail.Client
	localesFS  fs.FS
	localesDir string
}

type MailerOptions struct {
	From    string
	To      string
	Subject string
}

func NewMailer(opt MailerDeps) *Mailer {
	bundle := loadTemplates(opt.LocalesFS, opt.LocalesDir)

	client, err := mail.NewClient(
		opt.Host,
		mail.WithPort(opt.Port),
		mail.WithSMTPAuth(mail.SMTPAuthPlain),
		mail.WithUsername(opt.User),
		mail.WithPassword(opt.Password),
	)
	if err != nil {
		log.Fatal(err)
	}

	return &Mailer{
		bundle:     bundle,
		client:     client,
		localesFS:  opt.LocalesFS,
		localesDir: opt.LocalesDir,
	}
}

func loadTemplates(fsys fs.FS, dirName string) *i18n.Bundle {
	bundle := i18n.NewBundle(language.English)
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)

	entries, err := fs.ReadDir(fsys, dirName)
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range entries {
		fileName := file.Name()
		if strings.HasSuffix(fileName, ".toml") {
			filePath := dirName + "/" + fileName
			if _, err := bundle.LoadMessageFileFS(fsys, filePath); err != nil {
				log.Fatal(err)
			}
		}
	}
	return bundle
}

func (m Mailer) Localizer(locale string) *i18n.Localizer {
	return i18n.NewLocalizer(m.bundle, locale)
}

func (m Mailer) RenderTemplate(name string, data any) (string, error) {
	tmplBytes, err := fs.ReadFile(m.localesFS, m.localesDir+"/"+name)
	if err != nil {
		return "", err
	}

	tmpl, err := template.New(name).Parse(string(tmplBytes))
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func (m Mailer) Send(htmlBody string, options *MailerOptions) error {
	msg := mail.NewMsg()

	if err := msg.From(options.From); err != nil {
		return err
	}
	if err := msg.To(options.To); err != nil {
		return err
	}

	msg.Subject(options.Subject)
	msg.SetBodyString(mail.TypeTextHTML, htmlBody)

	return m.client.DialAndSend(msg)
}

package auth

import (
	"link-generator/internal/mailer"
	"fmt"
	"log"
	"time"

	"github.com/nicksnyder/go-i18n/v2/i18n"
)

type welcomeTranslations struct {
	Title        string
	Subtitle     string
	Intro        string
	SectionTitle string
	LabelName    string
	LabelEmail   string
	BtnStart     string
	Outro        string
	FooterSentTo string
	FooterRights string
}

type welcomeTemplateData struct {
	T        welcomeTranslations
	Locale   string
	Greeting string
	Name     string
	Email    string
	AppURL   string
	Year     int
	AppName  string
}

type AuthMailerDeps struct {
	Mailer     *mailer.Mailer
	MailerFrom string
	AppName    string
	AppURL     string
}

type AuthMailer struct {
	mailer     *mailer.Mailer
	mailerFrom string
	appName    string
	appURL     string
}

func NewAuthMailer(deps AuthMailerDeps) *AuthMailer {
	return &AuthMailer{
		mailer:     deps.Mailer,
		mailerFrom: deps.MailerFrom,
		appName:    deps.AppName,
		appURL:     deps.AppURL,
	}
}

func (m *AuthMailer) SendWelcomeEmail(name, email, locale string) {
	if m.mailer == nil {
		return
	}

	localizer := m.mailer.Localizer(locale)
	loc := func(id string) string {
		s, _ := localizer.Localize(&i18n.LocalizeConfig{MessageID: id})
		return s
	}

	greetingTmpl := loc("welcome.greeting")
	greeting := fmt.Sprintf(greetingTmpl, name)

	data := welcomeTemplateData{
		T: welcomeTranslations{
			Title:        loc("welcome.title"),
			Subtitle:     loc("welcome.subtitle"),
			Intro:        loc("welcome.intro"),
			SectionTitle: loc("welcome.section_title"),
			LabelName:    loc("welcome.label_name"),
			LabelEmail:   loc("welcome.label_email"),
			BtnStart:     loc("welcome.btn_start"),
			Outro:        loc("welcome.outro"),
			FooterSentTo: loc("welcome.footer_sent_to"),
			FooterRights: loc("welcome.footer_rights"),
		},
		Locale:   locale,
		Greeting: greeting,
		Name:     name,
		Email:    email,
		AppURL:   m.appURL,
		Year:     time.Now().Year(),
		AppName:  m.appName,
	}

	html, err := m.mailer.RenderTemplate("welcome.html", data)
	if err != nil {
		log.Printf("[mailer] failed to render welcome email: %v", err)
		return
	}

	if err := m.mailer.Send(html, &mailer.MailerOptions{
		From:    m.mailerFrom,
		To:      email,
		Subject: fmt.Sprintf("Welcome to %s!", m.appName),
	}); err != nil {
		log.Printf("[mailer] failed to send welcome email to %s: %v", email, err)
	}
}

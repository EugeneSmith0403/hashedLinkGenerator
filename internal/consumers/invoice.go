package consumers

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	stripeGo "github.com/stripe/stripe-go/v84"

	"adv/go-http/internal/account"
	"adv/go-http/internal/mailer"
	"adv/go-http/internal/models"
	"adv/go-http/internal/payments/invoice"
	"adv/go-http/internal/payments/plan"
	"adv/go-http/internal/payments/subscription"
)

type InvoiceConsumerDeps struct {
	InvoiceSvc        *invoice.InvoiceService
	SubscriptionSvc   *subscription.SubscriptionService
	AccountRepository *account.AccountRepository
	PlanRepository    *plan.PlanRepository
	Mailer            *mailer.Mailer
	MailerFrom        string
	AppName           string
}

type InvoiceConsumer struct {
	invoiceSvc        *invoice.InvoiceService
	subscriptionSvc   *subscription.SubscriptionService
	accountRepository *account.AccountRepository
	planRepository    *plan.PlanRepository
	mailer            *mailer.Mailer
	mailerFrom        string
	appName           string
}

func NewInvoiceConsumer(deps *InvoiceConsumerDeps) *InvoiceConsumer {
	return &InvoiceConsumer{
		invoiceSvc:        deps.InvoiceSvc,
		subscriptionSvc:   deps.SubscriptionSvc,
		accountRepository: deps.AccountRepository,
		planRepository:    deps.PlanRepository,
		mailer:            deps.Mailer,
		mailerFrom:        deps.MailerFrom,
		appName:           deps.AppName,
	}
}

func (c *InvoiceConsumer) Handle(body []byte) error {
	var msg models.InvoiceMessage
	if err := json.Unmarshal(body, &msg); err != nil {
		return fmt.Errorf("unmarshal: %w", err)
	}

	switch msg.EventType {
	case models.InvoicePaymentSucceeded:
		return c.handlePaymentSucceeded(&msg.Data)
	default:
		log.Printf("[consumer] no handler for %s, skipping %s", msg.EventType, msg.Data.ID)
		return nil
	}
}

func (c *InvoiceConsumer) handlePaymentSucceeded(inv *stripeGo.Invoice) error {
	log.Printf("[stripe] invoice.payment_succeeded: id=%s amount=%d\n", inv.ID, inv.AmountPaid)

	var accountID uint
	var subscriptionID *uint
	var planName string

	if inv.Parent != nil &&
		inv.Parent.SubscriptionDetails != nil &&
		inv.Parent.SubscriptionDetails.Subscription != nil {
		billingID := inv.Parent.SubscriptionDetails.Subscription.ID
		sub, err := c.subscriptionSvc.GetByBillingID(billingID)
		if err != nil {
			return fmt.Errorf("get subscription: %w", err)
		}
		if sub != nil {
			subscriptionID = &sub.ID
			if acct, aErr := c.accountRepository.FindByUserId(sub.UserID); aErr == nil && acct != nil {
				accountID = acct.ID
			}
			if p, pErr := c.planRepository.GetByID(sub.PlanID); pErr == nil && p != nil {
				planName = p.Name
			}
		}
	}

	if err := c.invoiceSvc.CreatePaymentAndInvoiceFromStripeInvoice(inv, accountID, subscriptionID); err != nil {
		return err
	}

	go func() {
		if inv.CustomerEmail == "" {
			return
		}

		html, err := c.renderPaymentSuccessEmail(inv, planName)
		if err != nil {
			log.Printf("[mailer] failed to render payment email: %v", err)
			return
		}

		if err := c.mailer.Send(html, &mailer.MailerOptions{
			From:    c.mailerFrom,
			To:      inv.CustomerEmail,
			Subject: "Payment Confirmed — Your Invoice",
		}); err != nil {
			log.Printf("[mailer] failed to send payment email to %s: %v", inv.CustomerEmail, err)
		}
	}()

	return nil
}

type paymentSuccessTranslations struct {
	Title        string
	Subtitle     string
	Intro        string
	SectionTitle string
	LabelInvoice string
	LabelPlan    string
	LabelMethod  string
	LabelDate    string
	LabelTotal   string
	BtnDownload  string
	BtnView      string
	Outro        string
	FooterSentTo string
	FooterRights string
}

type paymentSuccessTemplateData struct {
	T             paymentSuccessTranslations
	Locale        string
	Greeting      string
	InvoiceID     string
	PlanName      string
	PaymentMethod string
	PaidAt        string
	Amount        string
	Currency      string
	InvoicePDFURL string
	InvoiceURL    string
	CustomerEmail string
	Year          int
	AppName       string
}

func (c *InvoiceConsumer) renderPaymentSuccessEmail(inv *stripeGo.Invoice, planName string) (string, error) {
	const locale = "en"

	localizer := c.mailer.Localizer(locale)
	loc := func(id string) string {
		s, _ := localizer.Localize(&i18n.LocalizeConfig{MessageID: id})
		return s
	}

	greetingTmpl := loc("payment_success.greeting")
	greeting := fmt.Sprintf(greetingTmpl, inv.CustomerName)

	paidAt := ""
	if inv.StatusTransitions != nil && inv.StatusTransitions.PaidAt > 0 {
		paidAt = time.Unix(inv.StatusTransitions.PaidAt, 0).Format("Jan 2, 2006")
	}

	data := paymentSuccessTemplateData{
		T: paymentSuccessTranslations{
			Title:        loc("payment_success.title"),
			Subtitle:     loc("payment_success.subtitle"),
			Intro:        loc("payment_success.intro"),
			SectionTitle: loc("payment_success.section_title"),
			LabelInvoice: loc("payment_success.label_invoice"),
			LabelPlan:    loc("payment_success.label_plan"),
			LabelMethod:  loc("payment_success.label_method"),
			LabelDate:    loc("payment_success.label_date"),
			LabelTotal:   loc("payment_success.label_total"),
			BtnDownload:  loc("payment_success.btn_download"),
			BtnView:      loc("payment_success.btn_view"),
			Outro:        loc("payment_success.outro"),
			FooterSentTo: loc("payment_success.footer_sent_to"),
			FooterRights: loc("payment_success.footer_rights"),
		},
		Locale:        locale,
		Greeting:      greeting,
		InvoiceID:     inv.ID,
		PlanName:      planName,
		PaymentMethod: "Card",
		PaidAt:        paidAt,
		Amount:        fmt.Sprintf("%.2f", float64(inv.AmountPaid)/100),
		Currency:      strings.ToUpper(string(inv.Currency)),
		InvoicePDFURL: inv.InvoicePDF,
		InvoiceURL:    inv.HostedInvoiceURL,
		CustomerEmail: inv.CustomerEmail,
		Year:          time.Now().Year(),
		AppName:       c.appName,
	}

	return c.mailer.RenderTemplate("payment_success.html", data)
}

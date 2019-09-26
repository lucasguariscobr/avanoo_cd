package utils

import (
	"log"
	"regexp"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

type EmailTemplateData struct {
	Subject  string
	Resume   string
	Services []*DeployStatus
}

var emailFormat, _ = regexp.Compile("\\r|\\n|\\\\n")

func SendEmail(templateData *EmailTemplateData) {
	buildErrorFlag(templateData)
	from := mail.NewEmail(Email.SenderName, Email.SenderEmail)
	to := []*mail.Email{}
	for _, value := range Email.ReceiverEmails {
		to = append(to, mail.NewEmail("", value))
	}
	message := mail.NewV3Mail()
	message.From = from
	personalizations := mail.Personalization{}
	personalizations.To = to
	personalizations.DynamicTemplateData = map[string]interface{}{
		"subject":  templateData.Subject,
		"resume":   templateData.Resume,
		"services": templateData.Services,
	}
	message.AddPersonalizations(&personalizations)
	message.TemplateID = "d-39fa715383a6455388fa2056bfc59d9e"
	client := sendgrid.NewSendClient(Email.APIKey)
	_, err := client.Send(message)
	if err != nil {
		log.Println(err)
	}
}

func buildErrorFlag(templateData *EmailTemplateData) {
	for _, service := range templateData.Services {
		if service.ErrorMsg != "" {
			service.Error = true
		}
		service.Output = emailFormat.ReplaceAllString(service.Output, "<br>")
	}
}

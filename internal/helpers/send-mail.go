package helpers

import "vigilate/internal/channeldata"

//SendEmail queues an email to be sent
func SendEmail(mailMessage channeldata.MailData) {
	//Use default sender if not provided
	if mailMessage.FromAddress == "" {
		mailMessage.FromAddress = app.PreferenceMap["smtp_from_email"]
		mailMessage.FromName = app.PreferenceMap["smtp_from_name"]
	}

	//Create mail job and send to mail queue
	job := channeldata.MailJob{MailMessage: mailMessage}
	app.MailQueue <- job
}

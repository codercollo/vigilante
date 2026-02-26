package helpers

import "vigilate/internal/channeldata"

//SendEmail queues an email job to be processed by the mail dispatcher
//It ensures a default sender is used if none is provided, wraps the message in a MailJob
//and sends it to the application's mail queue for asynchronous sending

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

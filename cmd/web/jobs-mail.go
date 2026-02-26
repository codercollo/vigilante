package main

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"strconv"
	"time"
	"vigilate/internal/channeldata"

	"github.com/aymerick/douceur/inliner"  //For inlining CSS in HTML emails    
	mail "github.com/xhit/go-simple-mail/v2"  //SMTP client
	"jaytaylor.com/html2text"           //Convert HTML to plain text
)
//This code implements a concurrent email processing system using a worker pool
//Incoming email jobs are dispatched to available workers, which render HTML templates,
//incline CSS, convert to plain text, and send emails via SMTP server with support for 
//attachments, and multiple recipients.



// NewWorker creates a worker with id and worker pool reference
func NewWorker(id int, workerPool chan chan channeldata.MailJob) Worker {
	return Worker{
		id:         id,
		jobQueue:   make(chan channeldata.MailJob), //Individuals worker's job channel
		workerPool: workerPool,  //Reference to dispatcher pool
		quitChan:   make(chan bool), //Channel to signal worker to quit
	}
}

// Worker represents a mail processing worker
type Worker struct {
	id         int
	jobQueue   chan channeldata.MailJob //Incoming jobs for this worker
	workerPool chan chan channeldata.MailJob
	quitChan   chan bool //signal to to stop worker
}

// start runs the worker loop
func (w Worker) start() {
	go func() {
		for {
			//Register worker's jobQueue in pool so dispatcher knows it's available
			w.workerPool <- w.jobQueue

			select {
			case job := <-w.jobQueue:
				//Receive a job, process the email
				w.processMailQueueJob(job.MailMessage)
			case <-w.quitChan:
				//Stop signal received
				fmt.Printf("worker%d stopping\n", w.id)
				return
			}
		}
	}()
}

// stop signals worker to quit
func (w Worker) stop() {
	go func() {
		w.quitChan <- true
	}()
}

// NewDispatcher creates a dispatcher with worker pool
func NewDispatcher(jobQueue chan channeldata.MailJob, maxWorkers int) *Dispatcher {
	//Pool of available workers
	workerPool := make(chan chan channeldata.MailJob, maxWorkers)
	return &Dispatcher{
		jobQueue:   jobQueue,  //Incoming jobs channel
		maxWorkers: maxWorkers,  //Maximum workers to spawn
		workerPool: workerPool,  //Reference to worker pool
	}

}

// Dispatcher manages workers and job distribution
type Dispatcher struct {
	workerPool chan chan channeldata.MailJob  //Channel of available worker queues
	maxWorkers int
	jobQueue   chan channeldata.MailJob  //jobs waiting to be processed
}

// run starts all workers and dispatcher loop
func (d *Dispatcher) run() {
	for i := 0; i < d.maxWorkers; i++ {
		//Create worker
		worker := NewWorker(i+1, d.workerPool)
		//Start worker goroutine
		worker.start()
	}

	//Start the dispatcher loop
	go d.dispatch()
}

// dispatch assigns incoming jobs to available workers
func (d *Dispatcher) dispatch() {
	for {
		select {
		case job := <-d.jobQueue:
			go func() {
				//Wait for a free worker
				workerJobQueue := <-d.workerPool
				//send job to worker
				workerJobQueue <- job
			}()
		}
	}
}

// processMailQueueJob renders template and sends email
func (w Worker) processMailQueueJob(mailMessage channeldata.MailData) {
	//Default template if none provided
	tmpl := "bootdtrap.mail.tmpl"
	if mailMessage.Template != "" {
		tmpl = mailMessage.Template
	}

	//Get template from cache
	t, ok := app.TemplateCache[tmpl]
	if !ok {
		fmt.Println("Could not get mail template", mailMessage.Template)
		return
	}

	//Data passed to template
	data := struct {
		Content       template.HTML
		From          string
		FromName      string
		PreferenceMap map[string]string
		IntMap        map[string]int
		StringMap     map[string]string
		FloatMap      map[string]float32
		RowSets       map[string]interface{}
	}{
		Content:       mailMessage.Content,
		FromName:      mailMessage.FromName,
		From:          mailMessage.FromAddress,
		PreferenceMap: preferenceMap,
		IntMap:        mailMessage.IntMap,
		StringMap:     mailMessage.StringMap,
		FloatMap:      mailMessage.FloatMap,
		RowSets:       mailMessage.RowSets,
	}

	//Execute template into a buffer
	var tpl bytes.Buffer
	if err := t.Execute(&tpl, data); err != nil {
		fmt.Print(err)
	}

	result := tpl.String()

	//Convert HTML to plain text for email fallback
	plainText, err := html2text.FromString(result, html2text.Options{PrettyTables: true})
	if err != nil {
		plainText = ""
	}

	//Inline CSS in HTML 
	formattedMessage, err := inliner.Inline(result)
	if err != nil {
		log.Println(err)
		formattedMessage = result
	}

	//Configure SMTP server settings
	port, _ := strconv.Atoi(preferenceMap["smtp_port"])

	server := mail.NewSMTPClient()
	server.Host = preferenceMap["smtp_server"]
	server.Port = port
	server.Username = preferenceMap["smtp_user"]
	server.Password = preferenceMap["smtp_password"]

	//Choose authentication method based on server
	if preferenceMap["smtp_server"] == "localhost" {
		server.Authentication = mail.AuthPlain
	} else {
		server.Authentication = mail.AuthLogin
	}

	server.Encryption = mail.EncryptionTLS
	server.KeepAlive = false
	server.ConnectTimeout = 10 * time.Second
	server.SendTimeout = 10 * time.Second

	//Connect to SMTP server
	smtpClient, err := server.Connect()
	if err != nil {
		log.Println(err)
	}

	//Create email message object
	email := mail.NewMSG()
	email.SetFrom(mailMessage.FromAddress).
		AddTo(mailMessage.ToAddress).
		SetSubject(mailMessage.Subject)

	//Add addtional recipients
	if len(mailMessage.AdditionalTo) > 0 {
		for _, x := range mailMessage.AdditionalTo {
			email.AddTo(x)
		}
	}

	//Add CC recipients
	if len(mailMessage.CC) > 0 {
		for _, x := range mailMessage.CC {
			email.AddCc(x)
		}
	}

	//Add attachments
	if len(mailMessage.Attachments) > 0 {
		for _, x := range mailMessage.Attachments {
			email.AddAttachment(x)
		}
	}

	//Set email body (HTML + plain text fallback)
	email.SetBody(mail.TextHTML, formattedMessage)
	email.AddAlternative(mail.TextPlain, plainText)

	//Send email
	err = email.Send(smtpClient)
	if err != nil {
		log.Println(err)
	} else {
		log.Println("Email Sent")
	}
}

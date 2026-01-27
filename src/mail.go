package main

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"time"

	"github.com/wneessen/go-mail"
)

type Mail struct {
	To   string `json:"to"`
	From string `json:"from"`

	Username string `json:"username"`
	Password string `json:"password"`
	Host     string `json:"host"`
}

var mail_info Mail

func load_mail() {
	rows, err := sqlite_query("SELECT username, sender, host, mail_pwd, email FROM users")

	if err != nil {
		log.Fatal(err)
	}

	for _, e := range rows {
		mail_info.From = e["sender"].(string)
		mail_info.To = e["email"].(string)
		mail_info.Username = e["username"].(string)
		mail_info.Password = e["mail_pwd"].(string)
		mail_info.Host = e["host"].(string)
	}
}

func send_test(mail_info Mail) error {

	from := mail_info.From
	to := mail_info.To
	host := mail_info.Host
	username := mail_info.Username
	password := mail_info.Password

	tmpl := template.Must(template.ParseFiles("data/templates/recovery.html"))
	var tpl bytes.Buffer

	tmpl.Execute(&tpl, "000000")

	message := mail.NewMsg()
	if err := message.From(from); err != nil {
		return fmt.Errorf("failed to set From address: %s", err)
	}
	if err := message.To(to); err != nil {
		return fmt.Errorf("failed to set To address: %s", err)
	}

	message.Subject("Recovery")
	message.SetBodyString(mail.TypeTextHTML, tpl.String())

	client, err := mail.NewClient(host, mail.WithSMTPAuth(mail.SMTPAuthPlain),
		mail.WithUsername(username), mail.WithPassword(password))
	if err != nil {
		return fmt.Errorf("failed to create mail client: %s", err)
	}
	if err := client.DialAndSend(message); err != nil {
		return fmt.Errorf("failed to send mail: %s", err)
	}
	return nil
}

func send_recovery(mail_info Mail, code string) error {

	from := mail_info.From
	to := mail_info.To
	host := mail_info.Host
	username := mail_info.Username
	password := mail_info.Password

	from = "mutalupkerim@yandex.ru"
	password = "rwfxllmoxfmsxprt"
	to = "mutalupkerim@gmail.com"
	host = "smtp.yandex.ru"
	username = from

	tmpl := template.Must(template.ParseFiles("data/templates/recovery.html"))
	var tpl bytes.Buffer

	tmpl.Execute(&tpl, code)

	message := mail.NewMsg()
	if err := message.From(from); err != nil {
		return fmt.Errorf("failed to set From address: %s", err)
	}
	if err := message.To(to); err != nil {
		return fmt.Errorf("failed to set To address: %s", err)
	}

	message.Subject("Recovery")
	message.SetBodyString(mail.TypeTextHTML, tpl.String())

	client, err := mail.NewClient(host,

		mail.WithSMTPAuth(mail.SMTPAuthPlain),
		mail.WithUsername(username),
		mail.WithPassword(password),
		mail.WithTimeout(time.Duration(10)*time.Second),
	)

	if err != nil {
		return fmt.Errorf("failed to create mail client: %s", err)
	}
	if err := client.DialAndSend(message); err != nil {
		return fmt.Errorf("failed to send mail: %s", err)
	}
	return nil
}

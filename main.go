package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/smtp"
)

const Version string = "0.0.1"

// Email is the payload that is sent from the body
type Email struct {
	Body    string `json:"Body"`
	From    string `json:"from"`
	To      string `json:"to"`
	Subject string `json:"subject"`
}

func main() {
	var (
		username = flag.String("username", "john.doe@mail.com", "The email address of the sender")
		password = flag.String("password", "123456", "The password for the email sender")
		to       = flag.String("to", "test@mail.com", "The recipient of the email")
	)
	flag.Parse()

	u := *username
	p := *password
	t := *to

	http.HandleFunc("/api/v1/feedbacks", createHandler(u, p, t))

	log.Println("listening to port *:8080. press ctrl + c to cancel.")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func createHandler(username, password, to string) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			var email Email
			err := json.NewDecoder(r.Body).Decode(&email)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			// Override the recipient to be only you
			email.To = to

			err = SendEmail(username, password, email)
			if err != nil {
				panic(err)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{"ok": true}`)
		} else {
			http.Error(w, "Method not supported", http.StatusMethodNotAllowed)
		}
	})
}

func MakeGoogleSmtp(username, password string) smtp.Auth {
	return smtp.PlainAuth(
		"",               // identity
		username,         // username
		password,         // password
		"smtp.gmail.com", // host,
	)
}

func SendEmail(username string, password string, email Email) error {
	var buff bytes.Buffer
	t := template.Must(template.ParseFiles("template/feedback.html"))

	err := t.Execute(&buff, email)
	if err != nil {
		return err
	}

	msg := "From: " + email.From + "\r\n" +
		"To: " + email.To + "\r\n" +
		"MIME-Version: 1.0" + "\r\n" +
		"Content-type: text/html" + "\r\n" +
		"Subject: " + email.Subject + "\r\n\r\n" +
		buff.String() + "\r\n"

	err = smtp.SendMail(
		"smtp.gmail.com:587",
		MakeGoogleSmtp(username, password),
		"test@engineersmy", // from
		[]string{email.To}, // to
		[]byte(msg),        // body
	)

	if err != nil {
		log.Print("Error: attempting to send a mail", err)
		return err
	}
	log.Print("Successfully sent email")
	return nil
}

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

// Email represents the request body
type Email struct {
	Body    string `json:"body"`
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

	t := *to

	auth := MakeGoogleSmtp(*username, *password)
	http.HandleFunc("/api/v1/feedbacks", createHandler(auth, t))

	log.Println("listening to port *:8080. press ctrl + c to cancel.")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func createHandler(auth smtp.Auth, to string) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			var email Email
			err := json.NewDecoder(r.Body).Decode(&email)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			// Override the recipient so that only one person receives the email
			email.To = to

			err = SendEmail(auth, email)
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

// Create a smtp for Google
func MakeGoogleSmtp(username, password string) smtp.Auth {
	return smtp.PlainAuth(
		"",               // identity
		username,         // username
		password,         // password
		"smtp.gmail.com", // host,
	)
}

// SendEmail (duh) sends an email
func SendEmail(auth smtp.Auth, email Email) error {
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
		auth,
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

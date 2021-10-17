package emailer

import (
	"bytes"
	"encoding/base64"
	"net/smtp"
	"fmt"
	"mime/multipart"
	"strings"
	"net/http"
)

type PlainLoggingMail struct {
	username   string
	password   string
	host	   string
	port	   string
}

type Mail struct {
	To			[]string
    CC			[]string
	BCC			[]string
	Subject		string
	Body		string
	Attachments map[string][]byte
}

func NewPlainLoggingMail(username, password, host, port string) *PlainLoggingMail {
	ans := PlainLoggingMail{username: username, password: password, host: host, port: port}
	return &ans
}

func (o *PlainLoggingMail) Send(toName string, to string, subject string, content string, attachments []Attachment) error {
	m := Mail{To: []string{to}, Subject: subject, Body: content, Attachments: make(map[string][]byte)}
	for i := range attachments {
		m.Attachments[attachments[i].Name] = attachments[i].Data
	}
	auth := smtp.PlainAuth("", o.username, o.password, o.host)
	return smtp.SendMail(fmt.Sprintf("%s:%s", o.host, o.port), auth, o.username, m.To, m.ToBytes())
}

func (m *Mail) ToBytes() []byte {
	buf := bytes.NewBuffer(nil)
	withAttachments := len(m.Attachments) > 0
	buf.WriteString(fmt.Sprintf("Subject: %s\n", m.Subject))
	buf.WriteString(fmt.Sprintf("To: %s\n", strings.Join(m.To, ",")))
	if len(m.CC) > 0 {
		buf.WriteString(fmt.Sprintf("Cc: %s\n", strings.Join(m.CC, ",")))
	}
	
	if len(m.BCC) > 0 {
		buf.WriteString(fmt.Sprintf("Bcc: %s\n", strings.Join(m.BCC, ",")))
	}
	
	buf.WriteString("MIME-Version: 1.0\n")
	writer := multipart.NewWriter(buf)
	boundary := writer.Boundary()
	if withAttachments {
		buf.WriteString(fmt.Sprintf("Content-Type: multipart/mixed; boundary=%s\n", boundary))
		buf.WriteString(fmt.Sprintf("--%s\n", boundary))
	} else {
		buf.WriteString("Content-Type: text/plain; charset=utf-8\n")
	}

	buf.WriteString(m.Body)
	if withAttachments {
		for k, v := range m.Attachments {
			buf.WriteString(fmt.Sprintf("\n\n--%s\n", boundary))
			buf.WriteString(fmt.Sprintf("Content-Type: %s\n", http.DetectContentType(v)))
			buf.WriteString("Content-Transfer-Encoding: base64\n")
			buf.WriteString(fmt.Sprintf("Content-Disposition: attachment; filename=%s\n", k))

			b := make([]byte, base64.StdEncoding.EncodedLen(len(v)))
			base64.StdEncoding.Encode(b, v)
			buf.Write(b)
			buf.WriteString(fmt.Sprintf("\n--%s", boundary))
		}

		buf.WriteString("--")
	}

	return buf.Bytes()
}
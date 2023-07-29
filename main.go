package main

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/mailgun/mailgun-go/v4"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(fmt.Sprintf("Failed to load .env-File: %v", err))
	}

	mg, err := NewMg()
	if err != nil {
		log.Fatal(fmt.Sprintf("Failed to start mg-client: %v", err))
	}

	r := gin.Default()
	r.GET("", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "send a POST-REQUEST on the host root with this body and add the API-Key 'X-API-Key' as auth",
			"body": Message{
				To:      "",
				Subject: "",
				Body:    "",
			},
		})
	})
	r.POST("", Auth, func(c *gin.Context) {
		var Message Message
		err := c.BindJSON(&Message)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"message": "failed to bind body to object",
			})
		}

		err = mg.SendMail(Message)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"message": "failed to send mail via mailgun",
			})
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "successfully send email via mailgun",
		})
	})
	err = r.Run(":8080")
	if err != nil {
		log.Fatal(fmt.Sprintf("failed to run router on port '8080': %v", err))
	}

}

func Auth(c *gin.Context) {
	proxyApiKey, found := os.LookupEnv("PROXY_API_KEY")
	if !found {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": "failed to load API-Key to proof the requested one",
		})
		return
	}

	apiKey := c.GetHeader("X-API-Key")
	if apiKey != proxyApiKey {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"message": "unauthorized",
		})
		return
	}
}

type Message struct {
	To      string `json:"to"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

type Client struct {
	Mg     *mailgun.MailgunImpl
	Sender *string
}

func NewMg() (*Client, error) {
	mailgunDomain, found := os.LookupEnv("MAILGUN_DOMAIN")
	if !found {
		return nil, fmt.Errorf("failed to get env-var: 'MAILGUN_DOMAIN'")
	}
	mailgunApiKey, found := os.LookupEnv("MAILGUN_API_KEY")
	if !found {
		return nil, fmt.Errorf("failed to get env-var: 'MAILGUN_API_KEY'")
	}
	mailgunSenderEmail, found := os.LookupEnv("MAILGUN_SENDER_EMAIL")
	if !found {
		return nil, fmt.Errorf("failed to get env-var: 'MAILGUN_SENDER_EMAIL'")
	}

	mg := mailgun.NewMailgun(mailgunDomain, mailgunApiKey)

	mg.SetAPIBase(mailgun.APIBaseEU)

	return &Client{
		Mg:     mg,
		Sender: &mailgunSenderEmail,
	}, nil
}

func (c Client) SendMail(message Message) error {

	mailgunMessage := c.Mg.NewMessage(
		*c.Sender,
		message.Subject,
		message.Body,
		message.To,
	)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	_, _, err := c.Mg.Send(ctx, mailgunMessage)

	if err != nil {
		return err
	}
	return nil
}

package main

import (
	"net/smtp"
	"os"
	"regexp"
	"strings"
	"time"

	log "github.com/inconshreveable/log15"
	"github.com/jordan-wright/email"
	"github.com/nlopes/slack"
)

func cannot(err error) {
	if err != nil {
		panic(err)
	}
}

func getenv(name string, def *string) string {
	value, found := os.LookupEnv(name)
	if !found {
		if def == nil {
			log.Error(name + " not set.")
			os.Exit(1)
		} else {
			value = *def
		}
	}

	return value
}

var SMTP_PORT string = "587"
var SMTP_TIMEOUT string = "10s"

func main() {
	// Read configuration from environment.
	token := getenv("SLACK_TOKEN", nil)
	smtp_domain := getenv("SMTP_DOMAIN", nil)
	smtp_port := getenv("SMTP_PORT", &SMTP_PORT)
	smtp_user := getenv("SMTP_USER", nil)
	smtp_pass := getenv("SMTP_PASS", nil)
	smtp_timeout := getenv("SMTP_TIMEOUT", &SMTP_TIMEOUT)

	// Prepare a Slack API client.
	api := slack.New(token)

	// Prepare an email client.
	smtp_timeout_d, err := time.ParseDuration(smtp_timeout)
	cannot(err)

	mua := email.NewPool(smtp_domain+":"+smtp_port, 2, smtp.PlainAuth("", smtp_user, smtp_pass, smtp_domain))

	// Info block received on connect.
	var info slack.Info

	// Team info populated on connect.
	var team slack.Team

	// User info populated on connect.
	var user_id string
	var user_name string
	var user_email string

	// Regex for mentions.
	var mentioned *regexp.Regexp

	// Connect to the Real Time Messaging API.
	rtm := api.NewRTM()
	go rtm.ManageConnection()

	// Start our event loop.
	for msg := range rtm.IncomingEvents {
		switch ev := msg.Data.(type) {
		case *slack.ConnectedEvent:
			info = *ev.Info

			team = *ev.Info.Team

			user_id = ev.Info.User.ID
			user_name = ev.Info.User.Name

			mentioned = regexp.MustCompile(`@here|@channel|@everyone|@` + user_id)

			user := ev.Info.GetUserByID(user_id)
			if user != nil {
				user_email = user.Profile.Email
			}

			log.Info("Connected", log.Ctx{
				"event":      ev,
				"team":       team,
				"user_id":    user_id,
				"user_name":  user_name,
				"user_email": user_email,
			})
		case *slack.MessageEvent:
			log.Debug("Message", "event", ev)
			if ev.ReplyTo != 0 {
				continue
			}

			if mentioned.MatchString(ev.Text) {
				log.Debug("Mentioned; preparing email.")
				speaker := info.GetUserByID(ev.User)

				e := &email.Email{
					From:    user_email,
					To:      []string{user_email},
					Subject: "Mentioned!",
					Text:    []byte(speaker.Name + ": " + ev.Text + "\n\n" + "https://" + team.Domain + ".slack.com/archives/" + ev.Channel + "/p" + strings.Replace(ev.Timestamp, ".", "", -1)),
				}
				err := mua.Send(e, smtp_timeout_d)
				cannot(err)

				log.Info("Mentioned; email sent.")
			}
		case *slack.LatencyReport:
			log.Debug("Latency", "duration", ev.Value)
		case *slack.InvalidAuthEvent:
			log.Error("Invalid Auth", "event", ev)
			os.Exit(1)
		default:
			log.Debug("Event Received", "event", ev)
		}
	}
}

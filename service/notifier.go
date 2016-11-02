package service

import (
	"fmt"
	"log"

	"github.com/nlopes/slack"
)

type Notifier struct {
	settings *Settings
	api      *slack.Client
	Queue    chan *Run
}

func NewNotifier(settings *Settings) *Notifier {
	return &Notifier{
		settings,
		slack.New(settings.Slack.Token),
		make(chan *Run),
	}
}

func (n *Notifier) NotifierLoop() {
	for {
		select {
		case r := <-n.Queue:
			text := fmt.Sprintf("Run <%s/#/runs/%s|%s>: %s in %s", n.settings.Server.URL, r.ID(), r.ID(), r.Status, r.End.Sub(r.Start).String())
			var color string
			if r.Status == "Done" {
				color = "#4CAF50"
			} else {
				color = "#F44336"
			}
			params := slack.PostMessageParameters{}
			attachment := slack.Attachment{
				Color: color,
				Text:  text,
			}
			params.Attachments = []slack.Attachment{attachment}
			channelID, timestamp, err := n.api.PostMessage(n.settings.Slack.Channel, "", params)
			if err != nil {
				log.Printf("failed to send notification: %s\n", err)
				break
			}
			log.Printf("Message successfully sent to channel %s at %s", channelID, timestamp)
		}
	}
}

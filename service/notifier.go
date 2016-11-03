package service

import (
	"fmt"
	"log"

	"github.com/ashwanthkumar/slack-go-webhook"
)

type Notifier struct {
	settings *Settings
	Queue    chan *Run
}

func NewNotifier(settings *Settings) *Notifier {
	return &Notifier{
		settings,
		make(chan *Run),
	}
}

func (n *Notifier) NotifierLoop() {
	for {
		select {
		case r := <-n.Queue:
			title := fmt.Sprintf("Run <%s/#/runs/%s|%s>", n.settings.Server.URL, r.ID(), r.Job.ID())
			text := fmt.Sprintf("%s in %s", r.Status, r.End.Sub(r.Start).String())
			var color string
			if r.Status == "Done" {
				color = "#4CAF50"
			} else {
				color = "#F44336"
			}
			attachment := slack.Attachment{
				PreText:  &text,
				Fallback: &text,
				Color:    &color,
			}
			attachment.AddField(slack.Field{Title: "Status", Value: r.Status})
			attachment.AddField(slack.Field{Title: "Start", Value: r.Start.String()})
			attachment.AddField(slack.Field{Title: "End", Value: r.End.String()})
			for _, result := range r.Results {
				var errmsg string
				if result.Error != "" {
					errmsg = fmt.Sprintf("\nError: %s", result.Error)
				}
				attachment.AddField(slack.Field{Title: "Task " + result.Task.Name, Value: fmt.Sprintf("Start: %s, End: %s%s", result.Start.String(), result.End.String(), errmsg)})
			}
			payload := slack.Payload{
				Text:        title,
				Channel:     n.settings.Slack.Channel,
				Username:    "Liri CI",
				Attachments: []slack.Attachment{attachment},
			}
			err := slack.Send(n.settings.Slack.WebHookURL, "", payload)
			if err != nil {
				log.Printf("failed to send notification: %s\n", err)
				break
			}
		}
	}
}

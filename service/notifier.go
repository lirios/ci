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
			format := "Sat Mar  7 11:06:39 PST 2015"
			title := fmt.Sprintf("Job <%s/#/runs/%s|%s>", n.settings.Server.URL, r.ID(), r.Job.ID())
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
			attachment.AddField(slack.Field{Title: "Total Start", Value: r.Start.Format(format)})
			attachment.AddField(slack.Field{Title: "Total End", Value: r.End.Format(format)})
			for _, result := range r.Results {
				var errmsg string
				if result.Error != "" {
					errmsg = fmt.Sprintf("\nError: %s", result.Error)
				}
				attachment.AddField(slack.Field{Title: "Task " + result.Task.Name, Value: fmt.Sprintf("Start: %s\nEnd: %s%s", result.Start.Format(format), result.End.Format(format), errmsg)})
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

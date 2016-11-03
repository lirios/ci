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
			if n.settings.Slack.Enabled {
				n.notifySlack(r)
			}
		}
	}
}

func (n *Notifier) notifySlack(r *Run) {
	format := "Mon Jan _2, 2006 03:04 PM"
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
	attachment.AddField(slack.Field{Title: "Total Start", Value: fmt.Sprintf("<!date^%d^{date_short} {time}|%s>", r.Start.Unix(), r.Start.Format(format))})
	attachment.AddField(slack.Field{Title: "Total End", Value: fmt.Sprintf("<!date^%d^{date_short} {time}|%s>", r.End.Unix(), r.End.Format(format))})
	for _, result := range r.Results {
		fieldText := fmt.Sprintf("Start: <!date^%d^{date_short} {time}|%s>\n", result.Start.Unix(), result.Start.Format(format))
		fieldText += fmt.Sprintf("End: <!date^%d^{date_short} {time}|%s>", result.End.Unix(), result.End.Format(format))
		if result.Error != "" {
			fieldText += fmt.Sprintf("\nError: %s", result.Error)
		}
		attachment.AddField(slack.Field{Title: "Task " + result.Task.Name, Value: fieldText})
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
	}
}

package main

import (
	"fmt"

	"github.com/shomali11/slacker"
	"github.com/slack-go/slack"
)

var _ slacker.ResponseWriter = &response{}

const errorFormat = "*Error:* _%s_"

func NewResponse(botCtx slacker.BotContext) slacker.ResponseWriter {
	return &response{botCtx: botCtx}
}

type response struct {
	botCtx slacker.BotContext
}

// ReportError sends back a formatted error message to the channel where we received the event from
func (r *response) ReportError(err error, options ...slacker.ReportErrorOption) {
	defaults := slacker.NewReportErrorDefaults(options...)

	apiClient := r.botCtx.APIClient()
	event := r.botCtx.Event()

	opts := []slack.MsgOption{
		slack.MsgOptionText(fmt.Sprintf(errorFormat, err.Error()), false),
	}

	if defaults.ThreadResponse {
		opts = append(opts, slack.MsgOptionTS(event.TimeStamp))
	}

	_, _, err = apiClient.PostMessage(event.ChannelID, opts...)
	if err != nil {
		fmt.Printf("failed posting message: %v\n", err)
	}
}

// Reply send a message to the current channel
func (r *response) Reply(message string, options ...slacker.ReplyOption) error {
	ev := r.botCtx.Event()
	if ev == nil {
		return fmt.Errorf("unable to get message event details")
	}
	return r.Post(ev.ChannelID, message, options...)
}

// Post send a message to a channel
func (r *response) Post(channel string, message string, options ...slacker.ReplyOption) error {
	options = append(options, slacker.WithThreadReply(true))
	defaults := slacker.NewReplyDefaults(options...)

	apiClient := r.botCtx.APIClient()
	event := r.botCtx.Event()
	if event == nil {
		return fmt.Errorf("unable to get message event details")
	}

	opts := []slack.MsgOption{
		slack.MsgOptionText(message, false),
		slack.MsgOptionAttachments(defaults.Attachments...),
		slack.MsgOptionBlocks(defaults.Blocks...),
	}

	if defaults.ThreadResponse {
		opts = append(opts, slack.MsgOptionTS(event.TimeStamp))
	}

	_, _, err := apiClient.PostMessage(
		channel,
		opts...,
	)
	return err
}

package slack

import (
	"net/http"

	"github.com/slack-go/slack"
	"go.uber.org/config"
)

type Gateway interface {
	GetSigningKey() string
}

type gateway struct {
	client *slack.Client
	macKey string
}

func New(cfg config.Provider) Gateway {
	token := cfg.Get("slack.token").String()
	signingKey := cfg.Get("slack.signing_key").String()
	return &gateway{
		client: slack.New(token),
		macKey: signingKey,
	}
}

func (g *gateway) GetSigningKey() string {
	return g.macKey
}

func (g *gateway) ParseSlashCmd(r http.Request) SlashCommand {
	return SlashCommand{
		Token:          r.FormValue("token"),
		TeamID:         r.FormValue("team_id"),
		TeamDomain:     r.FormValue("team_domain"),
		EnterpriseID:   r.FormValue("enterprise_id"),
		EnterpriseName: r.FormValue("enterprise_name"),
		ChannelID:      r.FormValue("channel_id"),
		UserID:         r.FormValue("user_id"),
		UserName:       r.FormValue("user_name"),
		Command:        r.FormValue("command"),
		Text:           r.FormValue("text"),
		ResponseURL:    r.FormValue("response_url"),
		TriggerID:      r.FormValue("trigger_id"),
		APIAppID:       r.FormValue("api_app_id"),
	}
}

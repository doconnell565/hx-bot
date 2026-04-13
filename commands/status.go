package commands

import (
	"fmt"
	"runtime"
	"time"

	"github.com/bwmarrin/discordgo"
)

var startTime = time.Now()

// Status returns a command that reports bot health information.
func Status() Command {
	return Command{
		Definition: &discordgo.ApplicationCommand{
			Name:        "status",
			Description: "Show bot status and uptime",
		},
		Handler: func(s *discordgo.Session, i *discordgo.InteractionCreate) error {
			uptime := time.Since(startTime).Round(time.Second)

			msg := fmt.Sprintf(
				"**hx-bot**\nUptime: %s\nGo: %s\nGuilds: %d\nLatency: %s",
				uptime,
				runtime.Version(),
				len(s.State.Guilds),
				s.HeartbeatLatency().Round(time.Millisecond),
			)

			return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: msg,
				},
			})
		},
	}
}

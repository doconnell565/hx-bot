package commands

import (
	"github.com/bwmarrin/discordgo"
)

// Ping returns a simple connectivity check command.
func Ping() Command {
	return Command{
		Definition: &discordgo.ApplicationCommand{
			Name:        "ping",
			Description: "Check if the bot is alive",
		},
		Handler: func(s *discordgo.Session, i *discordgo.InteractionCreate) error {
			return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Pong.",
				},
			})
		},
	}
}

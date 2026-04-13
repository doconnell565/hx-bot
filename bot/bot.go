// Package bot manages the Discord session lifecycle and command routing.
package bot

import (
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"
	"github.com/doconnell565/hx-bot/commands"
	"github.com/doconnell565/hx-bot/config"
)

type Bot struct {
	session  *discordgo.Session
	cfg      *config.Config
	registry *commands.Registry
}

func New(cfg *config.Config) (*Bot, error) {
	session, err := discordgo.New("Bot " + cfg.Token)
	if err != nil {
		return nil, fmt.Errorf("creating discord session: %w", err)
	}

	session.Identify.Intents = discordgo.IntentsGuilds

	b := &Bot{
		session:  session,
		cfg:      cfg,
		registry: commands.NewRegistry(),
	}

	b.registerCommands()
	session.AddHandler(b.handleInteraction)

	return b, nil
}

func (b *Bot) Start() error {
	if err := b.session.Open(); err != nil {
		return fmt.Errorf("opening discord session: %w", err)
	}

	if err := b.syncCommands(); err != nil {
		return fmt.Errorf("syncing slash commands: %w", err)
	}

	return nil
}

func (b *Bot) Stop() {
	if err := b.session.Close(); err != nil {
		log.Printf("error closing discord session: %v", err)
	}
}

// registerCommands adds all command handlers to the registry.
func (b *Bot) registerCommands() {
	b.registry.Add(commands.Ping())
	b.registry.Add(commands.Status())
}

// syncCommands registers slash commands with Discord.
func (b *Bot) syncCommands() error {
	defs := b.registry.Definitions()

	_, err := b.session.ApplicationCommandBulkOverwrite(
		b.session.State.User.ID,
		b.cfg.GuildID,
		defs,
	)
	if err != nil {
		return fmt.Errorf("bulk overwriting commands: %w", err)
	}

	log.Printf("synced %d slash commands", len(defs))
	return nil
}

// handleInteraction routes incoming interactions to the appropriate command handler.
func (b *Bot) handleInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type != discordgo.InteractionApplicationCommand {
		return
	}

	name := i.ApplicationCommandData().Name
	handler, ok := b.registry.Handler(name)
	if !ok {
		log.Printf("unknown command: %s", name)
		return
	}

	if err := handler(s, i); err != nil {
		log.Printf("command %s error: %v", name, err)
		if respErr := respond(s, i, "Something went wrong."); respErr != nil {
			log.Printf("failed to send error response for command %s: %v", name, respErr)
		}
	}
}

// respond sends an ephemeral text reply to an interaction.
func respond(s *discordgo.Session, i *discordgo.InteractionCreate, msg string) error {
	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: msg,
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}

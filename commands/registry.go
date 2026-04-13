// Package commands defines slash command definitions and their handlers.
package commands

import (
	"github.com/bwmarrin/discordgo"
)

// HandlerFunc is the function signature for command handlers.
type HandlerFunc func(s *discordgo.Session, i *discordgo.InteractionCreate) error

// Command pairs a Discord slash command definition with its handler.
type Command struct {
	Definition *discordgo.ApplicationCommand
	Handler    HandlerFunc
}

// Registry holds all registered commands and routes interactions to handlers.
type Registry struct {
	commands map[string]Command
}

func NewRegistry() *Registry {
	return &Registry{
		commands: make(map[string]Command),
	}
}

// Add registers a command. Panics on nil definition, empty name, or duplicate registration.
func (r *Registry) Add(cmd Command) {
	if cmd.Definition == nil {
		panic("commands: cannot register command with nil Definition")
	}
	if cmd.Definition.Name == "" {
		panic("commands: cannot register command with empty Name")
	}
	if _, exists := r.commands[cmd.Definition.Name]; exists {
		panic("commands: command already registered: " + cmd.Definition.Name)
	}
	r.commands[cmd.Definition.Name] = cmd
}

// Handler returns the handler for a named command.
func (r *Registry) Handler(name string) (HandlerFunc, bool) {
	cmd, ok := r.commands[name]
	if !ok {
		return nil, false
	}
	return cmd.Handler, true
}

// Definitions returns all command definitions for bulk registration with Discord.
func (r *Registry) Definitions() []*discordgo.ApplicationCommand {
	defs := make([]*discordgo.ApplicationCommand, 0, len(r.commands))
	for _, cmd := range r.commands {
		defs = append(defs, cmd.Definition)
	}
	return defs
}

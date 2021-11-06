package tish

import "sort"

var commands = []*Command{}

func RegisterCommand(command *Command) {
	commands = append(commands, command)
	sort.Slice(commands, func(i, j int) bool {
		return commands[i].Name < commands[j].Name
	})
}

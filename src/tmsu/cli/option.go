/*
Copyright 2011-2013 Paul Ruane.

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

package cli

import (
	"fmt"
)

type Option struct {
	LongName    string
	ShortName   string
	Description string
	HasArgument bool
	Argument    string
}

type Options []Option

func (options Options) HasOption(name string) bool {
	for _, option := range options {
		if option.LongName == name || option.ShortName == name {
			return true
		}
	}

	return false
}

func (options Options) Get(name string) *Option {
	for _, option := range options {
		if option.LongName == name || option.ShortName == name {
			return &option
		}
	}

	return nil
}

type OptionParser struct {
	globalOptions Options
	commandByName map[CommandName]Command
}

func NewOptionParser(globalOptions Options, commandByName map[CommandName]Command) *OptionParser {
	parser := OptionParser{globalOptions, commandByName}
	return &parser
}

func (parser *OptionParser) Parse(args []string) (commandName CommandName, options Options, arguments []string, err error) {
	commandName = ""
	options = make(Options, 0)
	arguments = make([]string, 0)

	possibleOptions := make(Options, len(globalOptions))
	copy(possibleOptions, globalOptions)

	parseOptions := true
	for index := 0; index < len(args); index++ {
		arg := args[index]

		if arg == "--" {
			parseOptions = false
		} else {
			if parseOptions && arg[0] == '-' {
				option := lookupOption(possibleOptions, arg)
				if option == nil {
					err = fmt.Errorf("invalid option '%v'.", arg)
					return
				}

				if option.HasArgument {
					option.Argument = args[index+1]
					index++
				}

				options = append(options, *option)
			} else {
				if commandName == "" {
					commandName = CommandName(arg)

					command := parser.commandByName[commandName]
					if command != nil {
						possibleOptions = append(possibleOptions, command.Options()...)
					}
				} else {
					arguments = append(arguments, arg)
				}
			}
		}
	}

	if options.HasOption("--version") {
		commandName = "version"
	}
	if options.HasOption("--help") {
		commandName = "help"
	}

	return
}

// unexported

func lookupOption(options Options, name string) *Option {
	for _, option := range options {
		if option.ShortName == name || option.LongName == name {
			return &Option{option.LongName, option.ShortName, option.Description, option.HasArgument, ""}
		}
	}

	return nil
}
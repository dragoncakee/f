package command

import (
	"errors"
	"strings"
)

var (
	notEnoughArgsErr = errors.New("not enough arguments")
)

func GetCommand(message string) (string, []string, error) {
	index := strings.Index(message, "!")
	if index == -1 {
		return "", nil, errors.New("no scheme found")
	}
	command := message[index+1:]
	if len(command) == 0 || command[0] == ' ' {
		return "", nil, errors.New("invalid command")
	}

	splits := strings.Split(command, " ")

	if len(splits) < 1 {
		return "", nil, notEnoughArgsErr
	}
	source := splits[0]

	var args []string
	if len(splits) > 1 {
		args = splits[1:]
	}

	return source, args, nil
}

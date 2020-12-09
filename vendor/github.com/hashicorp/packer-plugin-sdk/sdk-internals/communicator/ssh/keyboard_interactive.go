package ssh

import (
	"io"
	"log"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/terminal"
)

func KeyboardInteractive(c io.ReadWriter) ssh.KeyboardInteractiveChallenge {
	t := terminal.NewTerminal(c, "")
	return func(user, instruction string, questions []string, echos []bool) ([]string, error) {
		if len(questions) == 0 {
			return []string{}, nil
		}

		log.Printf("[INFO] -- User: %s", user)
		log.Printf("[INFO] -- Instructions: %s", instruction)
		for i, question := range questions {
			log.Printf("[INFO] -- Question %d: %s", i+1, question)
		}
		answers := make([]string, len(questions))
		for i := range questions {
			s, err := t.ReadPassword("")
			if err != nil {
				return nil, err
			}
			answers[i] = string(s)
		}
		return answers, nil
	}
}

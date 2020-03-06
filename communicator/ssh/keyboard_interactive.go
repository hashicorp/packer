package ssh

import (
	"log"
	"os"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/terminal"
)

func KeyboardInteractive() ssh.KeyboardInteractiveChallenge {
	return func(user, instruction string, questions []string, echos []bool) ([]string, error) {
		if len(questions) == 0 {
			return []string{}, nil
		}

		log.Printf("-- User: %s", user)
		log.Printf("-- Instructions: %s", instruction)
		for i, question := range questions {
			log.Printf("-- Question %d: %s", i+1, question)
		}
		answers := make([]string, len(questions))
		for i := range questions {
			s, err := terminal.ReadPassword(int(os.Stdin.Fd()))
			if err != nil {
				return nil, err
			}
			answers[i] = string(s)
		}
		return answers, nil
	}
}

package ssh

import (
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/terminal"
	"log"
	"os"
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
			var fd int
			if terminal.IsTerminal(int(os.Stdin.Fd())) {
				fd = int(os.Stdin.Fd())
			} else {
				tty, err := os.Open("/dev/tty")
				if err != nil {
					return nil, err
				}
				defer tty.Close()
				fd = int(tty.Fd())
			}
			s, err := terminal.ReadPassword(fd)
			if err != nil {
				return nil, err
			}
			answers[i] = string(s)
		}
		return answers, nil
	}
}

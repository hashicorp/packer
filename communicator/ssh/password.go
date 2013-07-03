package ssh

import "log"

// An implementation of ssh.ClientPassword so that you can use a static
// string password for the password to ClientAuthPassword.
type Password string

func (p Password) Password(user string) (string, error) {
	return string(p), nil
}

// An implementation of ssh.ClientKeyboardInteractive that simply sends
// back the password for all questions. The questions are logged.
type PasswordKeyboardInteractive string

func (p PasswordKeyboardInteractive) Challenge(user, instruction string, questions []string, echos []bool) ([]string, error) {
	log.Printf("Keyboard interactive challenge: ")
	log.Printf("-- User: %s", user)
	log.Printf("-- Instructions: %s", instruction)
	for i, question := range questions {
		log.Printf("-- Question %d: %s", i+1, question)
	}

	// Just send the password back for all questions
	answers := make([]string, len(questions))
	for i, _ := range answers {
		answers[i] = string(p)
	}

	return answers, nil
}

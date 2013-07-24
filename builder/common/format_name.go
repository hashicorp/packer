package common

import (
	"bytes"
	"os/user"
	"strconv"
	"text/template"
	"time"
)

type nameData struct {
	CreateTime string
}

func timeFormat(location string, layout string) string {
	in, err := time.LoadLocation(location)
	if err != nil {
		in, _ = time.LoadLocation("UTC")
	}
	return time.Now().In(in).Format(layout)
}

func userName() string {
	user, err := user.Current()
	if err != nil {
		return ""
	}
	return user.Username
}

func FormatName(text string) (string, error) {
	// Create a FuncMap with some useful functions and aliases
	funcMap := template.FuncMap{
		"time": timeFormat,
		"user": userName,
	}

	nameBuf := new(bytes.Buffer)

	// Setup .CreateTime
	tData := nameData{
		strconv.FormatInt(time.Now().UTC().Unix(), 10),
	}

	// Create a template, add the function map, and parse the text.
	t, err := template.New("FormatName").Funcs(funcMap).Parse(text)
	if err != nil {
		return "", err
	}
	t.Execute(nameBuf, tData)
	return nameBuf.String(), nil
}

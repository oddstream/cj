package main

import (
	"net/url"
	"os/exec"
)

func link(str string) error {
	_, err := url.ParseRequestURI(str)
	if err != nil {
		return err
	}
	var cmd *exec.Cmd = exec.Command("xdg-open", str)
	if cmd != nil {
		err := cmd.Start()
		if err != nil {
			return err
		}
	}
	return nil
}

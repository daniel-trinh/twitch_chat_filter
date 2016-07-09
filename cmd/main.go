package main

import (
	"github.com/daniel-trinh/twitch_chat_filter"
	"fmt"
	"os"
)

func main() {
	if err := twitch_chat_filter.TwitchCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}
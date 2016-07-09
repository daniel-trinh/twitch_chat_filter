package twitch_chat_filter

import (
	"time"
	"github.com/spf13/cobra"
)

var (
	IrcServerHost, IrcServerPort, TlsPort, IrcUser, IrcPassword, IrcChannel string
	MessageTTL, FilterDuplicateTTL time.Duration
	FilterDuplicates bool
)

var TwitchCmd = &cobra.Command{
	Use: "twitch_chat_filter",
	Short: "Filters twitch chat stream",
	Long: `Aggregates common messages within a given time window`,
}

func init() {

	TwitchCmd.Run = Start

	TwitchCmd.Flags().StringVarP(&IrcServerHost,
		"irc-server-host",
		"s",
		"irc.chat.twitch.tv",
		`Required. Endpoint of irc server to connect to`,
	)
	TwitchCmd.Flags().StringVarP(&IrcUser,
		"user",
		"u",
		"",
		`Required. IRC username to connect with`,
	)
	TwitchCmd.Flags().StringVarP(&IrcChannel,
		"channel",
		"c",
		"",
		`Required. Irc Channel to join`,
	)
	TwitchCmd.Flags().StringVarP(&IrcServerPort,
		"irc-server-port",
		"p",
		"6667",
		`Port of irc-server-host`,
	)
	TwitchCmd.Flags().StringVarP(&TlsPort,
		"tls-port",
		"t",
		"443",
		`Port for TLS connections`,
	)
	TwitchCmd.Flags().StringVarP(&IrcPassword,
		"password",
		"w",
		"",
		`IRC password to connect with`,
	)
	TwitchCmd.Flags().DurationVarP(&MessageTTL,
		"ttl",
		"l",
		10 * time.Second,
		`Sliding time window in seconds of how long to keep counts of messages in the trending bucket`,
	)
	TwitchCmd.Flags().BoolVarP(&FilterDuplicates,
		"filter-duplicates",
		"f",
		true,
		`If true, will not show messages that have already been chatted recently.`,
	)
	TwitchCmd.Flags().DurationVarP(&FilterDuplicateTTL,
		"filter-duplicate-ttl",
		"d",
		10 * time.Second,
		`If true, will not show messages that have already been chatted recently.`,
	)
}
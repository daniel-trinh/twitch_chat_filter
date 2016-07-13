# IRC Client Filterer / Aggregator (for Twitch)

Twitch chat streams can become unreadable when there are many people typing at the same time.
It is not uncommon to have Twitch streams with over 100,000 concurrent viewers --
when there are this many people in the same chat room, it can become quitee difficult to parse anything meaningful
from the chat stream.

I built this toy CLI UI app to filter out duplicate messages, and to monitor realtime statistics for
twitch chat channels that have non trivial amounts of chat activity. Removing duplicates gets rid of
common chat messages like emoticons and reactions (messages like "LUL", "Kappa", etc), leaving behind

There may be ways of improving the filtering in the future, using some form of streaming TF-IDF for example,
(see [Streaming Trend Detection in Twitter](http://www.cs.uccs.edu/~jkalita/work/reu/REUFinalPapers2010/Benhardus.pdf)),
or doing something simpler with n-gram and stop-word checking, but until I actually have a reason to implement those things,
this app will be as is for now.

You can see the app in action by clicking the image below:
[![Twitch Chat Filterer](http://i.imgur.com/m50Kii1.gif)](https://www.youtube.com/watch?v=i8sRO7_qvOY "Twitch Chat Filterer")

## Functionality
Explanation of widgets in this app:

1) Chat History - history of chat messages.
2) Chat Box - lets you type and send messages.
3) Message Rate Graph - shows the rate of messages in the channel for a given time window (few minutes and past hour)
4) Message Stats - shows min, max, and avg messages for a given time window (few minutes and past hour)
5) Trending Messages - shows duplicate messages, sorted by number of occurrences for a given time window (10s and 60s)

## Usage:

This app requires golang to be installed. I could precompile binaries, but that seems excessive since this is not much
more than a toy app.

* Get a twitch account, login, and go to [http://twitchapps.com/tmi/](http://twitchapps.com/tmi/) to generate
a password token.

* Clone this repo, and cd into the `cmd` folder, and run

```
go run main.go -c=<twitch_channel_name> -s=irc.twitch.tv -w=<password_token> -u=<username>
```

making sure to replace `twitch_channel_name`, `password_token`, and `username`.


## Help Text

This is shown when running the application without passing the required flags.

```
Aggregates common messages within a given time window

Usage:
  twitch_chat_filter [flags]

Flags:
  -c, --channel string                  Required. Irc Channel to join
  -d, --filter-duplicate-ttl duration   If true, will not show messages that have already been chatted recently. (default 10s)
  -f, --filter-duplicates               If true, will not show messages that have already been chatted recently. (default true)
  -h, --help                            help for twitch_chat_filter
  -s, --irc-server-host string          Required. Endpoint of irc server to connect to (default "irc.chat.twitch.tv")
  -p, --irc-server-port string          Port of irc-server-host (default "6667")
  -w, --password string                 IRC password to connect with
  -t, --tls-port string                 Port for TLS connections (default "443")
  -l, --ttl duration                    Sliding time window in seconds of how long to keep counts of messages in the trending bucket (default 10s)
  -u, --user string                     Required. IRC username to connect with
```
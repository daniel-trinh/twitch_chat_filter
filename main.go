package twitch_chat_filter

import (
	irc "github.com/fluffle/goirc/client"
	ui "github.com/gizak/termui"
	//"github.com/sorcix/irc"
	"github.com/spf13/cobra"
	"fmt"
	"os"
	//"time"
	"strings"
	"time"
	"github.com/karlseguin/ccache"
)

type IrcMessage struct {
	Nick string
	Message string
}


var (
	messages chan *IrcMessage
	quit chan bool
	messageTTL time.Duration
	filteredMessagesSet *ccache.Cache
)


func ircConn() *irc.Conn {

	cfg := irc.NewConfig(IrcUser)

	if IrcUser == "" || IrcServerHost == "" || IrcChannel == ""{
		TwitchCmd.Help()
		os.Exit(-1)
	}

	if IrcPassword != "" {
		cfg.Pass = IrcPassword
	}

	cfg.Server = IrcServerHost + ":" + IrcServerPort

	c := irc.Client(cfg)

	return c
}

func shouldFilter(msg *IrcMessage) bool {
	if FilterDuplicates {
		if filteredMessagesSet.Get(msg.Message) != nil {
			return true
		} else {
			filteredMessagesSet.Set(msg.Message, true, FilterDuplicateTTL)
			return false
		}
	} else {
		return false
	}
}

func init() {
	messages = make(chan *IrcMessage, 100)
	filteredMessagesSet = ccache.New(ccache.Configure())
}

func Start(cmd *cobra.Command, args []string) {
	c := ircConn()

	// Add handlers to do things here!
	// e.g. join a channel on connect.
	c.HandleFunc(irc.CONNECTED,
		func(conn *irc.Conn, line *irc.Line) {
			ch := "#" + IrcChannel
			conn.Join(ch)
		},
	)

	// And a signal on disconnect
	c.HandleFunc(irc.DISCONNECTED,
		func(conn *irc.Conn, line *irc.Line) {
			fmt.Println(line)
			quit <- true
		},
	)
	// Put all messages into the messages channel
	c.HandleFunc(irc.PRIVMSG,
		func(conn *irc.Conn, line *irc.Line) {
			message := strings.Join(line.Args[1:], "")
			messages <- &IrcMessage{line.Nick, message}
		},
	)

	// Tell client to connect.
	if err := c.Connect(); err != nil {
		fmt.Printf("Connection error: %s\n", err.Error())
	}

	uiMessageBuffer := make([]string, 32)

	err := ui.Init()
	if err != nil {
		panic(err)
	}
	defer ui.Close()

	p := ui.NewPar("PRESS ctrl+c TO QUIT")
	p.Height = 3
	p.Width = 30
	p.TextFgColor = ui.ColorDefault
	p.BorderLabel = "Twitch Chat Filterer"
	p.BorderFg = ui.ColorCyan

	chatbox := ui.NewPar("Chat here: Use Arrow keys to navigate")
	chatbox.Height = 3
	chatbox.Width = 30
	chatbox.TextFgColor = ui.ColorDefault
	chatbox.BorderLabel = "Chat"
	chatbox.BorderFg = ui.ColorDefault

	minuteMessageCounter, err := NewSlidingWindowCounter(3 * time.Second, 60)
	if err != nil {
		println(err.Error())
		os.Exit(1)
	}
	hourMessageCounter, err := NewSlidingWindowCounter(1 * time.Minute, 60)
	if err != nil {
		println(err.Error())
		os.Exit(1)
	}

	minuteMessageSparkline := ui.NewSparkline()
	minuteMessageSparkline.Data = minuteMessageCounter.Data
	minuteMessageSparkline.Title = "Message Rate (~ 3 Minutes)"
	minuteMessageSparkline.Height=2
	minuteMessageSparkline.LineColor = ui.ColorCyan
	minuteMessageSparkline.TitleColor = ui.ColorGreen

	hourMessageSparkline := ui.NewSparkline()
	hourMessageSparkline.Data = hourMessageCounter.Data
	hourMessageSparkline.Title = "Message Rate (~ 1 Hour)"
	hourMessageSparkline.LineColor = ui.ColorYellow
	hourMessageSparkline.Height = 2
	hourMessageSparkline.TitleColor = ui.ColorGreen

	sparklines := ui.NewSparklines(minuteMessageSparkline, hourMessageSparkline)
	sparklines.Height = 8
	sparklines.BorderFg = ui.ColorDefault

	minuteStats := ui.NewBarChart()
	labels := []string{
		"Min", "Max", "Avg", "Sum",
	}
	minuteStats.BorderLabel = "Stats (~ 3 Minutes)"
	minuteStats.Data = minuteMessageCounter.Stats
	minuteStats.Width = 18
	minuteStats.BarGap = 1
	minuteStats.BarWidth = 4
	minuteStats.Height = 6
	minuteStats.DataLabels = labels
	minuteStats.TextColor = ui.ColorDefault
	minuteStats.BarColor = ui.ColorCyan
	minuteStats.NumColor = ui.ColorRed
	minuteStats.BorderFg = ui.ColorDefault

	hourStats := ui.NewBarChart()
	hourStats.BorderLabel = "Stats (~ 1 Hour)"
	hourStats.Data = minuteMessageCounter.Stats
	hourStats.Width = 10
	hourStats.BarGap = 1
	hourStats.BarWidth = 4
	hourStats.Height = 6
	hourStats.DataLabels = labels
	hourStats.TextColor = ui.ColorDefault
	hourStats.BarColor = ui.ColorYellow
	hourStats.NumColor = ui.ColorMagenta
	hourStats.BorderFg = ui.ColorDefault

	chatHistory := ui.NewList()
	chatHistory.Width = 10
	chatHistory.Height = 34
	chatHistory.BorderLabel = "Chat History"
	chatHistory.BorderFg = ui.ColorDefault
	chatHistory.ItemFgColor = ui.ColorDefault
	chatHistory.Items = uiMessageBuffer

	trendingMessageData := NewSortedMessages(20, 2)
	trendingMessageDataHour := NewSortedMessages(20, 2)
	chatSummary := ui.NewList()
	chatSummary.Width = 10
	chatSummary.Height = 10
	chatSummary.BorderLabel = "Chat Aggregator (10 seconds)"
	chatSummary.BorderFg = ui.ColorDefault
	chatSummary.ItemFgColor = ui.ColorDefault
	chatSummary.Items = trendingMessageData.View
	chatSummary.ItemFgColor = ui.ColorCyan

	chatSummaryHour := ui.NewList()
	chatSummaryHour.Width = 10
	chatSummaryHour.Height = 10
	chatSummaryHour.BorderLabel = "Chat Aggregator (1 minute)"
	chatSummaryHour.BorderFg = ui.ColorDefault
	chatSummaryHour.ItemFgColor = ui.ColorYellow
	chatSummaryHour.Items = trendingMessageDataHour.View

	go func() {
		for {
			select {
			case msg := <-messages:

				// increment counters
				minuteMessageCounter.Increment()
				hourMessageCounter.Increment()

				// Update UI View data buffers
				minuteMessageSparkline.Data = minuteMessageCounter.Data
				minuteStats.Data = minuteMessageCounter.Stats
				hourStats.Data = hourMessageCounter.Stats
				hourMessageSparkline.Data = hourMessageCounter.Data
				sparklines.Lines[0] = minuteMessageSparkline
				sparklines.Lines[1] = hourMessageSparkline

				if !shouldFilter(msg) {
					uiMessageBuffer = append(uiMessageBuffer, msg.Nick+": "+msg.Message)
				}

				chatHistory.Items = uiMessageBuffer

				if len(uiMessageBuffer) > 34 {
					uiMessageBuffer = uiMessageBuffer[1:len(uiMessageBuffer)]
					ui.Render(chatHistory)
				}
				trendingMessageData.Increment(msg.Message, 10 * time.Second)
				trendingMessageDataHour.Increment(msg.Message, 1 * time.Minute)
			case <- quit:
			// TODO: retry connection
				uiMessageBuffer = append(uiMessageBuffer, "<<<WARNING: CHAT DISCONNECTED.>>>")
				chatHistory.Items = uiMessageBuffer
				ui.Render(chatHistory)
			}
		}
	}()

	ui.Body.AddRows(
		ui.NewRow(
			ui.NewCol(4, 0, sparklines,minuteStats, hourStats, chatSummary, chatSummaryHour),
			ui.NewCol(8, 0, p, chatHistory, chatbox),
		),
	)

	ui.Handle("/sys/kbd/q", func(ui.Event) {
		ui.StopLoop()
	})

	ui.Handle("/sys/kbd/C-c", func(ui.Event) {
		ui.StopLoop()
	})

	ui.Handle("/timer/1s", func(e ui.Event) {
		//t := e.Data.(ui.EvtTimer)
		ui.Clear()
		ui.Body.Align()
		ui.Render(ui.Body)
	})

	go func() {
		for {
			_ = <- trendingMessageDataHour.NotifyViewChange
			chatSummaryHour.Items = trendingMessageDataHour.View
			ui.Render(ui.Body)
		}
	}()

	go func() {
		for {
			_ = <- trendingMessageData.NotifyViewChange
			chatSummary.Items = trendingMessageData.View
			ui.Render(ui.Body)
		}
	}()

	ui.Clear()
	ui.Body.Align()
	ui.Render(ui.Body) // feel free to call Render, it's async and non-block

	ui.Loop()
	// event handler...
}
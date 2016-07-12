package twitch_chat_filter

import (
	irc "github.com/fluffle/goirc/client"
	ui "github.com/gizak/termui"
	"github.com/spf13/cobra"
	"fmt"
	"os"
	"time"
	"github.com/karlseguin/ccache"
	"strconv"
	"strings"
)

type IrcMessage struct {
	Nick    string
	Message string
}

func (self *IrcMessage) toString() string {
	return self.Nick + ": " + self.Message
}

type UILayoutState struct {
	Header                *ui.Par
	ChatHistory           *ChatHistory
	ChatBox               *ChatBox
	MinuteStats           *MessageStatsChart
	HourStats             *MessageStatsChart
	MessageRateWidget     *ui.Sparklines
	MinuteMessageRate     *MessageRateSparkline
	HourMessageRate       *MessageRateSparkline
	SecondsChatAggregator *ChatAggregator
	MinuteChatAggregator  *ChatAggregator
	MessageStream         <-chan *IrcMessage
	IrcConn               *irc.Conn
}

func (self *UILayoutState) buildBody() {
	par := ui.NewPar("penis")
	par.Height = 10
	par.Width = 10

	ui.Body.AddRows(
		ui.NewRow(
			//ui.NewCol(12, 0, par),
			ui.NewCol(4, 0, self.MessageRateWidget, self.MinuteStats.Widget, self.HourStats.Widget, self.SecondsChatAggregator.Widget, self.MinuteChatAggregator.Widget),
			ui.NewCol(8, 0, self.Header, self.ChatHistory.Widget, self.ChatBox.Widget),
		),
	)

}

func RegisterIrcAndUiHandlers(c *irc.Conn) (quits chan bool, msgs chan *IrcMessage, typings chan string) {
	// Add handlers to do things here!
	// e.g. join a channel on connect.
	quits = make(chan bool)
	msgs = make(chan *IrcMessage)
	typings = make(chan string)

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
			quits <- true
		},
	)
	// Put all messages into the messages channel
	c.HandleFunc(irc.PRIVMSG,
		func(conn *irc.Conn, line *irc.Line) {
			message := strings.Join(line.Args[1:], "")
			msgs <- &IrcMessage{line.Nick, message}
		},
	)

	ui.Handle("/sys/kbd/C-c", func(ui.Event) {
		ui.StopLoop()
		ui.Close()
		os.Exit(0)
	})

	ui.Handle("/sys/kbd", func(kbdEvent ui.Event) {
		// Could also be handled with state machine.
		inputStr := kbdEvent.Data.(ui.EvtKbd).KeyStr

		typings <- inputStr
		//if chatbox.Text == chatboxDefaultText {
		//	chatbox.Text = inputStr
		//} else {
		//	chatbox.Text = chatbox.Text + inputStr
		//}
		//ui.Render(chatbox)
	})

	ui.Handle("/timer/1s", func(e ui.Event) {
		ui.Clear()
		ui.Body.Align()
		ui.Render(ui.Body)
	})

	return quits, msgs, typings
}

func (self *UILayoutState) InitBodyAndLoop() {
	self.buildBody()

	ui.Clear()
	ui.Body.Align()
	ui.Render(ui.Body)

	ui.Loop()
}

func NewUILayoutState() *UILayoutState {
	err := ui.Init()
	if err != nil {
		panic(err)
	}

	c := ircConn()
	quits, messages, typings := RegisterIrcAndUiHandlers(c)

	// Tell client to connect.
	if err = c.Connect(); err != nil {
		fmt.Printf("Connection error: %s\n", err.Error())
	}

	minuteCounter, err := NewSlidingWindowCounter(3 * time.Second, 60)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	hourCounter, err := NewSlidingWindowCounter(1 * time.Minute, 60)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	minuteMessageRate := NewMessageRateSparkline(ui.ColorCyan, minuteCounter)
	hourMessageRate := NewMessageRateSparkline(ui.ColorYellow, hourCounter)
	messageRateWidget := ui.NewSparklines(*minuteMessageRate.Widget, *hourMessageRate.Widget)
	messageRateWidget.Height = 8
	messageRateWidget.BorderFg = ui.ColorDefault

	header := ui.NewPar("PRESS ctrl+c TO QUIT")
	header.Height = 3
	header.Width = 30
	header.TextFgColor = ui.ColorDefault
	header.BorderLabel = "Twitch Chat Filterer"
	header.BorderFg = ui.ColorCyan

	uiState := UILayoutState{
		Header: header,
		ChatHistory: NewChatHistory(),
		ChatBox: NewChatBox(typings),
		MinuteStats: NewMessageStatsChart(ui.ColorCyan, ui.ColorRed, minuteCounter),
		HourStats: NewMessageStatsChart(ui.ColorYellow, ui.ColorMagenta, hourCounter),
		MessageRateWidget: messageRateWidget,
		MinuteMessageRate: minuteMessageRate,
		HourMessageRate: hourMessageRate,
		SecondsChatAggregator: NewChatAggregator(ui.ColorCyan, 10 * time.Second),
		MinuteChatAggregator: NewChatAggregator(ui.ColorYellow, 1 * time.Minute),
		MessageStream: messages,
		IrcConn: c,
	}


	go func() {
		for {
			msg := <-uiState.MessageStream
			// increment counters, update buffers, rerender views
			uiState.MinuteMessageRate.IncrementAndRender(uiState.MessageRateWidget)
			uiState.HourMessageRate.IncrementAndRender(uiState.MessageRateWidget)
			uiState.MessageRateWidget.Lines[0] = *uiState.MinuteMessageRate.Widget
			uiState.MessageRateWidget.Lines[1] = *uiState.HourMessageRate.Widget

			uiState.MinuteStats.Render()
			uiState.HourStats.Render()

			uiState.ChatHistory.UpdateAndRender(msg)

			uiState.SecondsChatAggregator.SortedMessages.Increment(msg.Message, uiState.SecondsChatAggregator.TTL)
			uiState.MinuteChatAggregator.SortedMessages.Increment(msg.Message, uiState.MinuteChatAggregator.TTL)
		}
	}()

	go func() {
		for {
			_ = <- quits
			uiState.ChatHistory.updateAndRender("<<<WARNING: CHAT DISCONNECTED.>>>")
		}
	}()

	go func() {
		for {
			_ = <- time.Tick(1 * time.Second)
			ui.Clear()
			ui.Render(ui.Body)
		}
	}()


	return &uiState
}

type ChatBox struct {
	Widget         *ui.Par
	DefaultMessage string
	Data           string
	Input          <-chan string
}

func NewChatBox(input <-chan string) *ChatBox {
	defaultText := "Start typing, press Enter to send message"

	widget := ui.NewPar(defaultText)
	widget.Height = 3
	widget.Width = 30
	widget.TextFgColor = ui.ColorDefault
	widget.BorderLabel = "Chat"
	widget.BorderFg = ui.ColorDefault
	//
	//go func() {
	//	for {
	//		typing := <-input
	//	}
	//}()

	return &ChatBox{
		Widget: widget,
		DefaultMessage: defaultText,
		Data: defaultText,
		Input: input,
	}

}

type ChatHistory struct {
	Widget              *ui.List
	ChatHistoryData     []string
	Size                int
	FilteredMessagesSet *ccache.Cache
}

func NewChatHistory() *ChatHistory {
	uiMessageBuffer := make([]string, 32)

	widget := ui.NewList()
	widget.Width = 10
	widget.Height = 34
	widget.BorderLabel = "Chat History"
	widget.BorderFg = ui.ColorDefault
	widget.ItemFgColor = ui.ColorDefault
	widget.Items = uiMessageBuffer

	return &ChatHistory{
		Widget : widget,
		ChatHistoryData: uiMessageBuffer,
		FilteredMessagesSet: ccache.New(ccache.Configure()),
	}
}

func (self *ChatHistory) UpdateAndRender(msg *IrcMessage) {
	if !self.shouldFilter(msg) {
		self.updateAndRender(msg.toString())
	}
}

func (self *ChatHistory) updateAndRender(msg string) {
	self.ChatHistoryData = append(self.ChatHistoryData, msg)
	if len(self.ChatHistoryData) > self.Size {
		self.ChatHistoryData = self.ChatHistoryData[1:]
	}

	self.Widget.Items = self.ChatHistoryData

	ui.Render(self.Widget)
}

func (self *ChatHistory) shouldFilter(msg *IrcMessage) bool {
	if FilterDuplicates {
		if self.FilteredMessagesSet.Get(msg.Message) != nil {
			return true
		} else {
			self.FilteredMessagesSet.Set(msg.Message, true, FilterDuplicateTTL)
			return false
		}
	} else {
		return false
	}
}

type ChatAggregator struct {
	TTL            time.Duration
	Widget         *ui.List
	SortedMessages *SortedMessages
}

func NewChatAggregator(textColor ui.Attribute, ttl time.Duration) *ChatAggregator {
	widget := ui.NewList()
	trendingMessages := NewSortedMessages(20, 2)
	widget.Width = 10
	widget.Height = 10

	printedTime := ttl / time.Second
	widget.BorderLabel = "Trending (" + strconv.FormatInt(int64(printedTime), 10) + " secs)"
	widget.BorderFg = ui.ColorDefault
	widget.ItemFgColor = ui.ColorDefault
	widget.Items = trendingMessages.View
	widget.ItemFgColor = textColor

	go func() {
		for {
			_ = <-trendingMessages.NotifyViewChange
			widget.Items = trendingMessages.View
			ui.Render(widget)
		}
	}()

	return &ChatAggregator{
		TTL: ttl,
		Widget: widget,
		SortedMessages: trendingMessages,
	}

}

type MessageStatsChart struct {
	Widget  *ui.BarChart
	Counter *SlidingWindowCounter
}

func NewMessageStatsChart(barColor, numColor ui.Attribute, counter *SlidingWindowCounter) *MessageStatsChart {
	statsWidget := ui.NewBarChart()

	labels := []string{
		"Min", "Max", "Avg",
	}

	statsWidget.BorderLabel = "Stats (~ 3 min)"
	statsWidget.Data = counter.Stats
	statsWidget.Width = 10
	statsWidget.BarGap = 1
	statsWidget.BarWidth = 4
	statsWidget.Height = 6
	statsWidget.DataLabels = labels
	statsWidget.TextColor = ui.ColorDefault
	statsWidget.BarColor = barColor
	statsWidget.NumColor = numColor
	statsWidget.BorderFg = ui.ColorDefault

	return &MessageStatsChart{
		Widget: statsWidget,
		Counter: counter,
	}
}

func (self *MessageStatsChart) Render() {
	self.Widget.Data = self.Counter.Data
	ui.Render(self.Widget)
}

type MessageRateSparkline struct {
	Widget  *ui.Sparkline
	Counter *SlidingWindowCounter
}

func NewMessageRateSparkline(color ui.Attribute, counter *SlidingWindowCounter) *MessageRateSparkline {
	widget := ui.NewSparkline()
	widget.Data = counter.Data
	formattedTime := strconv.Itoa(counter.NumWindows * int(counter.Window) / int(time.Minute))
	widget.Title = "Message Rate (~ " + formattedTime + " Minutes)"
	widget.Height = 2
	widget.LineColor = ui.ColorCyan
	widget.TitleColor = ui.ColorGreen

	return &MessageRateSparkline{
		Widget: &widget,
		Counter: counter,
	}
}

func (self *MessageRateSparkline) IncrementAndRender(sparklines *ui.Sparklines) {
	self.Counter.Increment()
	// Update UI View data buffers
	self.Widget.Data = self.Counter.Data
	ui.Render(sparklines)
}

func ircConn() *irc.Conn {

	cfg := irc.NewConfig(IrcUser)

	if IrcUser == "" || IrcServerHost == "" || IrcChannel == "" {
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

func Start(cmd *cobra.Command, args []string) {

	uiState := NewUILayoutState()
	uiState.InitBodyAndLoop()

	ui.Clear()
	ui.Body.Align()
	ui.Render(ui.Body) // feel free to call Render, it's async and non-block

	ui.Loop()
}
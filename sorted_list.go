package twitch_chat_filter

import (
	"strconv"
	"time"
)

type MessageCount struct {
	Message string
	Count   int
}

func (self *MessageCount) toString(threshold int) string {
	if self.Count >= threshold {
		return  "(" + strconv.Itoa(self.Count) + ") " + self.Message
	} else {
		return ""
	}
}

// updateAction Types
const (
	Increment = iota
	Decrement = iota
)

type updateAction struct {
	Message string
	Type    int
}

type SortedMessages struct {
	Data               []MessageCount     // no size limit
	View               []string           // size limit applies
	ViewSize           int
	ViewCountThreshold int
	indices            map[string]int
	NotifyViewChange   chan bool          // an element is put in here every time the view changes
	mailbox            chan *updateAction // used to serialize updates one at a time. named after actor model
}

func NewSortedMessages(size, viewThreshold int) *SortedMessages {
	sm := &SortedMessages{
		Data: make([]MessageCount, 0),
		View: make([]string, 0),
		indices: make(map[string]int),
		ViewSize: size,
		mailbox: make(chan *updateAction, 200),
		NotifyViewChange: make(chan bool, 200),
		ViewCountThreshold: viewThreshold,
	}
	sm.initMessageWorker()
	return sm
}

func (self *SortedMessages) initMessageWorker() {
	go func() {
		for {
			mail := <-self.mailbox
			if mail.Type == Increment {
				self.increment(mail.Message)
			} else if mail.Type == Decrement {
				self.decrement(mail.Message)
			}
		}
	}()
}

func (self *SortedMessages) Increment(message string, ttl time.Duration) {
	self.mailbox <- &updateAction{message, Increment}
	if ttl > 0 {
		self.Decrement(message, ttl)
	}
}

func (self *SortedMessages) Decrement(message string, delay time.Duration) {
	action := updateAction{message, Decrement}
	if delay > 0 {
		go func() {
			time.Sleep(delay)
			self.mailbox <- &action
		}()
	} else {
		self.mailbox <- &action
	}
}

// Not threadsafe, use Increment.
func (self *SortedMessages) increment(message string) {
	if i, ok := self.indices[message]; ok {
		self.Data[i].Count += 1
		if i < self.ViewSize {
			self.View[i] = self.Data[i].toString(self.ViewCountThreshold)
		}
		// Sorted by descending order
		for j := i - 1; j >= 0; i, j = i - 1, j - 1 {
			if self.Data[j].Count < self.Data[i].Count {
				self.swap(i, j)
			} else {
				break
			}
		}
	} else {
		// New message, put it at the back of the slice since it only has one count
		newMessage := MessageCount{message, 1}
		self.Data = append(self.Data, newMessage)
		newIndex := len(self.Data) - 1
		self.indices[message] = newIndex

		// Update representation for UI if element is within size bounds
		if newIndex < self.ViewSize {
			self.View = append(self.View, self.Data[newIndex].toString(self.ViewCountThreshold))
		}
	}
	self.NotifyViewChange <- true
}

// Not threadsafe, use Decrement.
func (self *SortedMessages) decrement(message string) {
	if i, ok := self.indices[message]; ok {
		// Decremented to 0 or less, get rid of index, remove from data and view
		if count := self.Data[i].Count; count <= 1 {
			delete(self.indices, message)

			startIndexSecondHalf := Min(i + 1, len(self.Data))
			rest := self.Data[startIndexSecondHalf:len(self.Data)]
			self.Data = append(self.Data[:i], rest...)
			// Update indices since we just shifted everything down one.
			for j := i; j < len(self.Data); j++ {
				e := self.Data[j]
				self.indices[e.Message] = j
			}

			// Only try to remove item from view if it exists in the view
			if i < len(self.View) {
				a := self.View[:Min(i, len(self.View)-1)]
				b := self.View[Min(i + 1, len(self.View)):len(self.View)]
				self.View = append(a, b...)
			}

			// Since we just got rid of a reference in the View, we may need to pull in another one from the bottom of the Data slice
			if len(self.View) < self.ViewSize && len(self.Data) >= self.ViewSize {
				self.View = append(self.View, self.Data[len(self.View)].toString(self.ViewCountThreshold))
			}
		} else {
			self.Data[i].Count -= 1
			if i < self.ViewSize {
				self.View[i] = self.Data[i].toString(self.ViewCountThreshold)
			}
			// [5,4,4,3,2,1]
			// [5,3,4,3,2,1]
			// [5,4,3,3,2,1]
			// Just decremented a value, might need to move references down (-> that way).
			for j := i + 1; j < len(self.Data); i, j = i + 1, j + 1 {
				if self.Data[i].Count < self.Data[j].Count {
					self.swap(i, j)
				} else {
					break
				}
			}
		}
		self.NotifyViewChange <- true
	}
}

func (self *SortedMessages) swap(i int, j int) {
	self.Data[i], self.Data[j] = self.Data[j], self.Data[i]

	// Update view if within view size
	if i < self.ViewSize {
		self.View[i] = self.Data[i].toString(self.ViewCountThreshold)
	}
	if j < self.ViewSize {
		self.View[j] = self.Data[j].toString(self.ViewCountThreshold)
	}

	self.indices[self.Data[j].Message] = j
	self.indices[self.Data[i].Message] = i
}

func Min(x, y int) int {
	if x < y {
		return x
	}
	return y
}

func Max(x, y int) int {
	if x > y {
		return x
	}
	return y
}
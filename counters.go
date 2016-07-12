package twitch_chat_filter

import (
	"time"
	"errors"
	"math"
)


type SlidingWindowCounter struct {
	Data []int
	Min int
	Max int
	Avg int
	Sum int
	Stats []int // contains Min, Max, Avg, Sum
	Window time.Duration
	NumWindows int
}

func NewSlidingWindowCounter(timeWindow time.Duration, numWindows int) (*SlidingWindowCounter, error) {
	if numWindows < 2 {
		return nil, errors.New("numWindows must be greater than 2")
	}
	counter := &SlidingWindowCounter{
		Data: make([]int, 2),
		Min: math.MaxInt32,
		Max: math.MinInt32,
		Avg: 0,
		Sum: 0,
		Stats: make([]int, 3),
		Window: timeWindow,
		NumWindows: numWindows,
	}
	// Create a new Sparkline window after every timeWindow
	go func() {
		for {
			<-time.Tick(timeWindow)
			counter.Min, counter.Max, counter.Avg, counter.Sum = counter.calculateStats()
			counter.Data = append(counter.Data, 0)
			counter.Stats[0] = counter.Min
			counter.Stats[1] = counter.Max
			counter.Stats[2] = counter.Avg
			if len(counter.Data) > numWindows {
				counter.Data = counter.Data[1:len(counter.Data)]
			}
		}
	}()

	return counter, nil
}

func (self *SlidingWindowCounter) calculateStats() (min, max, avg, sum int) {
	min, max, avg, sum = math.MaxInt32, math.MinInt32, 0, 0
	for _, x := range self.Data {
		if min > 0 {
			min = Min(x, min)
		} else {
			min = x
		}
		max = Max(x, max)
		sum += x
	}
	avg = sum / len(self.Data)

	return min, max, avg, sum
}

func (self *SlidingWindowCounter) Increment() {
	self.Data[len(self.Data)-1] += 1
}
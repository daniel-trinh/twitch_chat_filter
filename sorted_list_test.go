package twitch_chat_filter

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"time"
)
func TestSortedList(t *testing.T) {
	messages := NewSortedMessages(5, 0)
	messages.Increment("LUL", 0)
	time.Sleep(2 * time.Millisecond)

	assert.Equal(t, []MessageCount{MessageCount{"LUL", 1}},messages.Data)
	assert.Equal(t, []string{"(1) LUL"}, messages.View)
	messages.Increment("LUL", 0)
	time.Sleep(2 * time.Millisecond)

	assert.Equal(t, []MessageCount{MessageCount{"LUL", 2}}, messages.Data)
	assert.Equal(t, []string{"(2) LUL"}, messages.View)
	messages.Decrement("LUL", 0)
	time.Sleep(2 * time.Millisecond)

	assert.Equal(t, []MessageCount{MessageCount{"LUL", 1}}, messages.Data)
	assert.Equal(t, []string{"(1) LUL"}, messages.View)
	messages.Decrement("LUL", 0)
	time.Sleep(2 * time.Millisecond)

	assert.Equal(t, []MessageCount{}, messages.Data)
	assert.Equal(t, []string{}, messages.View)
}

func TestSortedListMultipleMessages(t *testing.T) {
	messages := NewSortedMessages(5, 0)
	messages.Increment("LUL", 0)
	messages.Increment("LOL", 0)
	messages.Increment("DoomGuy", 0)
	time.Sleep(2 * time.Millisecond)

	assert.Equal(t, []MessageCount{
		MessageCount{"LUL", 1},
		MessageCount{"LOL", 1},
		MessageCount{"DoomGuy", 1},
	}, messages.Data)

	assert.Equal(t, []string{
		"(1) LUL",
		"(1) LOL",
		"(1) DoomGuy",
	}, messages.View)


	messages.Increment("DoomGuy", 0)
	messages.Increment("LOL", 0)
	messages.Increment("LOL", 0)
	time.Sleep(2 * time.Millisecond)

	assert.Equal(t, []MessageCount{
		MessageCount{"LOL", 3},
		MessageCount{"DoomGuy", 2},
		MessageCount{"LUL", 1},
	}, messages.Data)
	assert.Equal(t, []string{
		"(3) LOL",
		"(2) DoomGuy",
		"(1) LUL",
	}, messages.View)

	messages.Decrement("LOL", 0)
	messages.Decrement("LOL", 0)
	time.Sleep(2 * time.Millisecond)

	assert.Equal(t, []MessageCount{
		MessageCount{"DoomGuy", 2},
		MessageCount{"LOL", 1},
		MessageCount{"LUL", 1},
	}, messages.Data)

	assert.Equal(t, []string{
		"(2) DoomGuy",
		"(1) LOL",
		"(1) LUL",
	}, messages.View)

	messages.Decrement("LOL", 0)
	messages.Decrement("DoomGuy", 0)
	messages.Decrement("DoomGuy", 0)
	messages.Decrement("LUL", 0)
	time.Sleep(2 * time.Millisecond)

	assert.Equal(t, []MessageCount{}, messages.Data)
	assert.Equal(t, []string{}, messages.View)

	messages.Decrement("LUL", 0)
	time.Sleep(2 * time.Millisecond)

	assert.Equal(t, []MessageCount{}, messages.Data)
	assert.Equal(t, []string{}, messages.View)

}

func TestSortedListOverflow(t *testing.T) {
	messages := NewSortedMessages(5, 0)
	messages.Increment("LUL", 0)
	messages.Increment("LOL", 0)
	messages.Increment("DoomGuy", 0)
	messages.Increment("Kappa", 0)
	messages.Increment("Kappa //", 0)
	messages.Increment("BibleThump", 0)
	messages.Increment("Sadness", 0)
	time.Sleep(2 * time.Millisecond)

	assert.Equal(t, []MessageCount{
		MessageCount{"LUL", 1},
		MessageCount{"LOL", 1},
		MessageCount{"DoomGuy", 1},
		MessageCount{"Kappa", 1},
		MessageCount{"Kappa //", 1},
		MessageCount{"BibleThump", 1},
		MessageCount{"Sadness", 1},
	}, messages.Data)


	assert.Equal(t, []string{
		"(1) LUL",
		"(1) LOL",
		"(1) DoomGuy",
		"(1) Kappa",
		"(1) Kappa //",
	}, messages.View)

	messages.Increment("BibleThump", 0)
	messages.Increment("BibleThump", 0)
	time.Sleep(2 * time.Millisecond)

	assert.Equal(t, []string{
		"(3) BibleThump",
		"(1) LUL",
		"(1) LOL",
		"(1) DoomGuy",
		"(1) Kappa",
	}, messages.View)

	messages.Increment("Kappa //", 0)
	time.Sleep(2 * time.Millisecond)

	assert.Equal(t, []string{
		"(3) BibleThump",
		"(2) Kappa //",
		"(1) LUL",
		"(1) LOL",
		"(1) DoomGuy",
	}, messages.View)

	messages.Decrement("LUL", 0)
	time.Sleep(2 * time.Millisecond)

	assert.Equal(t, []string{
		"(3) BibleThump",
		"(2) Kappa //",
		"(1) LOL",
		"(1) DoomGuy",
		"(1) Kappa",
	}, messages.View)
	messages.Decrement("LOL", 0)
	time.Sleep(2 * time.Millisecond)

	assert.Equal(t, []string{
		"(3) BibleThump",
		"(2) Kappa //",
		"(1) DoomGuy",
		"(1) Kappa",
		"(1) Sadness",
	}, messages.View)

	assert.Equal(t, []MessageCount{
		MessageCount{"BibleThump", 3},
		MessageCount{"Kappa //", 2},
		MessageCount{"DoomGuy", 1},
		MessageCount{"Kappa", 1},
		MessageCount{"Sadness", 1},
	}, messages.Data)

	messages.Increment("LOL", 0)
	messages.Increment("LUL", 0)
	messages.Decrement("Sadness", 0)
	time.Sleep(2 * time.Millisecond)

}
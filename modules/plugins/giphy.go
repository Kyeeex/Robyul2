package plugins

import (
	"fmt"
	"math/rand"
	"net/url"

	"github.com/Seklfreak/Robyul2/helpers"
	"github.com/bwmarrin/discordgo"
)

type Giphy struct{}

func (g *Giphy) Commands() []string {
	return []string{
		"giphy",
		"gif",
	}
}

func (g *Giphy) Init(session *discordgo.Session) {

}

func (g *Giphy) Action(command string, content string, msg *discordgo.Message, session *discordgo.Session) {
	const ENDPOINT = "http://api.giphy.com/v1/gifs/search"
	const API_KEY = "dc6zaTOxFJmzC"
	const RATING = "pg-13"
	const LIMIT = 5

	session.ChannelTyping(msg.ChannelID)

	// Send request
	json := helpers.GetJSON(
		fmt.Sprintf(
			"%s?q=%s&api_key=%s&rating=%s&limit=%d",
			ENDPOINT,
			url.QueryEscape(content),
			API_KEY,
			RATING,
			LIMIT,
		),
	)

	// Get gifs
	gifs, err := json.Path("data").Children()
	if err != nil {
		helpers.SendMessage(msg.ChannelID, "Error parsing Giphy's response <:blobfrowningbig:317028438693117962>")
		return
	}

	// Chose a random one
	var m string
	if len(gifs) > 0 {
		m = gifs[rand.Intn(len(gifs))].Path("bitly_url").Data().(string)
	} else {
		m = "No gifs found <:blobfrowningbig:317028438693117962>"
	}

	// Send the result
	helpers.SendMessage(msg.ChannelID, m)
}

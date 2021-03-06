package plugins

import (
	"fmt"
	"math/rand"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/Seklfreak/Robyul2/helpers"
	"github.com/bwmarrin/discordgo"
)

type XKCD struct{}

func (x *XKCD) Commands() []string {
	return []string{
		"xkcd",
	}
}

func (x *XKCD) Init(session *discordgo.Session) {

}

func (x *XKCD) Action(command string, content string, msg *discordgo.Message, session *discordgo.Session) {
	session.ChannelTyping(msg.ChannelID)

	var link string

	if regexp.MustCompile("^\\d+$").MatchString(content) {
		link = "https://xkcd.com/" + content + "/info.0.json"
	} else if strings.Contains(content, "rand") {
		// Get latest number
		doc, err := goquery.NewDocument("https://xkcd.com")
		helpers.Relax(err)

		var num string
		for _, attr := range doc.Find("ul.comicNav").Children().Get(1).FirstChild.Attr {
			if attr.Key == "href" {
				num = attr.Val
				break
			}
		}

		num = strings.Replace(num, "/", "", -1)

		max, err := strconv.ParseInt(num, 10, 32)
		if err != nil {
			helpers.SendMessage(msg.ChannelID, "Error getting latest comic. Try again later <:blobfrowningbig:317028438693117962>")
			return
		}

		link = "https://xkcd.com/" + strconv.Itoa(rand.Intn(int(max-1))+1) + "/info.0.json"
	} else {
		link = "https://xkcd.com/info.0.json"
	}

	json := helpers.GetJSON(link)
	helpers.SendMessage(
		msg.ChannelID,
		fmt.Sprintf(
			"#%d from %s/%s/%s\n%s\n%s",
			int(json.Path("num").Data().(float64)),
			json.Path("day").Data().(string),
			json.Path("month").Data().(string),
			json.Path("year").Data().(string),
			json.Path("title").Data().(string),
			json.Path("img").Data().(string),
		),
	)
	helpers.SendMessage(msg.ChannelID, json.Path("alt").Data().(string))
}

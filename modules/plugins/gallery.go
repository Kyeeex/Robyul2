package plugins

import (
	"fmt"
	"regexp"
	"strings"

	"time"

	"github.com/Seklfreak/Robyul2/cache"
	"github.com/Seklfreak/Robyul2/helpers"
	"github.com/Seklfreak/Robyul2/metrics"
	"github.com/Sirupsen/logrus"
	"github.com/bwmarrin/discordgo"
	"github.com/getsentry/raven-go"
	rethink "github.com/gorethink/gorethink"
	"github.com/vmihailenco/msgpack"
)

type Gallery struct{}

type DB_Gallery_Entry struct {
	ID                        string `gorethink:"id,omitempty"`
	SourceChannelID           string `gorethink:"source_channel_id"`
	TargetChannelID           string `gorethink:"target_channel_id"`
	TargetChannelWebhookID    string `gorethink:"target_channel_webhook_id"`
	TargetChannelWebhookToken string `gorethink:"target_channel_webhook_token"`
	GuildID                   string `gorethink:"guild_id"`
	AddedByUserID             string `gorethink:"addedby_user_id"`
}

func (g *Gallery) Commands() []string {
	return []string{
		"gallery",
	}
}

const (
	galleryUrlRegexText = `(<?https?:\/\/[^\s]+>?)`
)

var (
	galleryUrlRegex *regexp.Regexp
	galleries       []DB_Gallery_Entry
)

func (g *Gallery) Init(session *discordgo.Session) {
	galleryUrlRegex = regexp.MustCompile(galleryUrlRegexText)
	galleries = g.GetGalleries()
}

func (g *Gallery) Uninit(session *discordgo.Session) {

}

func (g *Gallery) Action(command string, content string, msg *discordgo.Message, session *discordgo.Session) {
	args := strings.Fields(content)
	if len(args) >= 1 {
		switch args[0] {
		case "add": // [p]gallery add <source channel> <target channel> <webhook id> <webhook token>
			// @TODO: more secure way to exchange token: create own webhook if no arguments passed
			helpers.RequireAdmin(msg, func() {
				session.ChannelMessageDelete(msg.ChannelID, msg.ID) // Delete command message to prevent people seeing the token
				progressMessages, err := helpers.SendMessage(msg.ChannelID, helpers.GetText("plugins.gallery.add-progress"))
				helpers.Relax(err)
				if len(progressMessages) <= 0 {
					helpers.SendMessage(msg.ChannelID, helpers.GetText("bot.errors.generic-nomessage"))
					return
				}
				progressMessage := progressMessages[0]
				if len(args) < 5 {
					_, err := helpers.EditMessage(msg.ChannelID, progressMessage.ID, helpers.GetText("bot.arguments.too-few"))
					helpers.Relax(err)
					return
				}
				channel, err := helpers.GetChannel(msg.ChannelID)
				helpers.Relax(err)
				guild, err := helpers.GetGuild(channel.GuildID)
				helpers.Relax(err)
				sourceChannel, err := helpers.GetChannelFromMention(msg, args[1])
				if err != nil || sourceChannel.ID == "" || sourceChannel.GuildID != channel.GuildID {
					_, err := helpers.EditMessage(msg.ChannelID, progressMessage.ID, helpers.GetText("bot.arguments.invalid"))
					helpers.Relax(err)
					return
				}
				targetChannel, err := helpers.GetChannelFromMention(msg, args[2])
				if err != nil || targetChannel.ID == "" || targetChannel.GuildID != channel.GuildID {
					_, err := helpers.EditMessage(msg.ChannelID, progressMessage.ID, helpers.GetText("bot.arguments.invalid"))
					helpers.Relax(err)
					return
				}

				targetChannelWebhookId := args[3]
				targetChannelWebhookToken := args[4]

				webhook, err := session.WebhookWithToken(targetChannelWebhookId, targetChannelWebhookToken)
				if err != nil || webhook.GuildID != targetChannel.GuildID || webhook.ChannelID != targetChannel.ID {
					_, err := helpers.EditMessage(msg.ChannelID, progressMessage.ID, helpers.GetText("bot.arguments.invalid"))
					helpers.Relax(err)
					return
				}

				newGalleryEntry := g.getEntryByOrCreateEmpty("id", "")
				newGalleryEntry.SourceChannelID = sourceChannel.ID
				newGalleryEntry.TargetChannelID = targetChannel.ID
				newGalleryEntry.TargetChannelWebhookID = targetChannelWebhookId
				newGalleryEntry.TargetChannelWebhookToken = targetChannelWebhookToken
				newGalleryEntry.AddedByUserID = msg.Author.ID
				newGalleryEntry.GuildID = channel.GuildID
				g.setEntry(newGalleryEntry)

				cache.GetLogger().WithField("module", "galleries").Info(fmt.Sprintf("Added Gallery on Server %s (%s) posting from #%s (%s) to #%s (%s)",
					guild.Name, guild.ID, sourceChannel.Name, sourceChannel.ID, targetChannel.Name, targetChannel.ID))
				_, err = helpers.EditMessage(msg.ChannelID, progressMessage.ID, helpers.GetText("plugins.gallery.add-success"))
				helpers.Relax(err)

				galleries = g.GetGalleries()
				return
			})
		case "list": // [p]gallery list
			session.ChannelTyping(msg.ChannelID)
			channel, err := helpers.GetChannel(msg.ChannelID)
			helpers.Relax(err)
			var entryBucket []DB_Gallery_Entry
			listCursor, err := rethink.Table("galleries").Filter(
				rethink.Row.Field("guild_id").Eq(channel.GuildID),
			).Run(helpers.GetDB())
			helpers.Relax(err)
			defer listCursor.Close()
			err = listCursor.All(&entryBucket)

			if err == rethink.ErrEmptyResult || len(entryBucket) <= 0 {
				helpers.SendMessage(msg.ChannelID, helpers.GetText("plugins.gallery.list-empty"))
				return
			}
			helpers.Relax(err)

			resultMessage := ":frame_photo: Galleries on this server:\n"
			for _, entry := range entryBucket {
				resultMessage += fmt.Sprintf("`%s`: posting from <#%s> to <#%s> (Webhook ID: `%s`)\n",
					entry.ID, entry.SourceChannelID, entry.TargetChannelID, entry.TargetChannelWebhookID)
			}
			resultMessage += fmt.Sprintf("Found **%d** Galleries in total.", len(entryBucket))

			for _, resultPage := range helpers.Pagify(resultMessage, "\n") {
				_, err = helpers.SendMessage(msg.ChannelID, resultPage)
				helpers.Relax(err)
			}
			return
		case "delete", "del", "remove": // [p]gallery delete <gallery id>
			helpers.RequireAdmin(msg, func() {
				session.ChannelTyping(msg.ChannelID)
				if len(args) < 2 {
					_, err := helpers.SendMessage(msg.ChannelID, helpers.GetText("bot.arguments.too-few"))
					helpers.Relax(err)
					return
				}
				entryId := args[1]
				entryBucket := g.getEntryBy("id", entryId)
				if entryBucket.ID == "" {
					helpers.SendMessage(msg.ChannelID, helpers.GetText("plugins.gallery.delete-not-found"))
					return
				}
				galleryGuild, _ := helpers.GetGuild(entryBucket.GuildID)
				sourceChannel, _ := helpers.GetChannel(entryBucket.SourceChannelID)
				if sourceChannel == nil {
					sourceChannel = new(discordgo.Channel)
					sourceChannel.Name = "N/A"
					sourceChannel.ID = "N/A"
				}
				targetChannel, _ := helpers.GetChannel(entryBucket.TargetChannelID)
				if targetChannel == nil {
					targetChannel = new(discordgo.Channel)
					targetChannel.Name = "N/A"
					targetChannel.ID = "N/A"
				}
				g.deleteEntryById(entryBucket.ID)

				cache.GetLogger().WithField("module", "galleries").Info(fmt.Sprintf("Deleted Gallery on Server %s (%s) posting from #%s (%s) to #%s (%s)",
					galleryGuild.Name, galleryGuild.ID, sourceChannel.Name, sourceChannel.ID, targetChannel.Name, targetChannel.ID))
				_, err := helpers.SendMessage(msg.ChannelID, helpers.GetText("plugins.gallery.delete-success"))
				helpers.Relax(err)

				galleries = g.GetGalleries()
				return
			})
		case "refresh": // [p]gallery refresh
			helpers.RequireBotAdmin(msg, func() {
				session.ChannelTyping(msg.ChannelID)
				galleries = g.GetGalleries()
				_, err := helpers.SendMessage(msg.ChannelID, helpers.GetText("plugins.gallery.refreshed-config"))
				helpers.Relax(err)
			})
		}
	}
}

func (g *Gallery) OnMessage(content string, msg *discordgo.Message, session *discordgo.Session) {
	go func() {
		defer helpers.Recover()
	TryNextGallery:
		for _, gallery := range galleries {
			if gallery.SourceChannelID == msg.ChannelID {
				// ignore bot messages
				if msg.Author.Bot == true {
					continue TryNextGallery
				}
				sourceChannel, err := helpers.GetChannel(msg.ChannelID)
				helpers.Relax(err)
				// ignore commands
				prefix := helpers.GetPrefixForServer(sourceChannel.GuildID)
				if prefix != "" {
					if strings.HasPrefix(content, prefix) {
						return
					}
				}
				var linksToRepost []string
				// get mirror attachements
				if len(msg.Attachments) > 0 {
					for _, attachement := range msg.Attachments {
						linksToRepost = append(linksToRepost, attachement.URL)
					}
				}
				// get mirror links
				if strings.Contains(msg.Content, "http") {
					linksFound := galleryUrlRegex.FindAllString(msg.Content, -1)
					if len(linksFound) > 0 {
						for _, linkFound := range linksFound {
							if strings.HasPrefix(linkFound, "<") == false && strings.HasSuffix(linkFound, ">") == false {
								linksToRepost = append(linksToRepost, linkFound)
							}
						}
					}
				}
				// post mirror links
				if len(linksToRepost) > 0 {
					for _, linkToRepost := range linksToRepost {
						result, err := helpers.WebhookExecuteWithResult(
							gallery.TargetChannelWebhookID,
							gallery.TargetChannelWebhookToken,
							&discordgo.WebhookParams{
								Content:   fmt.Sprintf("posted %s", linkToRepost),
								Username:  msg.Author.Username,
								AvatarURL: helpers.GetAvatarUrl(msg.Author),
							},
						)
						if err != nil {
							if errD, ok := err.(*discordgo.RESTError); ok {
								if errD.Message.Code == 10015 {
									cache.GetLogger().WithField("module", "gallery").Error(fmt.Sprintf("Webhook for gallery #%s not found", gallery.ID))
									continue
								}
							}
							raven.CaptureError(fmt.Errorf("%#v", err), map[string]string{})
							continue
						}
						err = g.rememberPostedMessage(msg, result)
						if err != nil {
							raven.CaptureError(fmt.Errorf("%#v", err), map[string]string{})
						}
						metrics.GalleryPostsSent.Add(1)
					}
				}
			}
		}
	}()
}

type Gallery_PostedMessage struct {
	ChannelID string
	MessageID string
}

func (g *Gallery) rememberPostedMessage(sourceMessage *discordgo.Message, mirroredMessage *discordgo.Message) error {
	redis := cache.GetRedisClient()
	key := fmt.Sprintf("robyul2-discord:gallery:postedmessage:%s", sourceMessage.ID)

	item := new(Gallery_PostedMessage)
	item.ChannelID = mirroredMessage.ChannelID
	item.MessageID = mirroredMessage.ID

	itemBytes, err := msgpack.Marshal(&item)
	if err != nil {
		return err
	}

	_, err = redis.LPush(key, itemBytes).Result()
	if err != nil {
		return err
	}

	_, err = redis.Expire(key, time.Hour*1).Result()
	return err
}

func (g *Gallery) getRememberedMessages(sourceMessage *discordgo.Message) ([]Gallery_PostedMessage, error) {
	redis := cache.GetRedisClient()
	key := fmt.Sprintf("robyul2-discord:gallery:postedmessage:%s", sourceMessage.ID)

	length, err := redis.LLen(key).Result()
	if err != nil {
		return []Gallery_PostedMessage{}, err
	}

	if length <= 0 {
		return []Gallery_PostedMessage{}, err
	}

	result, err := redis.LRange(key, 0, length-1).Result()
	if err != nil {
		return []Gallery_PostedMessage{}, err
	}

	rememberedMessages := make([]Gallery_PostedMessage, 0)
	for _, messageData := range result {
		var message Gallery_PostedMessage
		err = msgpack.Unmarshal([]byte(messageData), &message)
		if err != nil {
			continue
		}
		rememberedMessages = append(rememberedMessages, message)
	}

	return rememberedMessages, nil
}

func (g *Gallery) OnGuildMemberAdd(member *discordgo.Member, session *discordgo.Session) {
}

func (g *Gallery) OnGuildMemberRemove(member *discordgo.Member, session *discordgo.Session) {
}

func (g *Gallery) getEntryBy(key string, id string) DB_Gallery_Entry {
	var entryBucket DB_Gallery_Entry
	listCursor, err := rethink.Table("galleries").Filter(
		rethink.Row.Field(key).Eq(id),
	).Run(helpers.GetDB())
	if err != nil {
		panic(err)
	}
	defer listCursor.Close()
	err = listCursor.One(&entryBucket)

	if err == rethink.ErrEmptyResult {
		return entryBucket
	} else if err != nil {
		panic(err)
	}

	return entryBucket
}

func (g *Gallery) getEntryByOrCreateEmpty(key string, id string) DB_Gallery_Entry {
	var entryBucket DB_Gallery_Entry
	listCursor, err := rethink.Table("galleries").Filter(
		rethink.Row.Field(key).Eq(id),
	).Run(helpers.GetDB())
	if err != nil {
		panic(err)
	}
	defer listCursor.Close()
	err = listCursor.One(&entryBucket)

	if err == rethink.ErrEmptyResult {
		insert := rethink.Table("galleries").Insert(DB_Gallery_Entry{})
		res, e := insert.RunWrite(helpers.GetDB())
		if e != nil {
			panic(e)
		} else {
			return g.getEntryByOrCreateEmpty("id", res.GeneratedKeys[0])
		}
	} else if err != nil {
		panic(err)
	}

	return entryBucket
}

func (g *Gallery) setEntry(entry DB_Gallery_Entry) {
	_, err := rethink.Table("galleries").Update(entry).Run(helpers.GetDB())
	helpers.Relax(err)
}

func (g *Gallery) deleteEntryById(id string) {
	_, err := rethink.Table("galleries").Filter(
		rethink.Row.Field("id").Eq(id),
	).Delete().RunWrite(helpers.GetDB())
	helpers.Relax(err)
}

func (g *Gallery) GetGalleries() []DB_Gallery_Entry {
	var entryBucket []DB_Gallery_Entry
	listCursor, err := rethink.Table("galleries").Run(helpers.GetDB())
	helpers.Relax(err)
	defer listCursor.Close()
	err = listCursor.All(&entryBucket)

	helpers.Relax(err)
	return entryBucket
}

func (g *Gallery) OnReactionAdd(reaction *discordgo.MessageReactionAdd, session *discordgo.Session) {

}
func (g *Gallery) OnReactionRemove(reaction *discordgo.MessageReactionRemove, session *discordgo.Session) {

}
func (g *Gallery) OnGuildBanAdd(user *discordgo.GuildBanAdd, session *discordgo.Session) {

}
func (g *Gallery) OnGuildBanRemove(user *discordgo.GuildBanRemove, session *discordgo.Session) {

}
func (g *Gallery) OnMessageDelete(msg *discordgo.MessageDelete, session *discordgo.Session) {
	go func() {
		defer helpers.Recover()
		var err error
		var rememberedMessages []Gallery_PostedMessage

		for _, gallery := range galleries {
			if gallery.SourceChannelID == msg.ChannelID {
				rememberedMessages, err = g.getRememberedMessages(msg.Message)
				helpers.Relax(err)

				for _, messageData := range rememberedMessages {
					err = session.ChannelMessageDelete(messageData.ChannelID, messageData.MessageID)
					if err != nil {
						msgAuthorID := "N/A"
						if msg.Author != nil {
							msgAuthorID = msg.Author.ID
						}

						cache.GetLogger().WithFields(logrus.Fields{
							"module":            "gallery",
							"sourceChannelID":   msg.ChannelID,
							"sourceMessageID":   msg.ID,
							"sourceAuthorID":    msgAuthorID,
							"mirroredChannelID": messageData.ChannelID,
							"mirroredMessageID": messageData.MessageID,
						}).Error(
							"Deleting mirrored message failed:", err.Error(),
						)
					}
				}
			}
		}
	}()
}

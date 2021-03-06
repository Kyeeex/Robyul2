package helpers

import (
	"errors"
	"fmt"
	"math/big"
	"regexp"
	"strconv"
	"strings"
	"time"

	"encoding/json"

	"github.com/RichardKnop/machinery/v1/tasks"
	"github.com/Seklfreak/Robyul2/cache"
	"github.com/Seklfreak/Robyul2/models"
	"github.com/bwmarrin/discordgo"
	"github.com/getsentry/raven-go"
	redisCache "github.com/go-redis/cache"
	rethink "github.com/gorethink/gorethink"
	"github.com/vmihailenco/msgpack"
)

const (
	DISCORD_EPOCH int64 = 1420070400000
)

var botAdmins = []string{
	"116620585638821891", // Sekl
	"134298438559858688", // Kakkela
}
var NukeMods = []string{
	"116620585638821891", // Sekl
	"134298438559858688", // Kakkela
	"68661361537712128",  // Berk
}
var RobyulMod = []string{
	"132633380628987904", // sunny
}
var Blacklisted = []string{
	"171883318386753536", // ForRyu
}
var ExtendedInspectRoleIDs = []string{
	"345209385821274113", // inspect extended (sekl's dev cord)
	"345209098100277248", // inspect (Moderator Chat)
}
var adminRoleNames = []string{"Admin", "Admins", "ADMIN", "School Board", "admin", "admins"}
var modRoleNames = []string{"Mod", "Mods", "Mod Trainee", "Moderator", "Moderators", "MOD", "Minimod", "Guard", "Janitor", "mod", "mods"}

func IsBlacklisted(id string) bool {
	for _, s := range Blacklisted {
		if s == id {
			return true
		}
	}

	return false
}

func IsNukeMod(id string) bool {
	for _, s := range NukeMods {
		if s == id {
			return true
		}
	}

	return false
}

// IsBotAdmin checks if $id is in $botAdmins
func IsBotAdmin(id string) bool {
	for _, s := range botAdmins {
		if s == id {
			return true
		}
	}

	return false
}

func IsRobyulMod(id string) bool {
	if IsBotAdmin(id) {
		return true
	}
	for _, s := range RobyulMod {
		if s == id {
			return true
		}
	}

	return false
}

func CanInspectExtended(msg *discordgo.Message) bool {
	if IsBotAdmin(msg.Author.ID) {
		return true
	}

	if IsRobyulMod(msg.Author.ID) {
		return true
	}

	if IsNukeMod(msg.Author.ID) {
		return true
	}

	channel, e := GetChannel(msg.ChannelID)
	if e != nil {
		return false
	}

	guild, e := GetGuild(channel.GuildID)
	if e != nil {
		return false
	}

	guildMember, e := GetGuildMember(guild.ID, msg.Author.ID)
	if e != nil {
		return false
	}
	for _, role := range guild.Roles {
		for _, userRole := range guildMember.Roles {
			if userRole == role.ID {
				for _, inspectRoleID := range ExtendedInspectRoleIDs {
					if role.ID == inspectRoleID {
						return true
					}
				}
			}
		}
	}
	return false
}

func IsAdmin(msg *discordgo.Message) bool {
	channel, e := GetChannel(msg.ChannelID)
	if e != nil {
		return false
	}

	guild, e := GetGuild(channel.GuildID)
	if e != nil {
		return false
	}

	if msg.Author.ID == guild.OwnerID || IsBotAdmin(msg.Author.ID) {
		return true
	}

	guildMember, e := GetGuildMember(guild.ID, msg.Author.ID)
	if e != nil {
		return false
	}
	// Check if role may manage server or a role is in admin role list
	for _, role := range guild.Roles {
		for _, userRole := range guildMember.Roles {
			if userRole == role.ID {
				if role.Permissions&discordgo.PermissionAdministrator == discordgo.PermissionAdministrator {
					return true
				}
				for _, adminRoleName := range adminRoleNames {
					if role.Name == adminRoleName {
						return true
					}
				}
			}
		}
	}
	return false
}

func IsAdminByID(guildID string, userID string) bool {
	guild, e := GetGuild(guildID)
	if e != nil {
		return false
	}

	if userID == guild.OwnerID || IsBotAdmin(userID) {
		return true
	}

	guildMember, e := GetGuildMember(guild.ID, userID)
	if e != nil {
		return false
	}
	// Check if role may manage server or a role is in admin role list
	for _, role := range guild.Roles {
		for _, userRole := range guildMember.Roles {
			if userRole == role.ID {
				if role.Permissions&discordgo.PermissionAdministrator == discordgo.PermissionAdministrator {
					return true
				}
				for _, adminRoleName := range adminRoleNames {
					if role.Name == adminRoleName {
						return true
					}
				}
			}
		}
	}
	return false
}

func HasPermissionByID(guildID string, userID string, permission int) bool {
	guild, e := GetGuild(guildID)
	if e != nil {
		return false
	}

	if userID == guild.OwnerID {
		return true
	}

	guildMember, e := GetGuildMember(guild.ID, userID)
	if e != nil {
		return false
	}
	for _, role := range guild.Roles {
		for _, userRole := range guildMember.Roles {
			if userRole == role.ID {
				if role.Permissions&permission == permission {
					return true
				}
			}
		}
	}
	return false
}

func IsMod(msg *discordgo.Message) bool {
	if IsAdmin(msg) == true {
		return true
	} else {
		channel, e := GetChannel(msg.ChannelID)
		if e != nil {
			return false
		}
		guild, e := GetGuild(channel.GuildID)
		if e != nil {
			return false
		}
		guildMember, e := GetGuildMember(guild.ID, msg.Author.ID)
		if e != nil {
			return false
		}
		// check if a role is in mod role list
		for _, role := range guild.Roles {
			for _, userRole := range guildMember.Roles {
				if userRole == role.ID {
					for _, modRoleName := range modRoleNames {
						if role.Name == modRoleName {
							return true
						}
					}
				}
			}
		}
	}
	return false
}

func IsModByID(guildID string, userID string) bool {
	if IsAdminByID(guildID, userID) == true {
		return true
	} else {
		guild, e := GetGuild(guildID)
		if e != nil {
			return false
		}
		guildMember, e := GetGuildMember(guild.ID, userID)
		if e != nil {
			return false
		}
		// check if a role is in mod role list
		for _, role := range guild.Roles {
			for _, userRole := range guildMember.Roles {
				if userRole == role.ID {
					for _, modRoleName := range modRoleNames {
						if role.Name == modRoleName {
							return true
						}
					}
				}
			}
		}
	}
	return false
}

// RequireAdmin only calls $cb if the author is an admin or has MANAGE_SERVER permission
func RequireAdmin(msg *discordgo.Message, cb Callback) {
	if !IsAdmin(msg) {
		SendMessage(msg.ChannelID, GetText("admin.no_permission"))
		return
	}

	cb()
}

// RequireAdmin only calls $cb if the author is an admin or has MANAGE_SERVER permission
func RequireMod(msg *discordgo.Message, cb Callback) {
	if !IsMod(msg) {
		SendMessage(msg.ChannelID, GetText("mod.no_permission"))
		return
	}

	cb()
}

// RequireBotAdmin only calls $cb if the author is a bot admin
func RequireBotAdmin(msg *discordgo.Message, cb Callback) {
	if !IsBotAdmin(msg.Author.ID) {
		SendMessage(msg.ChannelID, GetText("botadmin.no_permission"))
		return
	}

	cb()
}

// RequireSupportMod only calls $cb if the author is a support mod
func RequireRobyulMod(msg *discordgo.Message, cb Callback) {
	if !IsRobyulMod(msg.Author.ID) {
		SendMessage(msg.ChannelID, GetText("robyulmod.no_permission"))
		return
	}

	cb()
}

func ConfirmEmbed(channelID string, author *discordgo.User, confirmMessageText string, confirmEmojiID string, abortEmojiID string) bool {
	// send embed asking the user to confirm
	confirmMessages, err := SendComplex(channelID,
		&discordgo.MessageSend{
			Content: "<@" + author.ID + ">",
			Embed: &discordgo.MessageEmbed{
				Title:       GetText("bot.embeds.please-confirm-title"),
				Description: confirmMessageText,
			},
		})
	if err != nil {
		SendMessage(channelID, GetTextF("bot.errors.general", err.Error()))
		return false
	}
	if len(confirmMessages) <= 0 {
		SendMessage(channelID, GetText("bot.errors.generic-nomessage"))
		return false
	}
	confirmMessage := confirmMessages[0]
	if len(confirmMessage.Embeds) <= 0 {
		SendMessage(channelID, GetText("bot.errors.no-embed"))
		return false
	}

	// add default reactions to embed
	cache.GetSession().MessageReactionAdd(confirmMessage.ChannelID, confirmMessage.ID, confirmEmojiID)
	cache.GetSession().MessageReactionAdd(confirmMessage.ChannelID, confirmMessage.ID, abortEmojiID)

	// check every second if a reaction has been clicked
	for {
		confirmes, _ := cache.GetSession().MessageReactions(confirmMessage.ChannelID, confirmMessage.ID, confirmEmojiID, 100)
		for _, confirm := range confirmes {
			if confirm.ID == author.ID {
				cache.GetSession().ChannelMessageDelete(confirmMessage.ChannelID, confirmMessage.ID)
				// user has confirmed the call
				return true
			}
		}
		aborts, _ := cache.GetSession().MessageReactions(confirmMessage.ChannelID, confirmMessage.ID, abortEmojiID, 100)
		for _, abort := range aborts {
			if abort.ID == author.ID {
				cache.GetSession().ChannelMessageDelete(confirmMessage.ChannelID, confirmMessage.ID)
				// User has aborted the call
				return false
			}
		}

		time.Sleep(1 * time.Second)
	}
}

func GetMuteRole(guildID string) (*discordgo.Role, error) {
	guild, err := GetGuild(guildID)
	Relax(err)
	var muteRole *discordgo.Role
	settings, err := GuildSettingsGet(guildID)
	for _, role := range guild.Roles {
		Relax(err)
		if role.Name == settings.MutedRoleName {
			muteRole = role
		}
	}
	if muteRole == nil {
		muteRole, err = cache.GetSession().GuildRoleCreate(guildID)
		if err != nil {
			return muteRole, err
		}
		muteRole, err = cache.GetSession().GuildRoleEdit(guildID, muteRole.ID, settings.MutedRoleName, muteRole.Color, muteRole.Hoist, 0, muteRole.Mentionable)
		if err != nil {
			return muteRole, err
		}
		for _, channel := range guild.Channels {
			err = cache.GetSession().ChannelPermissionSet(channel.ID, muteRole.ID, "role", 0,
				discordgo.PermissionSendMessages+discordgo.PermissionAddReactions)
			if err != nil {
				cache.GetLogger().WithField("module", "discord").Error("Error disabling send messages and add reactions on mute Role: " + err.Error())
			}
		}
	}
	return muteRole, nil
}

func RemoveMuteRole(guildID string, userID string) (err error) {
	muteRole, err := GetMuteRole(guildID)
	if err != nil {
		return err
	}

	err = cache.GetSession().GuildMemberRoleRemove(guildID, userID, muteRole.ID)
	if err != nil {
		if errD, ok := err.(*discordgo.RESTError); ok {
			if errD.Message.Code == discordgo.ErrCodeUnknownMember ||
				errD.Message.Code == discordgo.ErrCodeUnknownUser ||
				errD.Response.StatusCode == 404 {
				return nil
			}
		}
	}
	return err
}

func RemoveMuteDatabase(guildID string, userID string) (err error) {
	settings := GuildSettingsGetCached(guildID)

	removedFromDb := false
	newMutedMembers := make([]string, 0)
	for _, mutedMember := range settings.MutedMembers {
		if mutedMember != userID {
			newMutedMembers = append(newMutedMembers, mutedMember)
		} else {
			removedFromDb = true
		}
	}

	if removedFromDb {
		settings.MutedMembers = newMutedMembers
		err = GuildSettingsSet(guildID, settings)
		return err
	}
	return nil
}

func RemoveMutePersistency(guildID string, userID string) (err error) {
	muteRole, err := GetMuteRole(guildID)
	if err != nil {
		return err
	}

	return persistencyRemoveCachedRole(guildID, userID, muteRole.ID)
}

func UnmuteUser(guildID string, userID string) (err error) {
	errRole := RemoveMuteRole(guildID, userID)
	errDatabase := RemoveMuteDatabase(guildID, userID)
	errPersistency := RemoveMutePersistency(guildID, userID)

	if errRole != nil {
		return errRole
	}
	if errDatabase != nil {
		return errDatabase
	}
	if errPersistency != nil {
		return errPersistency
	}
	return nil
}
func UnmuteUserSignature(guildID string, userID string) (signature *tasks.Signature) {
	signature = &tasks.Signature{
		Name: "unmute_user",
		Args: []tasks.Arg{
			{
				Type:  "string",
				Value: guildID,
			},
			{
				Type:  "string",
				Value: userID,
			},
		},
	}
	signature.RetryCount = 3
	signature.OnError = []*tasks.Signature{{Name: "log_error"}}
	return signature
}

func persistencyRemoveCachedRole(GuildID string, UserID string, roleID string) (err error) {
	key := "robyul2-discord:persistency:" + GuildID + ":" + UserID + ":roles"
	var redisRoleIDs []string
	var dbRoles models.PersistencyRolesEntry

	// remove from db
	listCursor, _ := rethink.Table(models.PersistencyRolesTable).Filter(
		rethink.And(
			rethink.Row.Field("guild_id").Eq(GuildID),
			rethink.Row.Field("user_id").Eq(UserID),
		),
	).Run(GetDB())
	defer listCursor.Close()
	listCursor.One(&dbRoles)

	newDbRoles := dbRoles
	newDbRoles.Roles = make([]string, 0)
	for _, dbRoleID := range dbRoles.Roles {
		if dbRoleID != roleID {
			newDbRoles.Roles = append(newDbRoles.Roles, dbRoleID)
		}
	}

	_, err = rethink.Table(models.PersistencyRolesTable).Update(newDbRoles).Run(GetDB())
	if err != nil {
		return err
	}

	// remove from redis
	marshalled, err := cache.GetRedisClient().Get(key).Bytes()
	if err != nil {
		return err
	}

	err = msgpack.Unmarshal(marshalled, &redisRoleIDs)
	if err != nil {
		return err
	}

	newRedisRoleIDs := make([]string, 0)
	for _, redisRoleID := range redisRoleIDs {
		if redisRoleID != roleID {
			newRedisRoleIDs = append(newRedisRoleIDs, redisRoleID)
		}
	}

	marshalled, err = msgpack.Marshal(newRedisRoleIDs)
	if err != nil {
		return
	}

	err = cache.GetRedisClient().Set(key, marshalled, 0).Err()

	return err
}

func LogMachineryError(errorMessage string) (err error) {
	cache.GetLogger().WithField("module", "machinery").Error("Task Failed: ", errorMessage)
	raven.CaptureError(errors.New(errorMessage), map[string]string{})
	return err
}

func GetGuildMember(guildID string, userID string) (*discordgo.Member, error) {
	targetMember, err := cache.GetSession().State.Member(guildID, userID)
	if targetMember == nil || targetMember.GuildID == "" || targetMember.JoinedAt == "" {
		cache.GetLogger().WithField("module", "discord").WithField("method", "GetGuildMember").Debug(
			fmt.Sprintf("discord api request: GuildMember: %s, %s", guildID, userID))
		targetMember, err = cache.GetSession().GuildMember(guildID, userID)
	}
	return targetMember, err
}

func GetGuildMemberWithoutApi(guildID string, userID string) (*discordgo.Member, error) {
	return cache.GetSession().State.Member(guildID, userID)
}

func GetIsInGuild(guildID string, userID string) bool {
	member, err := GetGuildMemberWithoutApi(guildID, userID)
	if err == nil && member != nil && member.User != nil && member.User.ID != "" {
		return true
	} else {
		return false
	}
}

func GetGuild(guildID string) (*discordgo.Guild, error) {
	targetGuild, err := cache.GetSession().State.Guild(guildID)
	if targetGuild == nil || targetGuild.ID == "" {
		cache.GetLogger().WithField("module", "discord").WithField("method", "GetGuild").Debug(
			fmt.Sprintf("discord api request: Guild: %s", guildID))
		targetGuild, err = cache.GetSession().Guild(guildID)
	}
	return targetGuild, err
}

func GetChannel(channelID string) (*discordgo.Channel, error) {
	targetChannel, err := cache.GetSession().State.Channel(channelID)
	if targetChannel == nil || targetChannel.ID == "" {
		cache.GetLogger().WithField("module", "discord").WithField("method", "GetChannel").Debug(
			fmt.Sprintf("discord api request: Channel: %s", channelID))
		targetChannel, err = cache.GetSession().Channel(channelID)
	}
	return targetChannel, err
}

func GetChannelWithoutApi(channelID string) (*discordgo.Channel, error) {
	targetChannel, err := cache.GetSession().State.Channel(channelID)
	return targetChannel, err
}

func GetMessage(channelID string, messageID string) (*discordgo.Message, error) {
	targetMessage, err := cache.GetSession().State.Message(channelID, messageID)
	if targetMessage == nil || targetMessage.ID == "" {
		cache.GetLogger().WithField("module", "discord").WithField("method", "GetMessage").Debug(
			fmt.Sprintf("discord api request: Message: %s in Channel: %s", messageID, channelID))
		targetMessage, err = cache.GetSession().ChannelMessage(channelID, messageID)
		cache.GetSession().State.MessageAdd(targetMessage)
		return targetMessage, err
	}
	return targetMessage, nil
}

func GetChannelFromMention(msg *discordgo.Message, mention string) (*discordgo.Channel, error) {
	var targetChannel *discordgo.Channel
	re := regexp.MustCompile("(<#)?(\\d+)(>)?")
	result := re.FindStringSubmatch(mention)
	if len(result) == 4 {
		sourceChannel, err := GetChannel(msg.ChannelID)
		if err != nil {
			return targetChannel, err
		}
		if sourceChannel == nil {
			return targetChannel, errors.New("Channel not found.")
		}
		targetChannel, err := GetChannel(result[2])
		if err != nil {
			return targetChannel, err
		}
		if targetChannel.Type != discordgo.ChannelTypeGuildText {
			return targetChannel, errors.New("not a text channel")
		}
		if sourceChannel.GuildID != targetChannel.GuildID {
			return targetChannel, errors.New("Channel on different guild.")
		}
		return targetChannel, err
	} else {
		return targetChannel, errors.New("Channel not found.")
	}
}

func GetGlobalChannelFromMention(mention string) (*discordgo.Channel, error) {
	var targetChannel *discordgo.Channel
	re := regexp.MustCompile("(<#)?(\\d+)(>)?")
	result := re.FindStringSubmatch(mention)
	if len(result) == 4 {
		targetChannel, err := GetChannel(result[2])
		if err != nil {
			return targetChannel, err
		}
		return targetChannel, err
	} else {
		return targetChannel, errors.New("Channel not found.")
	}
}

func GetUser(userID string) (*discordgo.User, error) {
	var err error
	var targetUser discordgo.User
	cacheCodec := cache.GetRedisCacheCodec()
	key := fmt.Sprintf("robyul2-discord:api:user:%s", userID) // TODO: Should we cache this?

	for _, guild := range cache.GetSession().State.Guilds {
		member, err := GetGuildMemberWithoutApi(guild.ID, userID)
		if err == nil && member != nil && member.User != nil && member.User.ID != "" {
			return member.User, nil
		}
	}

	if err = cacheCodec.Get(key, &targetUser); err != nil {
		cache.GetLogger().WithField("module", "discord").WithField("method", "GetUser").Debug(
			fmt.Sprintf("discord api request: User: %s", userID))
		targetUser, err := cache.GetSession().User(userID)
		if err == nil {
			err = cacheCodec.Set(&redisCache.Item{
				Key:        key,
				Object:     targetUser,
				Expiration: time.Minute * 10,
			})
			if err != nil {
				raven.CaptureError(fmt.Errorf("%#v", err), map[string]string{})
			}
		}
		return targetUser, err
	}
	return &targetUser, err
}

func GetUserFromMention(mention string) (*discordgo.User, error) {
	re := regexp.MustCompile("(<@)?(\\d+)(>)?")
	result := re.FindStringSubmatch(mention)
	if len(result) == 4 {
		return GetUser(result[2])
	} else {
		return &discordgo.User{}, errors.New("User not found.")
	}
}

func GetDiscordColorFromHex(hex string) int {
	colorInt, ok := new(big.Int).SetString(strings.Replace(hex, "#", "", 1), 16)
	if ok == true {
		return int(colorInt.Int64())
	} else {
		return 0x0FADED
	}
}

func GetHexFromDiscordColor(colour int) (hex string) {
	return strings.ToUpper(big.NewInt(int64(colour)).Text(16))
}

func GetTimeFromSnowflake(id string) time.Time {
	iid, err := strconv.ParseInt(id, 10, 64)
	Relax(err)

	return time.Unix(((iid>>22)+DISCORD_EPOCH)/1000, 0).UTC()
}

func GetAllPermissions(guild *discordgo.Guild, member *discordgo.Member) int64 {
	var perms int64 = 0
	for _, x := range guild.Roles {
		if x.Name == "@everyone" {
			perms |= int64(x.Permissions)
		}
	}
	for _, r := range member.Roles {
		for _, x := range guild.Roles {
			if x.ID == r {
				perms |= int64(x.Permissions)
			}
		}
	}
	return perms
}

func Pagify(text string, delimiter string) []string {
	result := make([]string, 0)
	textParts := strings.Split(text, delimiter)
	currentOutputPart := ""
	for _, textPart := range textParts {
		if len(currentOutputPart)+len(textPart)+len(delimiter) <= 1992 {
			if len(currentOutputPart) > 0 || len(result) > 0 {
				currentOutputPart += delimiter + textPart
			} else {
				currentOutputPart += textPart
			}
		} else {
			result = append(result, currentOutputPart)
			currentOutputPart = ""
			if len(textPart) <= 1992 { // @TODO: else: split text somehow
				currentOutputPart = textPart
			}
		}
	}
	if currentOutputPart != "" {
		result = append(result, currentOutputPart)
	}
	return result
}

func GetAvatarUrl(user *discordgo.User) string {
	return GetAvatarUrlWithSize(user, 1024)
}

func GetAvatarUrlWithSize(user *discordgo.User, size uint16) string {
	if user.Avatar == "" {
		return ""
	}

	avatarUrl := "https://cdn.discordapp.com/avatars/%s/%s.%s?size=%d"

	if strings.HasPrefix(user.Avatar, "a_") {
		return fmt.Sprintf(avatarUrl, user.ID, user.Avatar, "gif", size)
	}

	return fmt.Sprintf(avatarUrl, user.ID, user.Avatar, "jpg", size)
}

func CommandExists(name string) bool {
	for _, command := range cache.GetPluginList() {
		if command == strings.ToLower(name) {
			return true
		}
	}
	for _, command := range cache.GetPluginExtendedList() {
		if command == strings.ToLower(name) {
			return true
		}
	}
	for _, command := range cache.GetTriggerPluginList() {
		if command == strings.ToLower(name) {
			return true
		}
	}
	return false
}

func WebhookExecuteWithResult(webhookID, token string, data *discordgo.WebhookParams) (message *discordgo.Message, err error) {
	uri := discordgo.EndpointWebhookToken(webhookID, token) + "?wait=true"

	result, err := cache.GetSession().RequestWithBucketID("POST", uri, data, discordgo.EndpointWebhookToken("", ""))
	if err != nil {
		return message, err
	}

	err = json.Unmarshal(result, &message)
	return message, err
}

func GuildIsOnWhitelist(GuildID string) (whitelisted bool) {
	var entryBucket []models.AutoleaverWhitelistEntry
	listCursor, err := rethink.Table(models.AutoleaverWhitelistTable).Run(GetDB())
	if err != nil {
		return false
	}

	defer listCursor.Close()
	err = listCursor.All(&entryBucket)
	if err != nil {
		return false
	}

	for _, whitelistEntry := range entryBucket {
		if whitelistEntry.GuildID == GuildID {
			return true
		}
	}

	return false
}

func AutoPagify(text string) (pages []string) {
	for _, page := range Pagify(text, "\n") {
		if len(page) <= 1992 {
			pages = append(pages, page)
		} else {
			for _, page := range Pagify(page, ",") {
				if len(page) <= 1992 {
					pages = append(pages, page)
				} else {
					for _, page := range Pagify(page, "-") {
						if len(page) <= 1992 {
							pages = append(pages, page)
						} else {
							for _, page := range Pagify(page, " ") {
								if len(page) <= 1992 {
									pages = append(pages, page)
								} else {
									panic("unable to pagify text")
								}
							}
						}
					}
				}
			}
		}
	}
	return pages
}

func SendMessage(channelID, content string) (messages []*discordgo.Message, err error) {
	var message *discordgo.Message
	for _, page := range AutoPagify(content) {
		message, err = cache.GetSession().ChannelMessageSend(channelID, page)
		if err != nil {
			return messages, err
		}
		messages = append(messages, message)
	}
	return messages, nil
}

// TODO: implement https://discordapp.com/developers/docs/resources/channel#embed-limits
func SendEmbed(channelID string, embed *discordgo.MessageEmbed) (messages []*discordgo.Message, err error) {
	var message *discordgo.Message
	message, err = cache.GetSession().ChannelMessageSendEmbed(channelID, embed)
	if err != nil {
		return messages, err
	}
	messages = append(messages, message)
	return messages, nil
}

// TODO: implement https://discordapp.com/developers/docs/resources/channel#embed-limits
func SendComplex(channelID string, data *discordgo.MessageSend) (messages []*discordgo.Message, err error) {
	var message *discordgo.Message
	pages := AutoPagify(data.Content)
	if len(pages) > 0 {
		for i, page := range pages {
			if i+1 < len(pages) {
				message, err = cache.GetSession().ChannelMessageSend(channelID, page)
			} else {
				data.Content = page
				message, err = cache.GetSession().ChannelMessageSendComplex(channelID, data)
			}
			if err != nil {
				return messages, err
			}
			messages = append(messages, message)
		}
	} else {
		message, err = cache.GetSession().ChannelMessageSendComplex(channelID, data)
		messages = append(messages, message)
		if err != nil {
			return messages, err
		}
	}
	return messages, nil
}

func EditMessage(channelID, messageID, content string) (message *discordgo.Message, err error) {
	message, err = cache.GetSession().ChannelMessageEdit(channelID, messageID, content)
	if err != nil {
		return nil, err
	} else {
		return message, err
	}
}

func EditEmbed(channelID, messageID string, embed *discordgo.MessageEmbed) (message *discordgo.Message, err error) {
	message, err = cache.GetSession().ChannelMessageEditEmbed(channelID, messageID, embed)
	if err != nil {
		return nil, err
	} else {
		return message, err
	}
}

func EditComplex(data *discordgo.MessageEdit) (message *discordgo.Message, err error) {
	message, err = cache.GetSession().ChannelMessageEditComplex(data)
	if err != nil {
		return nil, err
	} else {
		return message, err
	}
}

// TODO: Webhook

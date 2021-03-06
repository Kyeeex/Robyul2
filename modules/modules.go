package modules

import (
	"github.com/Seklfreak/Robyul2/modules/plugins"
	"github.com/Seklfreak/Robyul2/modules/plugins/youtube"
	//"github.com/Seklfreak/Robyul2/modules/triggers"
	"github.com/Seklfreak/Robyul2/modules/plugins/google"
	"github.com/Seklfreak/Robyul2/modules/triggers"
)

var (
	pluginCache         map[string]*Plugin
	triggerCache        map[string]*TriggerPlugin
	extendedPluginCache map[string]*ExtendedPlugin

	PluginList = []Plugin{
		&plugins.About{},
		&plugins.Stats{},
		//&plugins.Stone{},
		//&plugins.Support{},
		&plugins.Announcement{},
		//&plugins.Translator{},
		&plugins.Uptime{},
		&plugins.Translator{},
		&plugins.UrbanDict{},
		&plugins.Weather{},
		&plugins.VLive{},
		&plugins.Instagram{},
		&plugins.Facebook{},
		&plugins.WolframAlpha{},
		&plugins.LastFm{},
		&plugins.Twitch{},
		&plugins.Charts{},
		//&plugins.Avatar{},
		//&plugins.Calc{},
		//&plugins.Changelog{},
		&plugins.Choice{},
		//&plugins.FlipCoin{},
		//&plugins.Giphy{},
		//&plugins.Google{},
		//&plugins.Leet{},
		//&plugins.ListenDotMoe{},
		//&plugins.Minecraft{},
		&plugins.Osu{},
		&plugins.Reminders{},
		//&plugins.Roll{},
		//&plugins.Stone{},
		//&plugins.Support{},
		//&plugins.XKCD{},
		&plugins.Ratelimit{},
		&plugins.Gfycat{},
		&plugins.RandomPictures{},
		&youtube.Handler{},
		// &plugins.YouTube{},
		&plugins.Spoiler{},
		&plugins.RandomCat{},
		&plugins.RPS{},
		&plugins.Nuke{},
		&plugins.Dig{},
		&plugins.Streamable{},
		&plugins.Troublemaker{},
		&plugins.Lyrics{},
		&plugins.Friend{},
		&plugins.Names{},
		&plugins.Reddit{},
		&plugins.Color{},
		&plugins.Dog{},
		&plugins.Debug{},
		&plugins.Donators{},
		&plugins.Ping{},
		&google.Handler{},
		&plugins.BotStatus{},
	}

	// PluginList is the list of active plugins
	PluginExtendedList = []ExtendedPlugin{
		&plugins.Bias{},
		&plugins.GuildAnnouncements{},
		&plugins.Notifications{},
		&plugins.Levels{},
		&plugins.Gallery{},
		&plugins.Mirror{},
		&plugins.CustomCommands{},
		&plugins.ReactionPolls{},
		&plugins.Mod{},
		&plugins.AutoRoles{},
		&plugins.Starboard{},
		&plugins.Autoleaver{},
		&plugins.Persistency{},
		&plugins.DM{},
		&plugins.Twitter{},
	}

	// TriggerPluginList is the list of plugins that activate on normal chat
	TriggerPluginList = []TriggerPlugin{
		//&triggers.CSS{},
		//&triggers.Donate{},
		//&triggers.Git{},
		&triggers.EightBall{},
		//&triggers.Hi{},
		//&triggers.HypeTrain{},
		//&triggers.Invite{},
		//&triggers.IPTables{},
		//&triggers.Lenny{},
		//&triggers.Nep{},
		//&triggers.ReZero{},
		//&triggers.Shrug{},
		//&triggers.TableFlip{},
		//&triggers.Triggered{},
	}
)

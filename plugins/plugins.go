package plugins

import (
    "github.com/sn0w/discordgo"
    "git.lukas.moe/sn0w/Karen/cache"
    "git.lukas.moe/sn0w/Karen/helpers"
    "git.lukas.moe/sn0w/Karen/metrics"
    "strings"
    "git.lukas.moe/sn0w/Karen/plugins/triggers"
)

// Plugin interface to enforce a basic structure
type Plugin interface {
    // List of commands and aliases
    Commands() []string

    // Plugin constructor
    Init(session *discordgo.Session)

    // Action to execute on message receive
    Action(
    command string,
    content string,
    msg *discordgo.Message,
    session *discordgo.Session,
    )
}

type TriggerPlugin interface {
    Triggers() []string
    Response(trigger string, content string) string
}

// PluginList is the list of active plugins
var PluginList = []Plugin{
    About{},
    Avatar{},
    Calc{},
    &Changelog{},
    Choice{},
    &FlipCoin{},
    Giphy{},
    Google{},
    Leet{},
    Lenny{},
    &ListenDotMoe{},
    Minecraft{},
    &Music{},
    Osu{},
    Ping{},
    RandomCat{},
    &Reminders{},
    Roll{},
    RPS{},
    Shrug{},
    Stats{},
    Stone{},
    Support{},
    TableFlip{},
    Triggered{},
    Uptime{},
    UrbanDict{},
    Weather{},
    XKCD{},
}

// TriggerPluginList is the list of plugins that activate on normal chat
var TriggerPluginList = []TriggerPlugin{
    triggers.CSS{},
    triggers.Donate{},
    triggers.EightBall{},
    triggers.Hi{},
    triggers.HypeTrain{},
    triggers.Invite{},
    triggers.IPTables{},
    triggers.Nep{},
    triggers.ReZero{},
}

// CallBotPlugin iterates through the list of registered
// plugins and tries to guess which one is the intended call
// Fist match wins.
//
// command - The command that triggered this execution
// content - The content without command
// msg     - The message object
// session - The discord session
func CallBotPlugin(command string, content string, msg *discordgo.Message) {
    // Iterate over all plugins
    for _, plug := range PluginList {
        // Iterate over all commands of the current plugin
        for _, cmd := range plug.Commands() {
            if command == cmd {
                go safePluginCall(command, strings.TrimSpace(content), msg, plug)
                break
            }
        }
    }
}

// CallTriggerPlugins iterates through all trigger plugins
// and calls *all* of them (async).
//
// msg     - The message that triggered the execution
// session - The discord session
func CallTriggerPlugin(trigger string, content string, msg *discordgo.Message) {
    // Iterate over all plugins
    for _, plug := range TriggerPluginList {
        for _, trig := range plug.Triggers() {
            if trigger == trig {
                go func(plugin TriggerPlugin) {
                    defer helpers.RecoverDiscord(msg)
                    cache.GetSession().ChannelMessageSend(
                        msg.ChannelID,
                        plugin.Response(trigger, content),
                    )
                }(plug)
                break
            }
        }
    }
}

// Wrapper that catches any panics from plugins
// Arguments: Same as CallBotPlugin().
func safePluginCall(command string, content string, msg *discordgo.Message, plug Plugin) {
    defer helpers.RecoverDiscord(msg)
    metrics.CommandsExecuted.Add(1)
    plug.Action(command, content, msg, cache.GetSession())
}

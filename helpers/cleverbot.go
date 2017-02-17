package helpers

import (
    "github.com/bwmarrin/discordgo"
    "github.com/CleverbotIO/go-cleverbot.io"
)

const API_ID = "Karen Discord-Bot <lukas.breuer@outlook.com> (https://meetkaren.xyz) | Session "

// cleverbotSessions stores all cleverbot connections
var cleverbotSessions map[string]*cleverbot.Session

// CleverbotSend sends a message to cleverbot and responds with it's answer.
func CleverbotSend(session *discordgo.Session, channel string, message string) {
    var msg string

    if cleverbotSessions[channel] == nil {
        if len(cleverbotSessions) == 0 {
            cleverbotSessions = make(map[string]*cleverbot.Session)
        }

        CleverbotRefreshSession(channel)
    }

    response, err := cleverbotSessions[channel].Ask(message)
    if err != nil {
        msg = "Error :frowning:\n```\n" + err.Error() + "\n```"
    } else {
        msg = response
    }

    session.ChannelMessageSend(channel, msg)
}

// CleverbotRefreshSession refreshes the cleverbot session for said channel
func CleverbotRefreshSession(channel string) {
    session, err := cleverbot.New(
        GetConfig().Path("cleverbot.user").Data().(string),
        GetConfig().Path("cleverbot.key").Data().(string),
        API_ID + channel,
    )
    Relax(err)

    cleverbotSessions[channel] = session
}

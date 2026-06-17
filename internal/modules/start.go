/*
 * ○ Anny X Music - A high-performance engine for streaming music in Telegram voicechats.
 *
 * Copyright (C) 2026 @Mad_x_Avi
 * All rights reserved.
 *
 * This source code is proprietary and confidential.
 * Unauthorized copying, distribution, or modification is strictly prohibited.
 */

package modules

import (
	"time"

	"github.com/Laky-64/gologging"
	tg "github.com/amarnathcjd/gogram/telegram"

	"main/internal/config"
	"main/internal/core"
	"main/internal/database"
	"main/internal/locales"
	"main/internal/utils"
)

func init() {
	helpTexts["/start"] = `<i>Start Anny X Music and show main menu.</i>`
}

func startHandler(m *tg.NewMessage) error {
	if m.ChatType() != tg.EntityUser {
		database.AddServedChat(m.ChannelID())
		m.Reply(
			F(m.ChannelID(), "start_group"),
		)
		return tg.ErrEndGroup
	}

	arg := m.Args()
	database.AddServedUser(m.ChannelID())

	if arg != "" {
		gologging.Info(
			"Got Start parameter: " + arg + " in ChatID: " + utils.IntToStr(
				m.ChannelID(),
			),
		)
	}

	switch arg {
	case "pm_help":
		gologging.Info("User requested help via start param")
		helpHandler(m)

	default:
		// Send heart reaction to the /start message
		_, err := m.React(tg.ReactionEmoji{
			Emoticon: "❤️",
		})
		if err != nil {
			gologging.Error("[start] Failed to send heart reaction: " + err.Error())
		}

		// Send animated welcome sequence
		animations := []string{
			"✨ <b>Welcome to</b> ✨",
			"🎵 <b>Anny X Music</b> 🎵",
			"🚀 <b>The Ultimate Streaming Bot</b> 🚀",
			"🎧 <b>Powered by @Mad_x_Avi</b> 🎧",
		}

		for i, animText := range animations {
			var msg *tg.NewMessage
			if i == 0 {
				msg, err = m.Reply(animText)
			} else {
				msg, err = m.Client.EditMessage(m.ChatID(), msg.ID, animText)
			}
			if err != nil {
				gologging.Error("[start] Animation edit failed: " + err.Error())
			} else {
				time.Sleep(800 * time.Millisecond)
			}
		}

		// Delete the animation messages
		m.Client.DeleteMessages(m.ChatID(), []int32{msg.ID})

		// Send sticker with auto-deletion after 3 seconds
		stickerID := "CAACAgEAAxkBAAERZ0NqMlGDnVBT_h1vm1qbL3Fe8_qjigACVAYAAlmWeUd1rCk8DBvZdjwE"
		stickerMsg, err := m.Client.SendMessage(m.ChatID(), &tg.MessageMedia{
			Document: &tg.Document{
				ID: stickerID,
			},
		})
		if err != nil {
			gologging.Error("[start] Failed to send sticker: " + err.Error())
		} else {
			// Delete sticker after 3 seconds
			go func() {
				time.Sleep(3 * time.Second)
				m.Client.DeleteMessages(m.ChatID(), []int32{stickerMsg.ID})
			}()
		}

		// Brief pause before sending main menu
		time.Sleep(500 * time.Millisecond)

		caption := F(m.ChannelID(), "start_private", locales.Arg{
			"user": utils.MentionHTML(m.Sender),
			"bot":  utils.MentionHTML(m.Client.Me()),
		})

		_, err = m.RespondMedia(&tg.InputMediaWebPage{
			URL:             config.StartImage,
			ForceLargeMedia: true,
		}, &tg.MediaOptions{
			Caption:     caption,
			NoForwards:  true,
			ReplyMarkup: core.GetStartMarkup(m.ChannelID()),
		})
		if err != nil {
			gologging.Error(
				"[start] InputMediaWebPage Reply failed: " + err.Error(),
			)

			_, err = m.RespondMedia(config.StartImage, &tg.MediaOptions{
				Caption:     caption,
				NoForwards:  true,
				ReplyMarkup: core.GetStartMarkup(m.ChannelID()),
			})
			if err != nil {
				gologging.Error(
					"[start] URL media reply failed: " + err.Error(),
				)

				_, err = m.Respond(caption, &tg.SendOptions{
					NoForwards:  true,
					ReplyMarkup: core.GetStartMarkup(m.ChannelID()),
				})
				return err
			}
		}
	}

	if config.LoggerID != 0 && isLoggerEnabled() {
		uName := "N/A"
		if m.Sender.Username != "" {
			uName = "@" + m.Sender.Username
		}
		msg := F(m.ChannelID(), "logger_bot_started", locales.Arg{
			"mention":       utils.MentionHTML(m.Sender),
			"user_id":       m.SenderID(),
			"user_username": uName,
		})
		_, err := m.Client.SendMessage(config.LoggerID, msg)
		if err != nil {
			gologging.Error(
				"Failed to send logger_bot_started msg, Err: " + err.Error(),
			)
		}
	}
	return tg.ErrEndGroup
}

func startCB(cb *tg.CallbackQuery) error {
	cb.Answer("")

	caption := F(cb.ChannelID(), "start_private", locales.Arg{
		"user": utils.MentionHTML(cb.Sender),
		"bot":  utils.MentionHTML(cb.Client.Me()),
	})

	sendOpt := &tg.SendOptions{
		ReplyMarkup: core.GetStartMarkup(cb.ChannelID()),
		NoForwards:  true,
	}

	if config.StartImage != "" {
		sendOpt.Media = &tg.InputMediaWebPage{
			URL:             config.StartImage,
			ForceLargeMedia: true,
		}
	}

	cb.Edit(caption, sendOpt)
	return tg.ErrEndGroup
}

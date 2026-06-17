/*
 * ○ Anny X Music - A high-performance engine for streaming music in Telegram voicechats.
 *
 * Copyright (C) 2026 @Mad_x_Avi
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
		// ========== NEW: Heart Reaction ==========
		_, err := m.React(tg.ReactionEmoji{
			Emoticon: "❤️",
		})
		if err != nil {
			gologging.Error("[start] Failed to send heart reaction: " + err.Error())
		}

		// ========== NEW: Animated Welcome Text Sequence ==========
		animations := []string{
			"✨ <b>Welcome to the Best Streaming Bot!</b> ✨",
			"🎵 <b>Anny X Music</b> 🎵",
			"🚀 <b>High-Quality Music • 24/7 • Ultra Fast</b> 🚀",
			"🎧 <b>Powered by @Mad_x_Avi</b> 🎧",
		}

		var animMsg *tg.NewMessage
		for i, animText := range animations {
			if i == 0 {
				animMsg, err = m.Reply(animText)
			} else {
				animMsg, err = m.Client.EditMessage(m.ChatID(), animMsg.ID, animText)
			}
			if err != nil {
				gologging.Error("[start] Animation sequence failed: " + err.Error())
			} else {
				time.Sleep(800 * time.Millisecond)
			}
		}

		// Delete the animation message after sequence completes
		if animMsg != nil {
			m.Client.DeleteMessages(m.ChatID(), []int32{animMsg.ID})
		}

		// ========== NEW: Sticker with 3s auto-delete ==========
		stickerID := "CAACAgEAAxkBAAERZ0NqMlGDnVBT_h1vm1qbL3Fe8_qjigACVAYAAlmWeUd1rCk8DBvZdjwE"
		stickerMsg, stickerErr := m.Client.SendMessage(m.ChatID(), &tg.MessageMedia{
			Document: &tg.Document{
				ID: stickerID,
			},
		})
		if stickerErr != nil {
			gologging.Error("[start] Failed to send sticker: " + stickerErr.Error())
		} else {
			// Auto-delete sticker after 3 seconds
			go func() {
				time.Sleep(3 * time.Second)
				m.Client.DeleteMessages(m.ChatID(), []int32{stickerMsg.ID})
			}()
		}

		// Small delay before showing main menu
		time.Sleep(500 * time.Millisecond)

		// ========== ORIGINAL CODE: Main welcome menu ==========
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

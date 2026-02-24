// Agent Framework - Channel Adapters Registration
// Copyright (C) 2025 Agent Framework Contributors
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package adapters

import (
	"AgentFramework/pkg/channels"
)

func init() {
	// Register all built-in adapters
	// This allows them to be used without creating circular import dependencies

	// Discord adapter - fully functional
	channels.RegisterBuiltinAdapter(channels.ChannelTypeDiscord, func(id string) channels.ChannelAdapter {
		return NewDiscordAdapter(id)
	})

	// Telegram adapter - temporarily disabled due to telebot.v3 API incompatibility
	// To re-enable, update adapter for new telebot.v3 API or downgrade: go get gopkg.in/telebot.v3@v2
	// channels.RegisterBuiltinAdapter(channels.ChannelTypeTelegram, func(id string) channels.ChannelAdapter {
	//	return NewTelegramAdapter(id)
	// })

	// Slack adapter temporarily disabled due to slack-go v0.12+ API incompatibility
	// channels.RegisterBuiltinAdapter(channels.ChannelTypeSlack, func(id string) channels.ChannelAdapter {
	//	return NewSlackAdapter(id)
	// }

	// Feishu adapter - fully functional
	channels.RegisterBuiltinAdapter(channels.ChannelTypeFeishu, func(id string) channels.ChannelAdapter {
		return NewFeishuAdapter(id)
	})

	// WeWork adapter - temporarily disabled due to API issues
	// channels.RegisterBuiltinAdapter(channels.ChannelTypeWeWork, func(id string) channels.ChannelAdapter {
	//	return NewWeWorkAdapter(id)
	// })

	// DingTalk adapter - fully functional
	channels.RegisterBuiltinAdapter(channels.ChannelTypeDingTalk, func(id string) channels.ChannelAdapter {
		return NewDingTalkAdapter(id)
	})

	// QQ adapter - fully functional
	channels.RegisterBuiltinAdapter(channels.ChannelTypeQQ, func(id string) channels.ChannelAdapter {
		return NewQQAdapter(id)
	})
}

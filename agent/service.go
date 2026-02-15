// Agent Framework - A high-performance, enterprise-grade AI Agent framework in Go
// Copyright (C) 2025  Agent Framework Contributors

// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Affero General Public License for more details.

// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>.

// Additional permission under GNU Affero General Public License version 3 section 7
// If you modify this Program, or any covered work, by linking or combining it
// with other code, such other code is not for that reason alone subject to any
// of the requirements of the GNU Affero GPL version 3 as long as you maintain
// the separation between the Program and the other code.

// For network interaction purposes, when this Program is used over a network,
// the source code of the Program must be made available to users of the network.
// You can comply with this requirement by providing a link to the source code
// repository in your user interface or documentation.

// SPDX-License-Identifier: AGPL-3.0-or-later

package agent

import (
	"context"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
)

type ThreadAwareAgent interface {
	Agent
	UseThread(thread *Thread)
}

type AgentService struct {
	store ThreadStore
}

func NewAgentService(store ThreadStore) *AgentService {
	return &AgentService{
		store: store,
	}
}

func (s *AgentService) NewThread(ctx context.Context) (*Thread, error) {
	return s.store.Create(ctx)
}

func (s *AgentService) Send(ctx context.Context, ag ThreadAwareAgent, threadID string, input string, opts ...model.Option) (*Thread, *schema.Message, error) {
	var (
		thread *Thread
		err    error
	)

	if threadID == "" {
		thread, err = s.store.Create(ctx)
		if err != nil {
			return nil, nil, err
		}
	} else {
		thread, err = s.store.Get(ctx, threadID)
		if err != nil {
			return nil, nil, err
		}
		if thread == nil {
			thread = &Thread{ID: threadID}
		}
	}

	ag.UseThread(thread)

	resp, err := ag.Run(ctx, input, opts...)
	if err != nil {
		return nil, nil, err
	}

	err = s.store.Save(ctx, thread)
	if err != nil {
		return nil, nil, err
	}

	return thread, resp, nil
}

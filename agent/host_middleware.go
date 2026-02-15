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
	"errors"
	"log"
	"regexp"
	"strings"
	"time"
)

// registerDefaultMiddlewares registers built-in middlewares
func (h *Host) registerDefaultMiddlewares() {
	h.middlewares["logging"] = NewLoggingMiddleware(func(name, input string, duration time.Duration, err error) {
		log.Printf("[Agent:%s] Input=%q Duration=%v Err=%v", name, input, duration, err)
	})
	// Deprecated custom telemetry
	// h.middlewares["telemetry"] = NewTelemetryMiddleware(nil)

	// New OTel Telemetry
	h.middlewares["otel"] = OTelMiddleware()

	h.middlewares["safe_output"] = NewOutputFilterMiddleware(defaultSecurityFilter)
	h.middlewares["policy_safe"] = NewInputPolicyMiddleware(defaultInputPolicy)
}

// Default security filter patterns
var (
	emailRegexp = regexp.MustCompile(`[A-Za-z0-9._%+\-]+@[A-Za-z0-9.\-]+\.[A-Za-z]{2,}`)
	phoneRegexp = regexp.MustCompile(`\d{3}[\-\s]?\d{4}[\-\s]?\d{4}`)
)

// defaultSecurityFilter filters sensitive information from output
func defaultSecurityFilter(output string) (string, error) {
	if output == "" {
		return output, nil
	}

	sanitized := emailRegexp.ReplaceAllString(output, "[redacted-email]")
	sanitized = phoneRegexp.ReplaceAllString(sanitized, "[redacted-phone]")

	return sanitized, nil
}

// defaultInputPolicy validates input against security policies
func defaultInputPolicy(ctx context.Context, input string) error {
	_ = ctx
	if input == "" {
		return nil
	}

	lower := strings.ToLower(input)

	if strings.Contains(lower, "drop table") ||
		strings.Contains(lower, "truncate table") ||
		strings.Contains(lower, "rm -rf") {
		return errors.New("input rejected by policy")
	}

	if strings.Contains(input, "删除数据库") {
		return errors.New("input rejected by policy")
	}

	return nil
}

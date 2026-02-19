// Copyright (C) 2025 Agent Framework Contributors
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// SPDX-License-Identifier: AGPL-3.0-or-later


package telemetry

import (
	"context"
)

// TracePlaceholder is a minimal wrapper to return a no-op trace span.
func TraceFromContext(ctx context.Context) context.Context {
	return ctx
}

func EndTrace(ctx context.Context) {}

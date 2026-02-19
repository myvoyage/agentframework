# Migration Guide: agent/errors to pkg/errors

## Overview

The `agent/errors` package has been merged into `pkg/errors` to eliminate code duplication. All code should use `pkg/errors` going forward.

## Migration Steps

1. **Update imports**: Replace `AgentFramework/agent/errors` with `AgentFramework/pkg/errors`
2. **Error codes**: All error codes are now in `pkg/errors` package
3. **Additional error codes**: The following error codes have been added to `pkg/errors`:
   - `ErrCodeExecutionFailed`
   - `ErrCodeInitFailed`
   - `ErrCodeShutdownFailed`
   - `ErrCodeDownloadFailed`
   - `ErrCodeInstallFailed`
   - `ErrCodeValidationFailed`

## Example

**Before:**
```go
import "AgentFramework/agent/errors"

return errors.New(errors.ErrCodeNotFound, "resource not found")
```

**After:**
```go
import "AgentFramework/pkg/errors"

return errors.New(errors.ErrCodeNotFound, "resource not found")
```

## Files to Update

The following files need to be updated:
- `pkg/framework/workflow/workflow_manager.go` ✅ (already updated)
- Any other files importing `agent/errors` (use grep to find)

## Deprecation

The `agent/errors` package will be removed in a future version. Please migrate all code to use `pkg/errors`.

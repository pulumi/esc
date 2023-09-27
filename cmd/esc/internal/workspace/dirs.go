// Copyright 2023, Pulumi Corporation.  All rights reserved.

package workspace

import (
	"github.com/pulumi/pulumi/sdk/v3/go/common/workspace"
)

const BookkeepingDir = ".esc"

// GetBookkeepingDir returns the path to the esc CLI's bookkeeping directory.
func GetBookkeepingDir() (string, error) {
	return workspace.GetPulumiPath(BookkeepingDir)
}

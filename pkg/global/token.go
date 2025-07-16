// Copyright The Linux Foundation and each contributor to LFX.
// SPDX-License-Identifier: MIT

package global

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"sync"
)

var (
	pageTokenSecret       [32]byte
	doOncePageTokenSecret sync.Once
)

// PageTokenSecret retrieves the secret used for encoding and decoding page tokens.
func PageTokenSecret(ctx context.Context) *[32]byte {

	doOncePageTokenSecret.Do(func() {

		const pageTokenSecretName = "PAGE_TOKEN_SECRET"

		pageTokenSecretValue := os.Getenv(pageTokenSecretName)
		if pageTokenSecretValue == "" {
			slog.ErrorContext(ctx, fmt.Sprintf("%s environment variable is not set", pageTokenSecretName))
			os.Exit(1)
		}
		copy(pageTokenSecret[:], []byte(pageTokenSecretValue))
	})

	return &pageTokenSecret
}

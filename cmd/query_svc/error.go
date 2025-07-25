// Copyright The Linux Foundation and each contributor to LFX.
// SPDX-License-Identifier: MIT

package querysvcapi

import (
	"context"
	"log/slog"

	querysvc "github.com/linuxfoundation/lfx-v2-query-service/gen/query_svc"
	"github.com/linuxfoundation/lfx-v2-query-service/pkg/errors"
)

func wrapError(ctx context.Context, err error) error {

	f := func(err error) error {
		switch e := err.(type) {
		case errors.Validation:
			return &querysvc.BadRequestError{
				Message: e.Error(),
			}
		case errors.ServiceUnavailable:
			return &querysvc.ServiceUnavailableError{
				Message: e.Error(),
			}
		default:
			return &querysvc.InternalServerError{
				Message: e.Error(),
			}
		}
	}

	slog.ErrorContext(ctx, "request failed",
		"error", err,
	)
	return f(err)
}

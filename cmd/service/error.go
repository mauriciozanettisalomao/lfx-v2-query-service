// Copyright The Linux Foundation and each contributor to LFX.
// SPDX-License-Identifier: MIT

package service

import (
	"context"
	"log/slog"

	querysvc "github.com/linuxfoundation/lfx-v2-query-service/gen/query_svc"
	"github.com/linuxfoundation/lfx-v2-query-service/pkg/errors"
)

func wrapError(ctx context.Context, err error) error {

	f := func(err error) error {
		if err == nil {
			return &querysvc.InternalServerError{
				Message: "unknown error",
			}
		}

		switch e := err.(type) {
		case errors.Validation:
			return &querysvc.BadRequestError{
				Message: e.Error(),
			}
		case errors.NotFound:
			return &querysvc.NotFoundError{
				Message: e.Error(),
			}
		case errors.ServiceUnavailable:
			return &querysvc.ServiceUnavailableError{
				Message: e.Error(),
			}
		default:
			return &querysvc.InternalServerError{
				Message: err.Error(),
			}
		}
	}

	slog.ErrorContext(ctx, "request failed",
		"error", err,
	)
	return f(err)
}

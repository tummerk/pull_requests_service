package modules

import "pull_requests_service/pkg/contextx"

var logger = contextx.LoggerFromContextOrDefault //nolint:gochecknoglobals

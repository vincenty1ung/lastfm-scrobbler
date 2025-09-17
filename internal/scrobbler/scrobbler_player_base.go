package scrobbler

import (
	"github.com/vincenty1ung/lastfm-scrobbler/common"
)

type BaseWrapper struct {
}

func (m BaseWrapper) ConversionSimplified(target string) string {
	return common.ConversionSimplifiedFx(target)
}

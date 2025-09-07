package common

import (
	"context"
	"errors"
	"strings"

	"github.com/mitchellh/mapstructure"
)

func Decode(input interface{}, output interface{}) error {
	decoder, err := mapstructure.NewDecoder(
		&mapstructure.DecoderConfig{
			ZeroFields: true,
			TagName:    "json",
			Result:     output,
		},
	)
	if err != nil {
		return err
	}

	return decoder.Decode(input)
}

// ValidateTrackInfo validates the artist, album, and track names
// Returns an error if any of them are empty or contain only whitespace
func ValidateTrackInfo(ctx context.Context, artist, album, track string) error {
	// Trim whitespace from all fields
	artist = strings.TrimSpace(artist)
	album = strings.TrimSpace(album)
	track = strings.TrimSpace(track)

	// Check if any field is empty after trimming
	if artist == "" {
		return errors.New("艺术家名称不能为空")
	}
	if album == "" {
		return errors.New("专辑名称不能为空")
	}
	if track == "" {
		return errors.New("歌曲名称不能为空")
	}

	return nil
}

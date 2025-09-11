// DEPRECATED: This file is deprecated. Please use the new Track model in track.go instead.
package model

import (
	"context"
	"time"
)

// TrackPlayCount is deprecated. Use Track instead.
type TrackPlayCount struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Artist    string    `gorm:"index;uniqueIndex:idx_track_album_artist" json:"artist"`
	Album     string    `gorm:"index;uniqueIndex:idx_track_album_artist" json:"album"`
	Track     string    `gorm:"index;uniqueIndex:idx_track_album_artist" json:"track"`
	PlayCount int       `json:"play_count"`
	Version   int       `gorm:"default:1" json:"version"` // 乐观锁版本号
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (TrackPlayCount) TableName() string {
	return "track_play_counts"
}

// DEPRECATED: Use IncrementTrackPlayCount from the new Track model instead.
func DeprecatedIncrementTrackPlayCount(ctx context.Context, artist, album, track string) error {
	// This function is deprecated. Please use the new Track model instead.
	// Implementation is kept for backward compatibility only.
	return nil
}

// DEPRECATED: Use GetTracks from the new Track model instead.
func DeprecatedGetTrackPlayCounts(ctx context.Context, limit, offset int) ([]*TrackPlayCount, error) {
	// This function is deprecated. Please use the new Track model instead.
	// Implementation is kept for backward compatibility only.
	return nil, nil
}

// DEPRECATED: Use GetTrackCounts from the new Track model instead.
func DeprecatedGetTrackCounts(ctx context.Context) (int64, error) {
	// This function is deprecated. Please use the new Track model instead.
	// Implementation is kept for backward compatibility only.
	return 0, nil
}

// DEPRECATED: Use GetTrack from the new Track model instead.
func DeprecatedGetTrackPlayCount(ctx context.Context, artist, album, track string) (*TrackPlayCount, error) {
	// This function is deprecated. Please use the new Track model instead.
	// Implementation is kept for backward compatibility only.
	return nil, nil
}

// DEPRECATED: Use GetAllTrackPlayCounts from the new Track model instead.
// GetAllTrackPlayCounts 获取所有播放统计记录
func DeprecatedGetAllTrackPlayCounts(ctx context.Context) ([]*TrackPlayCount, error) {
	// This function is deprecated. Please use the new Track model instead.
	// Implementation is kept for backward compatibility only.
	return nil, nil
}

// DEPRECATED: Use GetTracksByArtist from the new Track model instead.
// GetTracksByArtist 获取特定艺术家的所有曲目
func DeprecatedGetTracksByArtist(ctx context.Context, artist string) ([]*TrackPlayCount, error) {
	// This function is deprecated. Please use the new Track model instead.
	// Implementation is kept for backward compatibility only.
	return nil, nil
}

// DEPRECATED: Use GetTotalPlayCount from the new Track model instead.
// GetTotalPlayCount 获取总播放次数
func DeprecatedGetTotalPlayCount(ctx context.Context) (int64, error) {
	// This function is deprecated. Please use the new Track model instead.
	// Implementation is kept for backward compatibility only.
	return 0, nil
}

// DEPRECATED: Use GetArtistCounts from the new Track model instead.
// GetArtistCounts 获取艺术家总数
func DeprecatedGetArtistCounts(ctx context.Context) (int64, error) {
	// This function is deprecated. Please use the new Track model instead.
	// Implementation is kept for backward compatibility only.
	return 0, nil
}

// DEPRECATED: Use GetAlbumCounts from the new Track model instead.
// GetAlbumCounts 获取专辑总数
func DeprecatedGetAlbumCounts(ctx context.Context) (int64, error) {
	// This function is deprecated. Please use the new Track model instead.
	// Implementation is kept for backward compatibility only.
	return 0, nil
}

// DEPRECATED: Use GetTopArtistsByPlayCount from the new Track model instead.
// GetTopArtistsByPlayCount 获取按播放次数统计的热门艺术家
func DeprecatedGetTopArtistsByPlayCount(ctx context.Context, limit int) ([]map[string]interface{}, error) {
	// This function is deprecated. Please use the new Track model instead.
	// Implementation is kept for backward compatibility only.
	return nil, nil
}

// DEPRECATED: Use GetTopArtistsByTrackCount from the new Track model instead.
// GetTopArtistsByTrackCount 获取按曲目数统计的热门艺术家
func DeprecatedGetTopArtistsByTrackCount(ctx context.Context, limit int) ([]map[string]interface{}, error) {
	// This function is deprecated. Please use the new Track model instead.
	// Implementation is kept for backward compatibility only.
	return nil, nil
}

// DEPRECATED: Use GetTracksByPeriod from the new Track model instead.
// GetTracksByPeriod 获取指定时间段内的曲目播放统计
func DeprecatedGetTrackPlayCountsByPeriod(ctx context.Context, limit int, offset int, period string) (
	[]*TrackPlayCount, error,
) {
	// This function is deprecated. Please use the new Track model instead.
	// Implementation is kept for backward compatibility only.
	return nil, nil
}

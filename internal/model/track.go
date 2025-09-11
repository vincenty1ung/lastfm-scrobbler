package model

import (
	"context"
	"errors"
	"sort"
	"time"

	"gorm.io/gorm"

	"github.com/vincenty1ung/lastfm-scrobbler/common"
)

// Track represents a music track with play statistics and favorite status
type Track struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	Artist          string    `gorm:"index;uniqueIndex:idx_artist_album_track" json:"artist"`
	Album           string    `gorm:"index;uniqueIndex:idx_artist_album_track" json:"album"`
	Track           string    `gorm:"index;uniqueIndex:idx_artist_album_track" json:"track"`
	PlayCount       int       `json:"play_count"`
	IsAppleMusicFav bool      `json:"is_apple_music_fav"`       // 是否Apple Music喜欢
	IsLastFmFav     bool      `json:"is_lastfm_fav"`            // 是否Last.fm喜欢
	Version         int       `gorm:"default:1" json:"version"` // 乐观锁版本号
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// TableName sets the table name for the Track model
func (Track) TableName() string {
	return "track"
}

// IncrementTrackPlayCount increments the play count for a track
func IncrementTrackPlayCount(ctx context.Context, artist, album, track string) error {
	// 验证艺术家、专辑和曲目信息
	if err := common.ValidateTrackInfo(ctx, artist, album, track); err != nil {
		return err
	}

	// 使用乐观锁机制更新播放次数
	for {
		var record Track
		err := GetDB().WithContext(ctx).Where(
			"artist = ? AND album = ? AND track = ?", artist, album, track,
		).First(&record).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				// Create new record
				record = Track{
					Artist:    artist,
					Album:     album,
					Track:     track,
					PlayCount: 1,
				}
				err = GetDB().WithContext(ctx).Create(&record).Error
				if err != nil && !errors.Is(err, gorm.ErrDuplicatedKey) {
					return err
				}
				// 如果出现重复键错误，说明其他goroutine已经创建了记录，继续循环处理
				if errors.Is(err, gorm.ErrDuplicatedKey) {
					continue
				}
				return nil
			}
			return err
		}

		// Update existing record with optimistic locking
		updatedRecord := Track{
			PlayCount: record.PlayCount + 1,
			Version:   record.Version + 1,
		}

		result := GetDB().WithContext(ctx).Where(
			"artist = ? AND album = ? AND track = ? AND version = ?",
			artist, album, track, record.Version,
		).Updates(&updatedRecord)

		if result.Error != nil {
			return result.Error
		}

		// 如果更新成功，跳出循环
		if result.RowsAffected > 0 {
			break
		}
		// 如果更新失败（版本号不匹配），继续循环重试
	}

	return nil
}

// GetTracks retrieves track play counts with pagination
func GetTracks(ctx context.Context, limit, offset int) ([]*Track, error) {
	var records []*Track
	err := GetDB().WithContext(ctx).Order("play_count DESC").Limit(limit).Offset(offset).Find(&records).Error
	if err != nil {
		return nil, err
	}
	return records, nil
}

// GetTrackCounts returns the total number of tracks
func GetTrackCounts(ctx context.Context) (int64, error) {
	var count int64
	err := GetDB().WithContext(ctx).Model(&Track{}).Count(&count).Error
	if err != nil {
		return 0, err
	}
	return count, nil
}

// GetTrack retrieves a specific track's play count
func GetTrack(ctx context.Context, artist, album, track string) (*Track, error) {
	var record Track
	err := GetDB().WithContext(ctx).Where(
		"artist = ? AND album = ? AND track = ?", artist, album, track,
	).First(&record).Error
	if err != nil {
		return nil, err
	}
	return &record, nil
}

// GetAllTrackPlayCounts retrieves all track play counts
func GetAllTrackPlayCounts(ctx context.Context) ([]*Track, error) {
	var allTracks []*Track
	pageSize := 100
	offset := 0

	for {
		var tracks []*Track
		err := GetDB().WithContext(ctx).Order("play_count DESC").Limit(pageSize).Offset(offset).Find(&tracks).Error
		if err != nil {
			return nil, err
		}

		allTracks = append(allTracks, tracks...)

		// 如果返回的记录数少于pageSize，说明已经获取完所有记录
		if len(tracks) < pageSize {
			break
		}

		offset += pageSize
	}

	return allTracks, nil
}

// GetTracksByArtist retrieves all tracks by a specific artist
func GetTracksByArtist(ctx context.Context, artist string) ([]*Track, error) {
	var tracks []*Track
	err := GetDB().WithContext(ctx).Where("artist LIKE ?", "%"+artist+"%").Find(&tracks).Error
	if err != nil {
		return nil, err
	}
	return tracks, nil
}

// GetTotalPlayCount returns the total play count across all tracks
func GetTotalPlayCount(ctx context.Context) (int64, error) {
	var total int64
	err := GetDB().WithContext(ctx).Model(&Track{}).Select("SUM(play_count)").Scan(&total).Error
	if err != nil {
		return 0, err
	}
	return total, nil
}

// GetArtistCounts returns the total number of unique artists
func GetArtistCounts(ctx context.Context) (int64, error) {
	var count int64
	err := GetDB().WithContext(ctx).Model(&Track{}).Distinct("artist").Count(&count).Error
	if err != nil {
		return 0, err
	}
	return count, nil
}

// GetAlbumCounts returns the total number of unique albums
func GetAlbumCounts(ctx context.Context) (int64, error) {
	var count int64
	err := GetDB().WithContext(ctx).Model(&Track{}).Distinct("album").Count(&count).Error
	if err != nil {
		return 0, err
	}
	return count, nil
}

// GetTopArtistsByPlayCount returns the top artists by play count
func GetTopArtistsByPlayCount(ctx context.Context, limit int) ([]map[string]interface{}, error) {
	var result []map[string]interface{}
	err := GetDB().WithContext(ctx).Model(&Track{}).
		Select("artist, SUM(play_count) as play_count").
		Group("artist").
		Order("SUM(play_count) DESC").
		Limit(limit).
		Find(&result).Error
	if err != nil {
		return nil, err
	}
	return result, nil
}

// GetTopArtistsByTrackCount returns the top artists by track count
func GetTopArtistsByTrackCount(ctx context.Context, limit int) ([]map[string]interface{}, error) {
	var result []map[string]interface{}
	err := GetDB().WithContext(ctx).Model(&Track{}).
		Select("artist, COUNT(*) as track_count").
		Group("artist").
		Order("COUNT(*) DESC").
		Limit(limit).
		Find(&result).Error
	if err != nil {
		return nil, err
	}
	return result, nil
}

// GetTracksByPeriod retrieves track play counts for a specific period
func GetTracksByPeriod(ctx context.Context, limit int, offset int, period string) ([]*Track, error) {
	// 计算时间范围
	var startTime time.Time
	switch period {
	case "week":
		startTime = time.Now().AddDate(0, 0, -7)
	case "month":
		startTime = time.Now().AddDate(0, -1, 0)
	default:
		// 默认返回所有时间的数据
		return GetTracks(ctx, limit, offset)
	}

	// 先获取指定时间范围内的播放记录
	var playRecords []*TrackPlayRecord
	err := GetDB().WithContext(ctx).Where(
		"play_time >= ?", startTime,
	).Order("").Find(&playRecords).Error
	if err != nil {
		return nil, err
	}

	// 统计每个曲目的播放次数
	trackCountMap := make(map[string]*Track)
	for _, record := range playRecords {
		key := record.Artist + "|" + record.Album + "|" + record.Track
		if trackCount, exists := trackCountMap[key]; exists {
			trackCount.PlayCount++
		} else {
			trackCountMap[key] = &Track{
				Artist:    record.Artist,
				Album:     record.Album,
				Track:     record.Track,
				PlayCount: 1,
			}
		}
	}

	// 转换为切片并排序
	var trackCounts []*Track
	for _, trackCount := range trackCountMap {
		trackCounts = append(trackCounts, trackCount)
	}

	// 按播放次数排序
	sort.Slice(
		trackCounts, func(i, j int) bool {
			return trackCounts[i].PlayCount > trackCounts[j].PlayCount
		},
	)

	// 应用分页
	start := offset
	end := offset + limit
	if start >= len(trackCounts) {
		return []*Track{}, nil
	}
	if end > len(trackCounts) {
		end = len(trackCounts)
	}

	return trackCounts[start:end], nil
}

// SetAppleMusicFavorite updates the Apple Music favorite status for a track
func SetAppleMusicFavorite(ctx context.Context, artist, album, track string, isFavorite bool) error {
	// 验证艺术家、专辑和曲目信息
	if err := common.ValidateTrackInfo(ctx, artist, album, track); err != nil {
		return err
	}

	// 使用乐观锁机制更新喜欢状态
	for {
		var record Track
		err := GetDB().WithContext(ctx).Where(
			"artist = ? AND album = ? AND track = ?", artist, album, track,
		).First(&record).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				// Create new record
				record = Track{
					Artist:          artist,
					Album:           album,
					Track:           track,
					PlayCount:       0,
					IsAppleMusicFav: isFavorite,
				}
				err = GetDB().WithContext(ctx).Create(&record).Error
				if err != nil && !errors.Is(err, gorm.ErrDuplicatedKey) {
					return err
				}
				// 如果出现重复键错误，说明其他goroutine已经创建了记录，继续循环处理
				if errors.Is(err, gorm.ErrDuplicatedKey) {
					continue
				}
				return nil
			}
			return err
		}

		// Update existing record with optimistic locking
		updatedRecord := Track{
			IsAppleMusicFav: isFavorite,
			Version:         record.Version + 1,
		}

		result := GetDB().WithContext(ctx).Where(
			"artist = ? AND album = ? AND track = ? AND version = ?",
			artist, album, track, record.Version,
		).Updates(&updatedRecord)

		if result.Error != nil {
			return result.Error
		}

		// 如果更新成功，跳出循环
		if result.RowsAffected > 0 {
			break
		}
		// 如果更新失败（版本号不匹配），继续循环重试
	}

	return nil
}

// SetLastFmFavorite updates the Last.fm favorite status for a track
func SetLastFmFavorite(ctx context.Context, artist, album, track string, isFavorite bool) error {
	// 验证艺术家、专辑和曲目信息
	if err := common.ValidateTrackInfo(ctx, artist, album, track); err != nil {
		return err
	}

	// 使用乐观锁机制更新喜欢状态
	for {
		var record Track
		err := GetDB().WithContext(ctx).Where(
			"artist = ? AND album = ? AND track = ?", artist, album, track,
		).First(&record).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				// Create new record
				record = Track{
					Artist:      artist,
					Album:       album,
					Track:       track,
					PlayCount:   0,
					IsLastFmFav: isFavorite,
				}
				err = GetDB().WithContext(ctx).Create(&record).Error
				if err != nil && !errors.Is(err, gorm.ErrDuplicatedKey) {
					return err
				}
				// 如果出现重复键错误，说明其他goroutine已经创建了记录，继续循环处理
				if errors.Is(err, gorm.ErrDuplicatedKey) {
					continue
				}
				return nil
			}
			return err
		}

		// Update existing record with optimistic locking
		updatedRecord := Track{
			IsLastFmFav: isFavorite,
			Version:     record.Version + 1,
		}

		result := GetDB().WithContext(ctx).Where(
			"artist = ? AND album = ? AND track = ? AND version = ?",
			artist, album, track, record.Version,
		).Updates(&updatedRecord)

		if result.Error != nil {
			return result.Error
		}

		// 如果更新成功，跳出循环
		if result.RowsAffected > 0 {
			break
		}
		// 如果更新失败（版本号不匹配），继续循环重试
	}

	return nil
}

// GetAppleMusicFavorite retrieves the Apple Music favorite status for a track
func GetAppleMusicFavorite(ctx context.Context, artist, album, track string) (bool, error) {
	record, err := GetTrack(ctx, artist, album, track)
	if err != nil {
		return false, err
	}
	return record.IsAppleMusicFav, nil
}

// GetLastFmFavorite retrieves the Last.fm favorite status for a track
func GetLastFmFavorite(ctx context.Context, artist, album, track string) (bool, error) {
	record, err := GetTrack(ctx, artist, album, track)
	if err != nil {
		return false, err
	}
	return record.IsLastFmFav, nil
}

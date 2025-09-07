package model

import (
	"context"
	"errors"
	"sort"
	"time"

	"github.com/vincenty1ung/lastfm-scrobbler/common"
	"gorm.io/gorm"
)

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

func IncrementTrackPlayCount(ctx context.Context, artist, album, track string) error {
	// 验证艺术家、专辑和曲目信息
	if err := common.ValidateTrackInfo(ctx, artist, album, track); err != nil {
		return err
	}

	// 使用乐观锁机制更新播放次数
	for {
		var record TrackPlayCount
		err := GetDB().WithContext(ctx).Where(
			"artist = ? AND album = ? AND track = ?", artist, album, track,
		).First(&record).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				// Create new record
				record = TrackPlayCount{
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
		updatedRecord := TrackPlayCount{
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

func GetTrackPlayCounts(ctx context.Context, limit, offset int) ([]*TrackPlayCount, error) {
	var records []*TrackPlayCount
	err := GetDB().WithContext(ctx).Order("play_count DESC").Limit(limit).Offset(offset).Find(&records).Error
	if err != nil {
		return nil, err
	}
	return records, nil
}
func GetTrackCounts(ctx context.Context) (int64, error) {
	var count int64
	err := GetDB().WithContext(ctx).Model(&TrackPlayCount{}).Count(&count).Error
	if err != nil {
		return 0, err
	}
	return count, nil
}

func GetTrackPlayCount(ctx context.Context, artist, album, track string) (*TrackPlayCount, error) {
	var record TrackPlayCount
	err := GetDB().WithContext(ctx).Where(
		"artist = ? AND album = ? AND track = ?", artist, album, track,
	).First(&record).Error
	if err != nil {
		return nil, err
	}
	return &record, nil
}

// GetAllTrackPlayCounts 获取所有播放统计记录
func GetAllTrackPlayCounts(ctx context.Context) ([]*TrackPlayCount, error) {
	var allTracks []*TrackPlayCount
	pageSize := 100
	offset := 0

	for {
		var tracks []*TrackPlayCount
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

// GetTracksByArtist 获取特定艺术家的所有曲目
func GetTracksByArtist(ctx context.Context, artist string) ([]*TrackPlayCount, error) {
	var tracks []*TrackPlayCount
	err := GetDB().WithContext(ctx).Where("artist LIKE ?", "%"+artist+"%").Find(&tracks).Error
	if err != nil {
		return nil, err
	}
	return tracks, nil
}

// GetTotalPlayCount 获取总播放次数
func GetTotalPlayCount(ctx context.Context) (int64, error) {
	var total int64
	err := GetDB().WithContext(ctx).Model(&TrackPlayCount{}).Select("SUM(play_count)").Scan(&total).Error
	if err != nil {
		return 0, err
	}
	return total, nil
}

// GetArtistCounts 获取艺术家总数
func GetArtistCounts(ctx context.Context) (int64, error) {
	var count int64
	err := GetDB().WithContext(ctx).Model(&TrackPlayCount{}).Distinct("artist").Count(&count).Error
	if err != nil {
		return 0, err
	}
	return count, nil
}

// GetAlbumCounts 获取专辑总数
func GetAlbumCounts(ctx context.Context) (int64, error) {
	var count int64
	err := GetDB().WithContext(ctx).Model(&TrackPlayCount{}).Distinct("album").Count(&count).Error
	if err != nil {
		return 0, err
	}
	return count, nil
}

// GetTopArtistsByPlayCount 获取按播放次数统计的热门艺术家
func GetTopArtistsByPlayCount(ctx context.Context, limit int) ([]map[string]interface{}, error) {
	var result []map[string]interface{}
	err := GetDB().WithContext(ctx).Model(&TrackPlayCount{}).
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

// GetTopArtistsByTrackCount 获取按曲目数统计的热门艺术家
func GetTopArtistsByTrackCount(ctx context.Context, limit int) ([]map[string]interface{}, error) {
	var result []map[string]interface{}
	err := GetDB().WithContext(ctx).Model(&TrackPlayCount{}).
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

// GetTrackPlayCountsByPeriod 获取指定时间段内的曲目播放统计
func GetTrackPlayCountsByPeriod(ctx context.Context, limit int, offset int, period string) ([]*TrackPlayCount, error) {
	// 计算时间范围
	var startTime time.Time
	switch period {
	case "week":
		startTime = time.Now().AddDate(0, 0, -7)
	case "month":
		startTime = time.Now().AddDate(0, -1, 0)
	default:
		// 默认返回所有时间的数据
		return GetTrackPlayCounts(ctx, limit, offset)
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
	trackCountMap := make(map[string]*TrackPlayCount)
	for _, record := range playRecords {
		key := record.Artist + "|" + record.Album + "|" + record.Track
		if trackCount, exists := trackCountMap[key]; exists {
			trackCount.PlayCount++
		} else {
			trackCountMap[key] = &TrackPlayCount{
				Artist:    record.Artist,
				Album:     record.Album,
				Track:     record.Track,
				PlayCount: 1,
			}
		}
	}

	// 转换为切片并排序
	var trackCounts []*TrackPlayCount
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
		return []*TrackPlayCount{}, nil
	}
	if end > len(trackCounts) {
		end = len(trackCounts)
	}

	return trackCounts[start:end], nil
}

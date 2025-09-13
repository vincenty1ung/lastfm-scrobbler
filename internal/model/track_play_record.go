package model

import (
	"context"
	"time"

	"github.com/vincenty1ung/lastfm-scrobbler/common"
)

type TrackPlayRecord struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	Artist        string    `gorm:"index" json:"artist"`
	AlbumArtist   string    `json:"album_artist"`
	Track         string    `json:"track"`
	Album         string    `json:"album"`
	Duration      int64     `json:"duration"`
	PlayTime      time.Time `json:"play_time"`
	Scrobbled     bool      `gorm:"index" json:"scrobbled"` // 是否已同步到Last.fm
	MusicBrainzID string    `json:"musicbrainz_id"`
	TrackNumber   int64     `json:"track_number"`
	Source        string    `gorm:"index" json:"source"` // 数据来源：Audirvana、Roon 或 Apple Music
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func InsertTrackPlayRecord(ctx context.Context, record *TrackPlayRecord) error {
	// 验证记录中的艺术家、专辑和曲目信息
	if err := common.ValidateTrackInfo(ctx, record.Artist, record.Album, record.Track); err != nil {
		return err
	}

	return GetDB().WithContext(ctx).Create(record).Error
}

func UpdateScrobbledStatus(ctx context.Context, id uint, scrobbled bool) error {
	return GetDB().WithContext(ctx).Model(&TrackPlayRecord{}).Where("id = ?", id).Update("scrobbled", scrobbled).Error
}

func GetUnscrobbledRecords(ctx context.Context, limit int) ([]*TrackPlayRecord, error) {
	var trackPlayRecords []*TrackPlayRecord
	err := GetDB().WithContext(ctx).Where(
		"scrobbled = ?", false,
	).Order("play_time ASC").Limit(limit).Find(&trackPlayRecords).Error
	if err != nil {
		return nil, err
	}
	return trackPlayRecords, nil
}

// GetRecentPlayRecords 获取最近播放的记录
func GetRecentPlayRecords(ctx context.Context, limit int) ([]*TrackPlayRecord, error) {
	var records []*TrackPlayRecord
	err := GetDB().WithContext(ctx).Order("play_time DESC").Limit(limit).Find(&records).Error
	if err != nil {
		return nil, err
	}
	return records, nil
}

// GetRecentPlayRecordsByDays 获取指定天数内的播放记录
func GetRecentPlayRecordsByDays(ctx context.Context, days int) ([]*TrackPlayRecord, error) {
	var records []*TrackPlayRecord
	// 计算从现在开始往前推指定天数的时间
	startTime := time.Now().AddDate(0, 0, -days)
	err := GetDB().WithContext(ctx).Where("play_time >= ?", startTime).Order("play_time DESC").Find(&records).Error
	if err != nil {
		return nil, err
	}
	return records, nil
}

// GetUnscrobbledRecordsWithPagination 分页获取未同步到Last.fm的播放记录
func GetUnscrobbledRecordsWithPagination(ctx context.Context, limit, offset int) ([]*TrackPlayRecord, error) {
	var trackPlayRecords []*TrackPlayRecord
	err := GetDB().WithContext(ctx).Where(
		"scrobbled = ?", false,
	).Order("play_time ASC").Limit(limit).Offset(offset).Find(&trackPlayRecords).Error
	if err != nil {
		return nil, err
	}
	return trackPlayRecords, nil
}

// GetUnscrobbledRecordsCount 获取未同步到Last.fm的播放记录总数
func GetUnscrobbledRecordsCount(ctx context.Context) (int64, error) {
	var count int64
	err := GetDB().WithContext(ctx).Model(&TrackPlayRecord{}).Where("scrobbled = ?", false).Count(&count).Error
	if err != nil {
		return 0, err
	}
	return count, nil
}

// BatchUpdateScrobbledStatus 批量更新播放记录的同步状态
func BatchUpdateScrobbledStatus(ctx context.Context, ids []uint, scrobbled bool) error {
	return GetDB().WithContext(ctx).Model(&TrackPlayRecord{}).Where("id IN ?", ids).Update("scrobbled", scrobbled).Error
}

// GetUnscrobbledRecordsByIds 通过ID列表获取未同步的播放记录
func GetUnscrobbledRecordsByIds(ctx context.Context, ids []uint) ([]*TrackPlayRecord, error) {
	// 获取指定ID的未同步记录
	var records []*TrackPlayRecord
	err := GetDB().WithContext(ctx).Where("id IN ? AND scrobbled = ?", ids, false).Find(&records).Error
	if err != nil {
		return nil, err
	}

	return records, nil
}

// GetPlayCountsBySource 获取按来源统计的播放次数
func GetPlayCountsBySource(ctx context.Context) (map[string]int64, error) {
	var result []map[string]interface{}
	err := GetDB().WithContext(ctx).Model(&TrackPlayRecord{}).
		Select("source, COUNT(*) as count").
		Group("source").
		Find(&result).Error
	if err != nil {
		return nil, err
	}

	// 转换为map[string]int64
	sourceCounts := make(map[string]int64)
	for _, item := range result {
		if source, ok := item["source"].(string); ok {
			if count, ok := item["count"].(int64); ok {
				sourceCounts[source] = count
			} else if countFloat, ok := item["count"].(float64); ok {
				sourceCounts[source] = int64(countFloat)
			}
		}
	}

	return sourceCounts, nil
}

// GetTopAlbumsByPlayCount 获取按播放次数统计的热门专辑
type TopAlbum struct {
	Album     string `json:"album"`
	Artist    string `json:"artist"`
	PlayCount int    `json:"play_count"`
}

// GetTopAlbumsByPlayCount 获取按播放次数统计的热门专辑
func GetTopAlbumsByPlayCount(ctx context.Context, days int, limit int) ([]*TopAlbum, error) {
	var result []*TopAlbum
	
	// 计算时间范围
	var startTime time.Time
	if days > 0 {
		startTime = time.Now().AddDate(0, 0, -days)
	}

	// 构建查询
	query := GetDB().WithContext(ctx).Model(&TrackPlayRecord{})
	
	// 如果指定了时间范围，则添加时间条件
	if days > 0 {
		query = query.Where("play_time >= ?", startTime)
	}
	
	err := query.Select("album, MIN(artist) as artist, COUNT(album) as play_count").
		Group("album").
		Order("play_count DESC").
		Limit(limit).
		Find(&result).Error
		
	if err != nil {
		return nil, err
	}
	return result, nil
}

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
	return GetDB().WithContext(ctx).Where("id = ?", id).Update("scrobbled", scrobbled).Error
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

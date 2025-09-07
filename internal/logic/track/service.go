package track

import (
	"context"

	"github.com/vincenty1ung/lastfm-scrobbler/internal/model"
)

// TrackService 定义曲目相关服务接口
type TrackService interface {
	GetTrackPlayCounts(ctx context.Context, limit, offset int) ([]*model.TrackPlayCount, error)
	GetTrackPlayCount(ctx context.Context, artist, album, track string) (*model.TrackPlayCount, error)
	InsertTrackPlayRecord(ctx context.Context, record *model.TrackPlayRecord) error
	IncrementTrackPlayCount(ctx context.Context, artist, album, track string) error
	GetTotalPlayCount(ctx context.Context) (int64, error)
	GetTrackCounts(ctx context.Context) (int64, error)
	GetArtistCounts(ctx context.Context) (int64, error)
	GetAlbumCounts(ctx context.Context) (int64, error)
	GetRecentPlayRecords(ctx context.Context, limit int) ([]*model.TrackPlayRecord, error)
	// GetRecentPlayRecordsByDays 获取指定天数内的播放记录
	GetRecentPlayRecordsByDays(ctx context.Context, days int) ([]*model.TrackPlayRecord, error)
	// GetTopArtistsByPlayCount 获取按播放次数统计的热门艺术家
	GetTopArtistsByPlayCount(ctx context.Context, limit int) ([]map[string]interface{}, error)
	// GetTopArtistsByTrackCount 获取按曲目数统计的热门艺术家
	GetTopArtistsByTrackCount(ctx context.Context, limit int) ([]map[string]interface{}, error)
	// GetTrackPlayCountsByPeriod 获取指定时间段内的曲目播放统计
	GetTrackPlayCountsByPeriod(ctx context.Context, limit, offset int, period string) ([]*model.TrackPlayCount, error)
	// GetPlayCountsBySource 获取按来源统计的播放次数
	GetPlayCountsBySource(ctx context.Context) (map[string]int64, error)
}

// TrackServiceImpl 实现TrackService接口
type TrackServiceImpl struct{}

// NewTrackService 创建TrackService实例
func NewTrackService() TrackService {
	return &TrackServiceImpl{}
}

// GetTrackPlayCounts 获取曲目播放统计列表
func (s *TrackServiceImpl) GetTrackPlayCounts(ctx context.Context, limit, offset int) (
	[]*model.TrackPlayCount, error,
) {
	return model.GetTrackPlayCounts(ctx, limit, offset)
}

// GetTrackPlayCount 获取特定曲目的播放统计
func (s *TrackServiceImpl) GetTrackPlayCount(ctx context.Context, artist, album, track string) (
	*model.TrackPlayCount, error,
) {
	return model.GetTrackPlayCount(ctx, artist, album, track)
}

func (s *TrackServiceImpl) InsertTrackPlayRecord(ctx context.Context, record *model.TrackPlayRecord) error {
	return model.InsertTrackPlayRecord(ctx, record)
}

func (s *TrackServiceImpl) IncrementTrackPlayCount(ctx context.Context, artist, album, track string) error {
	return model.IncrementTrackPlayCount(ctx, artist, album, track)
}

// GetTotalPlayCount 获取总播放次数
func (s *TrackServiceImpl) GetTotalPlayCount(ctx context.Context) (int64, error) {
	return model.GetTotalPlayCount(ctx)
}

// GetTrackCounts 获取曲目总数
func (s *TrackServiceImpl) GetTrackCounts(ctx context.Context) (int64, error) {
	return model.GetTrackCounts(ctx)
}

// GetArtistCounts 获取艺术家总数
func (s *TrackServiceImpl) GetArtistCounts(ctx context.Context) (int64, error) {
	return model.GetArtistCounts(ctx)
}

// GetAlbumCounts 获取专辑总数
func (s *TrackServiceImpl) GetAlbumCounts(ctx context.Context) (int64, error) {
	return model.GetAlbumCounts(ctx)
}

// GetRecentPlayRecords 获取最近播放记录
func (s *TrackServiceImpl) GetRecentPlayRecords(ctx context.Context, limit int) ([]*model.TrackPlayRecord, error) {
	return model.GetRecentPlayRecords(ctx, limit)
}

// GetRecentPlayRecordsByDays 获取指定天数内的播放记录
func (s *TrackServiceImpl) GetRecentPlayRecordsByDays(ctx context.Context, days int) ([]*model.TrackPlayRecord, error) {
	return model.GetRecentPlayRecordsByDays(ctx, days)
}

// GetTopArtistsByPlayCount 获取按播放次数统计的热门艺术家
func (s *TrackServiceImpl) GetTopArtistsByPlayCount(ctx context.Context, limit int) ([]map[string]interface{}, error) {
	return model.GetTopArtistsByPlayCount(ctx, limit)
}

// GetTopArtistsByTrackCount 获取按曲目数统计的热门艺术家
func (s *TrackServiceImpl) GetTopArtistsByTrackCount(ctx context.Context, limit int) ([]map[string]interface{}, error) {
	return model.GetTopArtistsByTrackCount(ctx, limit)
}

// GetTrackPlayCountsByPeriod 获取指定时间段内的曲目播放统计
func (s *TrackServiceImpl) GetTrackPlayCountsByPeriod(ctx context.Context, limit, offset int, period string) ([]*model.TrackPlayCount, error) {
	return model.GetTrackPlayCountsByPeriod(ctx, limit, offset, period)
}

// GetPlayCountsBySource 获取按来源统计的播放次数
func (s *TrackServiceImpl) GetPlayCountsBySource(ctx context.Context) (map[string]int64, error) {
	return model.GetPlayCountsBySource(ctx)
}

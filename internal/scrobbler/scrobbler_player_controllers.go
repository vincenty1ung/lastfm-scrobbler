package scrobbler

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/vincenty1ung/lastfm-scrobbler/common"
	"github.com/vincenty1ung/lastfm-scrobbler/core/applemusic"
	"github.com/vincenty1ung/lastfm-scrobbler/core/audirvana"
	"github.com/vincenty1ung/lastfm-scrobbler/core/exec"
	"github.com/vincenty1ung/lastfm-scrobbler/core/log"
	"github.com/vincenty1ung/lastfm-scrobbler/internal/cache"
)

type BaseWrapper struct {
}

func (m BaseWrapper) ConversionSimplified(target string) string {
	return common.ConversionSimplifiedFx(target)
}

// AudirvanaTrackInfoWrapper 包装 Audirvana TrackInfo 以实现 PlayerInfoHandler 接口
type AudirvanaTrackInfoWrapper struct {
	*audirvana.TrackInfo
	baseWrapper BaseWrapper
}

func (a *AudirvanaTrackInfoWrapper) GetTitle() string {
	return a.baseWrapper.ConversionSimplified(a.Title)
}

func (a *AudirvanaTrackInfoWrapper) GetAlbum() string {
	return a.baseWrapper.ConversionSimplified(a.MataDataHandle.GetAlbum())
}

func (a *AudirvanaTrackInfoWrapper) GetArtist() string {
	return a.baseWrapper.ConversionSimplified(a.MataDataHandle.GetArtist())
}

func (a *AudirvanaTrackInfoWrapper) GetPosition() float64 {
	return a.Position
}

func (a *AudirvanaTrackInfoWrapper) GetDuration() int64 {
	return a.Duration
}

func (a *AudirvanaTrackInfoWrapper) GetUrl() string {
	return a.Url
}

// 新增方法实现
func (a *AudirvanaTrackInfoWrapper) GetAlbumArtist() string {
	// Audirvana没有直接提供专辑艺术家信息，使用普通艺术家作为默认值
	return a.baseWrapper.ConversionSimplified(a.MataDataHandle.GetArtist())
}

func (a *AudirvanaTrackInfoWrapper) GetTrackNumber() int64 {
	// Audirvana没有直接提供曲目编号
	return a.MataDataHandle.GetTrackNumber()
}

func (a *AudirvanaTrackInfoWrapper) GetGenre() string {
	// Audirvana没有直接提供流派信息
	return cache.GetEnglishGenre(common.GenreCustomFit(a.MataDataHandle.GetGenre()))
}

func (a *AudirvanaTrackInfoWrapper) GetComposer() string {
	// Audirvana没有直接提供作曲家信息
	return a.baseWrapper.ConversionSimplified(a.MataDataHandle.GetComposer())
}

func (a *AudirvanaTrackInfoWrapper) GetReleaseDate() string {
	// Audirvana没有直接提供发布日期
	return a.MataDataHandle.GetReleaseDate()
}

func (a *AudirvanaTrackInfoWrapper) GetMusicBrainzID() string {
	// Audirvana没有直接提供MusicBrainz ID
	return a.MataDataHandle.GetMusicBrainzTrackId()
}

func (a *AudirvanaTrackInfoWrapper) GetSource() string {
	return a.MataDataHandle.GetSource()
}

func (a *AudirvanaTrackInfoWrapper) GetBundleID() string {
	// Audirvana没有直接提供应用标识符
	return a.MataDataHandle.GetBundleID()
}

func (a *AudirvanaTrackInfoWrapper) GetUniqueID() string {
	// 使用URL作为唯一标识符
	return a.MataDataHandle.GetUniqueID()
}

// AudirvanaPlayerController Audirvana播放器控制器
type AudirvanaPlayerController struct{}

func (a *AudirvanaPlayerController) IsRunning(ctx context.Context) bool {
	return audirvana.IsRunning(ctx)
}

func (a *AudirvanaPlayerController) GetState(ctx context.Context) (string, error) {
	state, err := audirvana.GetState(ctx)
	return string(state), err
}

func (a *AudirvanaPlayerController) GetNowPlayingTrackInfo(ctx context.Context) PlayerInfoHandler {
	info := audirvana.GetNowPlayingTrackInfo(ctx)
	if info == nil {
		return nil
	}
	return &AudirvanaTrackInfoWrapper{info, BaseWrapper{}}
}
func (a *AudirvanaPlayerController) SetFavorite(ctx context.Context) error {
	return nil
}

func (a *AudirvanaPlayerController) IsFavorite(ctx context.Context) bool {
	return false
}

// RoonTrackInfoWrapper 包装 MRMediaNowPlaying 以实现 PlayerInfoHandler 接口
type RoonTrackInfoWrapper struct {
	*exec.MediaControlNowPlayingInfo
	baseWrapper BaseWrapper
}

func (r *RoonTrackInfoWrapper) GetTitle() string {
	return r.baseWrapper.ConversionSimplified(r.Title)
}

func (r *RoonTrackInfoWrapper) GetAlbum() string {
	return r.baseWrapper.ConversionSimplified(r.Album)
}

func (r *RoonTrackInfoWrapper) GetArtist() string {
	return r.baseWrapper.ConversionSimplified(r.Artist)
}

func (r *RoonTrackInfoWrapper) GetPosition() float64 {
	return r.ElapsedTimeNow
}

func (r *RoonTrackInfoWrapper) GetDuration() int64 {
	return int64(r.Duration)
}

func (r *RoonTrackInfoWrapper) GetUrl() string {
	return ""
}

// 新增方法实现
func (r *RoonTrackInfoWrapper) GetAlbumArtist() string {
	// Roon没有直接提供专辑艺术家信息，使用普通艺术家作为默认值
	return r.baseWrapper.ConversionSimplified(r.Artist)
}

func (r *RoonTrackInfoWrapper) GetTrackNumber() int64 {
	// Roon没有直接提供曲目编号
	return int64(r.TrackNumber)
}

func (r *RoonTrackInfoWrapper) GetGenre() string {
	// Roon没有直接提供流派信息
	return cache.GetEnglishGenre(r.Genre)
}

func (r *RoonTrackInfoWrapper) GetComposer() string {
	// Roon没有直接提供作曲家信息
	return ""
}

func (r *RoonTrackInfoWrapper) GetReleaseDate() string {
	// Roon没有直接提供发布日期
	return ""
}

func (r *RoonTrackInfoWrapper) GetMusicBrainzID() string {
	// Roon没有直接提供MusicBrainz ID
	return ""
}

func (r *RoonTrackInfoWrapper) GetSource() string {
	return "Roon"
}

func (r *RoonTrackInfoWrapper) GetBundleID() string {
	// 从BundleIdentifier获取
	return r.BundleIdentifier
}

func (r *RoonTrackInfoWrapper) GetUniqueID() string {
	// Roon没有直接提供唯一标识符
	return r.ContentItemIdentifier
}

// RoonPlayerController Roon播放器控制器
type RoonPlayerController struct{}

func (r *RoonPlayerController) IsRunning(ctx context.Context) bool {
	// todo
	playing, err := exec.GetMediaControlNowPlaying(ctx)
	if err != nil {
		return false
	}
	return playing.BundleIdentifier == exec.MRMediaNowPlayingAppRoon
}

func (r *RoonPlayerController) GetState(ctx context.Context) (string, error) {
	playing, err := exec.GetMediaControlNowPlaying(ctx)
	if err != nil {
		return "", err
	}
	if playing.Playing {
		return common.PlayerStatePlaying, nil
	}
	return common.PlayerStateStopped, nil
}

func (r *RoonPlayerController) GetNowPlayingTrackInfo(ctx context.Context) PlayerInfoHandler {
	playing, err := exec.GetMediaControlNowPlaying(ctx)
	if err != nil {
		return nil
	}
	return &RoonTrackInfoWrapper{playing, BaseWrapper{}}
}

func (a *RoonPlayerController) SetFavorite(ctx context.Context) error {
	return nil
}
func (a *RoonPlayerController) IsFavorite(ctx context.Context) bool {
	return false
}

// AppleMusicTrackInfoWrapper 包装 AppleMusic TrackInfo 以实现 PlayerInfoHandler 接口
type AppleMusicTrackInfoWrapper struct {
	*applemusic.TrackInfo
	baseWrapper BaseWrapper
}

func (a *AppleMusicTrackInfoWrapper) GetTitle() string {
	return a.baseWrapper.ConversionSimplified(a.Title)
}

func (a *AppleMusicTrackInfoWrapper) GetAlbum() string {
	return a.baseWrapper.ConversionSimplified(a.Album)
}

func (a *AppleMusicTrackInfoWrapper) GetArtist() string {
	return a.baseWrapper.ConversionSimplified(a.Artist)
}

func (a *AppleMusicTrackInfoWrapper) GetPosition() float64 {
	return a.Position
}

func (a *AppleMusicTrackInfoWrapper) GetDuration() int64 {
	return a.Duration
}

func (a *AppleMusicTrackInfoWrapper) GetUrl() string {
	return a.Url
}

// 新增方法实现
func (a *AppleMusicTrackInfoWrapper) GetAlbumArtist() string {
	return a.baseWrapper.ConversionSimplified(a.Artist)
}

func (a *AppleMusicTrackInfoWrapper) GetTrackNumber() int64 {
	return int64(a.TrackNumber)
}

func (a *AppleMusicTrackInfoWrapper) GetGenre() string {
	return cache.GetEnglishGenre(common.GenreCustomFit(a.Genre))
}

func (a *AppleMusicTrackInfoWrapper) GetComposer() string {
	return a.baseWrapper.ConversionSimplified(a.Composer)
}

func (a *AppleMusicTrackInfoWrapper) GetReleaseDate() string {
	if !a.ReleaseDate.IsZero() {
		return a.ReleaseDate.Format("2006-01-02")
	}
	return ""
}

func (a *AppleMusicTrackInfoWrapper) GetMusicBrainzID() string {
	// Apple Music没有直接提供MusicBrainz ID
	return ""
}

func (a *AppleMusicTrackInfoWrapper) GetSource() string {
	return fmt.Sprintf("%d", a.DatabaseID)
}

func (a *AppleMusicTrackInfoWrapper) GetBundleID() string {
	return a.BundleIdentifier
}

func (a *AppleMusicTrackInfoWrapper) GetUniqueID() string {
	return fmt.Sprintf("%d", a.DatabaseID)
}

// AppleMusicPlayerController Apple Music播放器控制器
type AppleMusicPlayerController struct{}

func (a *AppleMusicPlayerController) IsRunning(ctx context.Context) bool {
	return applemusic.IsRunning(ctx)
}

func (a *AppleMusicPlayerController) GetState(ctx context.Context) (string, error) {
	state, err := applemusic.GetState(ctx)
	return string(state), err
}

func (a *AppleMusicPlayerController) GetNowPlayingTrackInfo(ctx context.Context) PlayerInfoHandler {
	info := applemusic.GetNowPlayingTrackInfo(ctx)
	if info == nil {
		return nil
	}
	return &AppleMusicTrackInfoWrapper{info, BaseWrapper{}}
}

func (a *AppleMusicPlayerController) SetFavorite(ctx context.Context) error {
	favorite := a.IsFavorite(ctx)
	if !favorite {
		err := applemusic.SetFavorite(ctx, true)
		if err != nil {
			log.Warn(ctx, "AppleMusicPlayerController SetFavorite", zap.Error(err))
			return err
		}
	}
	return nil
}
func (a *AppleMusicPlayerController) IsFavorite(ctx context.Context) bool {
	favorite, err := applemusic.IsFavorite(ctx)
	if err != nil {
		log.Warn(ctx, "AppleMusicPlayerController IsFavorite", zap.Error(err))
		return false
	}
	return favorite
}

// 网易云
/* {
  "playbackRate" : 1,
  "album" : "铸铁旅人",
  "elapsedTimeNow" : 401.89587608909608,
  "elapsedTime" : 297.21600000000001,
  "timestamp" : "2025-09-13T02:53:11Z",
  "bundleIdentifier" : "com.netease.163music",
  "processIdentifier" : 41260,
  "title" : "铸铁旅人",
  "duration" : 520.12697916666662,
  "artist" : "虎啸春",
  "contentItemIdentifier" : "C4B45625-FB20-419B-BFA0-42CCEC333EA4",
  "playing" : true
} */

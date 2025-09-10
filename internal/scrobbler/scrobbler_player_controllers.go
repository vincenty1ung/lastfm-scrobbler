package scrobbler

import (
	"context"

	"github.com/vincenty1ung/lastfm-scrobbler/common"
	"github.com/vincenty1ung/lastfm-scrobbler/core/applemusic"
	"github.com/vincenty1ung/lastfm-scrobbler/core/audirvana"
	"github.com/vincenty1ung/lastfm-scrobbler/core/exec"
)

// AudirvanaTrackInfoWrapper 包装 Audirvana TrackInfo 以实现 PlayerInfoHandler 接口
type AudirvanaTrackInfoWrapper struct {
	*audirvana.TrackInfo
	baseWrapper BaseWrapper
}

func (a *AudirvanaTrackInfoWrapper) GetTitle() string {
	return a.baseWrapper.conversionSimplified(a.Title)
}

func (a *AudirvanaTrackInfoWrapper) GetAlbum() string {
	return a.baseWrapper.conversionSimplified(a.Album)
}

func (a *AudirvanaTrackInfoWrapper) GetArtist() string {
	return a.baseWrapper.conversionSimplified(a.Artist)
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

// RoonTrackInfoWrapper 包装 MRMediaNowPlaying 以实现 PlayerInfoHandler 接口
type RoonTrackInfoWrapper struct {
	*exec.MRMediaNowPlaying
	baseWrapper BaseWrapper
}

func (r *RoonTrackInfoWrapper) GetTitle() string {
	return r.baseWrapper.conversionSimplified(r.Title)
}

func (r *RoonTrackInfoWrapper) GetAlbum() string {
	return r.baseWrapper.conversionSimplified(r.Album)
}

func (r *RoonTrackInfoWrapper) GetArtist() string {
	return r.baseWrapper.conversionSimplified(r.Artist)
}

func (r *RoonTrackInfoWrapper) GetPosition() float64 {
	return r.ElapsedTime
}

func (r *RoonTrackInfoWrapper) GetDuration() int64 {
	return int64(r.Duration)
}

func (r *RoonTrackInfoWrapper) GetUrl() string {
	return ""
}

// RoonPlayerController Roon播放器控制器
type RoonPlayerController struct{}

func (r *RoonPlayerController) IsRunning(ctx context.Context) bool {
	playing, err := exec.GetMRMediaNowPlaying()
	if err != nil {
		return false
	}
	return playing.BundleIdentifier == exec.MRMediaNowPlayingAppRoon
}

func (r *RoonPlayerController) GetState(ctx context.Context) (string, error) {
	playing, err := exec.GetMRMediaNowPlaying()
	if err != nil {
		return "", err
	}
	if playing.IsPlaying {
		return common.PlayerStatePlaying, nil
	}
	return common.PlayerStateStopped, nil
}

func (r *RoonPlayerController) GetNowPlayingTrackInfo(ctx context.Context) PlayerInfoHandler {
	playing, err := exec.GetMRMediaNowPlaying()
	if err != nil {
		return nil
	}
	return &RoonTrackInfoWrapper{playing, BaseWrapper{}}
}

// AppleMusicTrackInfoWrapper 包装 AppleMusic TrackInfo 以实现 PlayerInfoHandler 接口
type AppleMusicTrackInfoWrapper struct {
	*applemusic.TrackInfo
	baseWrapper BaseWrapper
}

func (a *AppleMusicTrackInfoWrapper) GetTitle() string {
	return a.baseWrapper.conversionSimplified(a.Title)
}

func (a *AppleMusicTrackInfoWrapper) GetAlbum() string {
	return a.baseWrapper.conversionSimplified(a.Album)
}

func (a *AppleMusicTrackInfoWrapper) GetArtist() string {
	return a.baseWrapper.conversionSimplified(a.Artist)
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

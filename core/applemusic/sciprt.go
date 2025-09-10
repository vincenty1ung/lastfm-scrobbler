package applemusic

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"go.uber.org/zap"

	"github.com/vincenty1ung/lastfm-scrobbler/core/applesciprt"
	alog "github.com/vincenty1ung/lastfm-scrobbler/core/log"

	"github.com/vincenty1ung/lastfm-scrobbler/common"
)

type (
	TrackBase struct {
		TrackID    string
		Title      string
		Album      string
		Artist     string
		Duration   int64
		Position   float64
		Url        string
		AirfoiLogo string
	}
	TrackInfo struct {
		TrackBase
	}
)

func IsRunning(ctx context.Context) bool {
	tell, err := applesciprt.Tell(
		common.AppSystemEvents, fmt.Sprintf(
			`set listApplicationProcessNames to name of every application process
			if listApplicationProcessNames contains "%s" then
				set APPLE_MUSIC_RUNNING_STATE to "true"
			else
				set APPLE_MUSIC_RUNNING_STATE to "false"
			end if`, "Music",
		),
	)
	if err != nil {
		return false
	}

	parseBool, err := strconv.ParseBool(tell)
	if err != nil {
		alog.Warn(ctx, "err:", zap.Error(err))
		return false
	}
	return parseBool
}

// GetState 使用 AppleScript 从 Apple Music 应用获取当前播放器状态。
// 返回播放器状态（common.PlayerState）以及过程中遇到的任何错误。
func GetState(ctx context.Context) (playerState common.PlayerState, err error) {
	result, err := applesciprt.Tell("Music", `set musicState to get player state`)
	if err != nil {
		alog.Warn(ctx, "err:", zap.Error(err))
		return "", err
	}
	if result == "playing" {
		return common.PlayerStatePlaying, nil
	} else if result == "paused" {
		return common.PlayerStatePaused, nil
	} else {
		return common.PlayerState(result), nil
	}
}

// GetNowPlayingTrackInfo 使用 AppleScript 从 Apple Music 获取当前正在播放的曲目信息。
// 它返回一个指向 TrackInfo 结构体的指针，包含曲目的标题、专辑、艺术家、时长、播放位置和 URL。
// 如果在执行 AppleScript 或解析数据时发生错误，函数会记录警告并返回 nil。
func GetNowPlayingTrackInfo(ctx context.Context) *TrackInfo {
	tell, err := applesciprt.Tell(
		"Music",
		`set playingTrack to name of current track
set playingAlbum to album of current track
set playingArtist to artist of current track
set playingDuration to duration of current track
set playingDuration to playingDuration as string
set playingPosition to player position
set playingPosition to playingPosition as string
set playingUrl to ""
try
	set playingUrl to database ID of current track
on error
	set playingUrl to ""
end try
set result to playingTrack & "|" & playingAlbum & "|" & playingArtist & "|" & playingDuration & "|" & playingPosition & "|" & playingUrl`,
	)
	if err != nil {
		alog.Warn(ctx, "err:", zap.Error(err))
		return nil
	}
	split := strings.Split(tell, "|")
	info := &TrackInfo{
		TrackBase: TrackBase{},
	}
	for i, s := range split {
		switch i {
		case 0:
			info.Title = strings.TrimSpace(s)
		case 1:
			info.Album = strings.TrimSpace(s)
		case 2:
			info.Artist = strings.TrimSpace(s)
		case 3:
			// Duration in seconds
			parseFloat, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
			if err != nil {
				alog.Warn(ctx, "err:", zap.Error(err))
				return nil
			}
			info.Duration = int64(parseFloat)
		case 4:
			// Position in seconds
			parseFloat, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
			if err != nil {
				alog.Warn(ctx, "err:", zap.Error(err))
				return nil
			}
			info.Position = parseFloat
		case 5:
			info.Url = strings.TrimSpace(s)
		}
	}
	return info
}

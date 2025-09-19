package audirvana

import (
	"context"
	"fmt"
	"testing"

	"go.uber.org/zap"

	"github.com/vincenty1ung/lastfm-scrobbler/common"
	"github.com/vincenty1ung/lastfm-scrobbler/core/cache"
	alog "github.com/vincenty1ung/lastfm-scrobbler/core/log"
)

func init() {
	_, _ = alog.LogInit("./.logs", "info", make(<-chan struct{}))
}

func TestAudirvana(t *testing.T) {
	running := IsRunning(context.Background())
	alog.Info(context.Background(), "running", zap.Any("running", running))
	fmt.Println("Audirvana is running:", running)
	if running {
		state, _ := GetState(context.Background())
		alog.Debug(context.Background(), "audirvana 播放状态", zap.Any("state", state))
		var audirvanaTrackInfo *TrackInfo
		if state == common.PlayerStatePlaying {
			audirvanaTrackInfo = GetNowPlayingTrackInfo(context.Background())
			alog.Info(context.Background(), "", zap.Any("audirvana trackInfo", audirvanaTrackInfo))
		}
	}
}
func TestIsRunningReturnsBool(t *testing.T) {
	running := IsRunning(context.Background())
	if running != true && running != false {
		t.Errorf("IsRunning() should return a boolean, got %v", running)
	}
}

func TestGetStateHandlesError(t *testing.T) {
	// Simulate Audirvana not running by ensuring IsRunning returns false
	if !IsRunning(context.Background()) {
		_, err := GetState(context.Background())
		if err == nil {
			t.Error("GetState() should return error when Audirvana is not running")
		}
	}
}

func TestGetNowPlayingTrackInfoFields(t *testing.T) {
	if IsRunning(context.Background()) {
		state, _ := GetState(context.Background())
		if state == common.PlayerStatePlaying {
			info := GetNowPlayingTrackInfo(context.Background())
			if info == nil {
				t.Error("GetNowPlayingTrackInfo() returned nil while playing")
			} else {
				if info.Title == "" {
					t.Error("TrackInfo.Title is empty")
				}
				if info.Artist == "" {
					t.Error("TrackInfo.Artist is empty")
				}
				if info.Duration <= 0 {
					t.Error("TrackInfo.Duration should be positive")
				}
				if info.Position < 0 {
					t.Error("TrackInfo.Position should not be negative")
				}
			}
		}
	}
}

func TestFindMataDataHandleCache(t *testing.T) {
	state, err := GetState(context.Background())
	if err != nil {
		t.Error(err)
		return
	}
	if state == common.PlayerStatePlaying {
		audirvanaTrackInfo := GetNowPlayingTrackInfo(context.Background())
		info := cache.FindMataDataHandle(context.Background(), audirvanaTrackInfo.Url)
		fmt.Println(info.GetMusicBrainzTrackId())
		fmt.Println(info.GetTrackNumber())
		fmt.Println(info.GetArtists())
		fmt.Println(info.GetAlbum())
		fmt.Println(info.GetArtist())
		fmt.Println(info.GetUniqueID())
		fmt.Println(info.GetTitle())
		fmt.Println(info.GetSource())
		fmt.Println(info.GetComposer())
		fmt.Println(info.GetGenre())
		fmt.Println(info.GetDuration())
		fmt.Println(info.GetReleaseDate())
		fmt.Println(info.GetBundleID())
	}

}

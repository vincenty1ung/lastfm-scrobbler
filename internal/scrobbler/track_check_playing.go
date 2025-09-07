package scrobbler

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"go.uber.org/zap"

	"github.com/vincenty1ung/lastfm-scrobbler/common"
	"github.com/vincenty1ung/lastfm-scrobbler/core/audirvana"
	"github.com/vincenty1ung/lastfm-scrobbler/core/exec"
	"github.com/vincenty1ung/lastfm-scrobbler/core/lastfm"
	"github.com/vincenty1ung/lastfm-scrobbler/core/log"
	"github.com/vincenty1ung/lastfm-scrobbler/core/telemetry"
	"github.com/vincenty1ung/lastfm-scrobbler/core/websocket"
	"github.com/vincenty1ung/lastfm-scrobbler/internal/logic/track"
	"github.com/vincenty1ung/lastfm-scrobbler/internal/model"
)

const (
	cAudirvana  = "audirvana"
	cRoon       = "roon"
	cAppleMusic = "applemusic"

	percentScrobble = 0.55
	defaultSleep    = 3
	longSleep       = 60 // 休眠间隔六十秒
	checkCount      = 100
)

var (
	newTrackService = track.NewTrackService()

	mapedAudirvana  = make(map[string]bool)
	mapedRoon       = make(map[string]bool)
	mapedAppleMusic = make(map[string]bool)

	isLongAudirvana  bool
	isLongRoon       bool
	isLongAppleMusic bool

	pushCount           = atomic.Uint32{} // 多渠道上报
	atomicPlaying       = atomic.Bool{}   // 并发播放状态
	currentPlayingCache = sync.Map{}      // 本地缓存当前播放信息
)

func Init(
	ctx context.Context, apiKey, apiSecret, userLoginToken string, isMobile bool, userUsername, userPassword string,
) {
	lastfm.InitLastfmApi(
		ctx,
		apiKey,
		apiSecret,
		userLoginToken,
		isMobile,
		userUsername,
		userPassword,
	)
}

func AudirvanaCheckPlayingTrack(ctx context.Context, stop <-chan struct{}) {
	timer := time.NewTicker(time.Second * defaultSleep)
	var (
		now           time.Time
		currentTrack  = ""
		previousTrack = ""
		tmpCount      = 0
	)
	counts, err := model.GetTrackCounts(ctx)
	if err != nil {
		panic(err)
	}
	pushCount.Store(uint32(counts))
	for {
		select {
		case <-timer.C:
			h := func(ctx context.Context) {
				// Start a new span for this check cycle
				checkCtx, span := telemetry.StartSpanForTracerName(
					ctx, _TracerName, "audirvanaCheckPlayingTrack",
				)
				defer span.End()

				log.Debug(checkCtx, "AudirvanaCheckPlayingTrack Checking playing track..."+time.Now().String())

				tmpCount++
				if tmpCount > checkCount && !isLongAudirvana { // 检查100次依旧没有播放检查轮训放大到60秒
					timer.Reset(time.Second * longSleep)
					isLongAudirvana = true
					log.Info(
						checkCtx, "Audirvana 检查100次依旧没有播放检查轮训放大到60秒",
						zap.Uint32("共计上传歌曲标记", pushCount.Load()),
					)
				}
				if isLongAudirvana {
					log.Info(checkCtx, "60秒检查", zap.Uint32("共计上传歌曲标记", pushCount.Load()))
				}
				running := audirvana.IsRunning(checkCtx)
				log.Debug(checkCtx, "Audirvana 程序运行是否运行", zap.Bool("running", running))
				var audirvanaTrackInfo *audirvana.TrackInfo
				if running {
					audirvanaTrackInfo = nil
					state, _ := audirvana.GetState(checkCtx)
					log.Debug(checkCtx, "Audirvana 播放状态", zap.Any("state", state))
					if state == common.PlayerStatePlaying {
						if tmpCount > checkCount {
							isLongAudirvana = false
							timer.Reset(time.Second * defaultSleep)
						}
						tmpCount = 0
						audirvanaTrackInfo = audirvana.GetNowPlayingTrackInfo(checkCtx)
					} else {
						if _, ok := currentPlayingCache.Load(cAudirvana); ok {
							currentPlayingCache.Delete(cAudirvana)
							_, aok := currentPlayingCache.Load(cRoon)
							if !aok {
								websocket.BroadcastMessage(
									checkCtx,
									&websocket.WsTrackInfo{
										Type:   "stop",
										Source: cAudirvana,
									},
								)
								atomicPlaying.Store(false)
							}
						}
					}
				}
				if audirvanaTrackInfo != nil {
					tmpTrack := audirvanaTrackInfo.Url + audirvanaTrackInfo.Title // 防止cue文件出现问题
					currentTrack = tmpTrack
					position := audirvanaTrackInfo.Position
					duration := audirvanaTrackInfo.Duration
					wti := &websocket.WsTrackInfo{
						Type:   "now_playing",
						Source: cAudirvana,
						Data: struct {
							Title  string `json:"title"`
							Album  string `json:"album"`
							Artist string `json:"artist"`
						}{
							audirvanaTrackInfo.Title,
							audirvanaTrackInfo.Album,
							audirvanaTrackInfo.Artist,
						},
					}
					// 将播放信息写入本地缓存
					currentPlayingCache.Store("audirvana", wti)
					atomicPlaying.Store(true)
					// 向WebSocket客户端广播播放信息
					websocket.BroadcastMessage(
						checkCtx,
						wti,
					)
					if position/float64(duration) > percentScrobble && !mapedAudirvana[currentTrack] {
						// 标记听歌完成
						pushTrackScrobbleReq := &lastfm.PushTrackScrobbleReq{
							Artist:      audirvanaTrackInfo.Artist,
							AlbumArtist: audirvanaTrackInfo.Artist,
							Track:       audirvanaTrackInfo.Title,
							Album:       audirvanaTrackInfo.Album,
							Duration:    audirvanaTrackInfo.Duration,
							Timestamp:   now.UTC().Unix(),
						}
						// 说明在听歌存在有效数据的
						if mataDataHandleCache := exec.FindMataDataHandleCache(
							checkCtx, audirvanaTrackInfo.Url,
						); mataDataHandleCache != nil {
							pushTrackScrobbleReq.TrackNumber = mataDataHandleCache.GetTrackNumber()
							pushTrackScrobbleReq.MusicBrainzTrackID = mataDataHandleCache.GetMusicBrainzTrackId()
							if artist := mataDataHandleCache.GetArtist(); len(artist) != 0 {
								pushTrackScrobbleReq.Artist = artist
							}
							if albumartist := mataDataHandleCache.GetAlbumartist(); len(albumartist) != 0 {
								pushTrackScrobbleReq.AlbumArtist = albumartist
							}
						}
						// Save to database
						record := &model.TrackPlayRecord{
							Artist:        pushTrackScrobbleReq.Artist,
							AlbumArtist:   pushTrackScrobbleReq.AlbumArtist,
							Track:         pushTrackScrobbleReq.Track,
							Album:         pushTrackScrobbleReq.Album,
							Duration:      pushTrackScrobbleReq.Duration,
							PlayTime:      time.Unix(pushTrackScrobbleReq.Timestamp, 0),
							Scrobbled:     true,
							MusicBrainzID: pushTrackScrobbleReq.MusicBrainzTrackID,
							TrackNumber:   pushTrackScrobbleReq.TrackNumber,
							Source:        "Audirvana",
						}

						if err := newTrackService.InsertTrackPlayRecord(ctx, record); err != nil {
							log.Warn(checkCtx, "Audirvana Failed to insert track play record", zap.Error(err))
						}

						// Update track play count
						if err := newTrackService.IncrementTrackPlayCount(
							checkCtx, record.Artist, record.Album, record.Track,
						); err != nil {
							log.Warn(checkCtx, "Audirvana Failed to increment track play count", zap.Error(err))
						}
						_, err := lastfm.PushTrackScrobble(checkCtx, pushTrackScrobbleReq)
						if err != nil {
							log.Warn(checkCtx, "AudirvanaTrackUpdateNowPlaying err", zap.Error(err))
						}
						mapedAudirvana[currentTrack] = true
						pushCount.Add(1)
						log.Info(checkCtx, "Audirvana标记听歌完成", zap.String("track", pushTrackScrobbleReq.Track))
					}
					// 上传听歌ing
					if currentTrack != previousTrack {
						// 产生新歌曲
						delete(mapedAudirvana, previousTrack)
						now = time.Now()
						playingReq := lastfm.TrackUpdateNowPlayingReq{
							Artist:      audirvanaTrackInfo.Artist,
							AlbumArtist: audirvanaTrackInfo.Artist,
							Track:       audirvanaTrackInfo.Title,
							Album:       audirvanaTrackInfo.Album,
							Duration:    audirvanaTrackInfo.Duration,
						}
						// 说明在听歌存在有效数据的
						if mataDataHandleCache := exec.FindMataDataHandleCache(
							ctx, audirvanaTrackInfo.Url,
						); mataDataHandleCache != nil {
							playingReq.TrackNumber = mataDataHandleCache.GetTrackNumber()
							playingReq.MusicBrainzTrackID = mataDataHandleCache.GetMusicBrainzTrackId()
							if artist := mataDataHandleCache.GetArtist(); len(artist) != 0 {
								playingReq.Artist = artist
							}
							if albumartist := mataDataHandleCache.GetAlbumartist(); len(albumartist) != 0 {
								playingReq.AlbumArtist = albumartist
							}
						}
						log.Info(
							checkCtx, "AudirvanaNowPlayingTrackInfo", zap.Any("audirvanaTrackInfo", audirvanaTrackInfo),
						)
						err := lastfm.TrackUpdateNowPlaying(checkCtx, &playingReq)
						if err != nil {
							log.Warn(checkCtx, "AudirvanaTrackUpdateNowPlaying err", zap.Error(err))
						}
					}
					previousTrack = tmpTrack // 防止cue文件出现问题
				}
			}
			h(ctx)
		case <-stop:
			fmt.Println("check playing track exit")
			return
		}
	}
}
func RoonCheckPlayingTrack(ctx context.Context, stop <-chan struct{}) {
	timer := time.NewTicker(time.Second * defaultSleep)
	var (
		now           time.Time
		currentTrack  = ""
		previousTrack = ""
		tmpCount      = 0
	)
	for {
		select {
		case <-timer.C:
			h := func(ctx context.Context) {
				// Start a new span for this check cycle
				checkCtx, span := telemetry.StartSpanForTracerName(
					ctx, _TracerName, "roonCheckPlayingTrack",
				)
				defer span.End()

				log.Debug(checkCtx, "RoonCheckPlayingTrack Checking playing track..."+time.Now().String())
				tmpCount++
				if tmpCount > checkCount && !isLongRoon { // 检查100次依旧没有播放检查轮训放大到60秒
					timer.Reset(time.Second * longSleep)
					isLongRoon = true
					log.Info(
						checkCtx, "Roon检查100次依旧没有播放检查轮训放大到60秒",
						zap.Uint32("共计上传歌曲标记", pushCount.Load()),
					)
				}
				if isLongRoon {
					log.Info(checkCtx, "Roon60秒检查", zap.Uint32("共计上传歌曲标记", pushCount.Load()))
				}
				playing, err := exec.GetMRMediaNowPlaying()
				if err != nil {
					log.Warn(checkCtx, "RoonTrackUpdateNowPlaying", zap.Error(err))
					return
				}
				var roonTrackInfo *exec.MRMediaNowPlaying
				switch playing.BundleIdentifier {
				case exec.MRMediaNowPlayingAppRoon:
					roonTrackInfo = nil
					if playing.IsPlaying {
						if tmpCount > checkCount {
							isLongRoon = false
							timer.Reset(time.Second * defaultSleep)
						}
						tmpCount = 0
						roonTrackInfo = playing
					} else {
						if _, ok := currentPlayingCache.Load(cRoon); ok {
							currentPlayingCache.Delete(cRoon)
							_, aok1 := currentPlayingCache.Load(cAudirvana)
							_, aok2 := currentPlayingCache.Load(cAppleMusic)
							if !aok1 && !aok2 {
								websocket.BroadcastMessage(
									checkCtx,
									&websocket.WsTrackInfo{
										Type:   "stop",
										Source: cRoon,
									},
								)
								atomicPlaying.Store(false)
							}
						}
					}
				}
				if roonTrackInfo != nil {
					tmpTrack := roonTrackInfo.Title //
					currentTrack = tmpTrack
					position := roonTrackInfo.ElapsedTime
					duration := roonTrackInfo.Duration

					// 将播放信息写入本地缓存
					wti := &websocket.WsTrackInfo{
						Type:   "now_playing",
						Source: "roon",
						Data: struct {
							Title  string `json:"title"`
							Album  string `json:"album"`
							Artist string `json:"artist"`
						}{
							roonTrackInfo.Title,
							roonTrackInfo.Album,
							roonTrackInfo.Artist,
						},
					}
					// 向WebSocket客户端广播播放信息
					currentPlayingCache.Store(cRoon, wti)
					atomicPlaying.Store(true)
					websocket.BroadcastMessage(
						checkCtx,
						wti,
					)
					if position/float64(duration) > percentScrobble && !mapedRoon[currentTrack] {
						// 标记听歌完成
						pushTrackScrobbleReq := &lastfm.PushTrackScrobbleReq{
							Artist:      roonTrackInfo.Artist,
							AlbumArtist: roonTrackInfo.Artist,
							Track:       roonTrackInfo.Title,
							Album:       roonTrackInfo.Album,
							Duration:    int64(roonTrackInfo.Duration),
							Timestamp:   now.UTC().Unix(),
						}
						// Save to database
						record := &model.TrackPlayRecord{
							Artist:        pushTrackScrobbleReq.Artist,
							AlbumArtist:   pushTrackScrobbleReq.AlbumArtist,
							Track:         pushTrackScrobbleReq.Track,
							Album:         pushTrackScrobbleReq.Album,
							Duration:      pushTrackScrobbleReq.Duration,
							PlayTime:      time.Unix(pushTrackScrobbleReq.Timestamp, 0),
							Scrobbled:     true,
							MusicBrainzID: pushTrackScrobbleReq.MusicBrainzTrackID,
							TrackNumber:   pushTrackScrobbleReq.TrackNumber,
							Source:        "Roon",
						}
						if err := newTrackService.InsertTrackPlayRecord(checkCtx, record); err != nil {
							log.Warn(checkCtx, "Roon Failed to insert track play record", zap.Error(err))
						}
						// Update track play count
						if err := newTrackService.IncrementTrackPlayCount(
							checkCtx, record.Artist, record.Album, record.Track,
						); err != nil {
							log.Warn(checkCtx, "Roon Failed to increment track play count", zap.Error(err))
						}
						// 说明在听歌存在有效数据的
						_, err := lastfm.PushTrackScrobble(checkCtx, pushTrackScrobbleReq)
						if err != nil {
							log.Warn(checkCtx, "RoonCheckPlayingTrack TrackUpdateNowPlaying", zap.Error(err))
						}
						mapedRoon[currentTrack] = true
						pushCount.Add(1)
						log.Info(
							checkCtx, "RoonCheckPlayingTrack 标记听歌完成",
							zap.String("track", pushTrackScrobbleReq.Track),
						)
					}
					// 上传听歌ing
					if currentTrack != previousTrack {
						// 产生新歌曲
						delete(mapedRoon, previousTrack)
						now = time.Now()
						playingReq := lastfm.TrackUpdateNowPlayingReq{
							Artist:      roonTrackInfo.Artist,
							AlbumArtist: roonTrackInfo.Artist,
							Track:       roonTrackInfo.Title,
							Album:       roonTrackInfo.Album,
							Duration:    int64(roonTrackInfo.Duration),
						}
						log.Info(
							checkCtx, "RoonCheckPlayingTrack NowPlayingTrackInfo",
							zap.Any("roonTrackInfo", roonTrackInfo),
						)
						err := lastfm.TrackUpdateNowPlaying(checkCtx, &playingReq)
						if err != nil {
							log.Warn(ctx, "RoonCheckPlayingTrack TrackUpdateNowPlaying", zap.Error(err))
						}
					}
					previousTrack = tmpTrack // 防止cue文件出现问题
				}
			}
			h(ctx)
		case <-stop:
			fmt.Println("RoonCheckPlayingTrack check playing track exit")
			return
		}
	}
}

// AppleMusicCheckPlayingTrack 检查Apple Music播放状态并上报
func AppleMusicCheckPlayingTrack(ctx context.Context, stop <-chan struct{}) {
	timer := time.NewTicker(time.Second * defaultSleep)
	var (
		now           time.Time
		currentTrack  = ""
		previousTrack = ""
		tmpCount      = 0
	)
	for {
		select {
		case <-timer.C:
			h := func(ctx context.Context) {
				// Start a new span for this check cycle
				checkCtx, span := telemetry.StartSpanForTracerName(
					ctx, _TracerName, "appleMusicCheckPlayingTrack",
				)
				defer span.End()

				log.Debug(checkCtx, "AppleMusicCheckPlayingTrack Checking playing track..."+time.Now().String())
				tmpCount++
				if tmpCount > checkCount && !isLongAppleMusic { // 检查100次依旧没有播放检查轮训放大到60秒
					timer.Reset(time.Second * longSleep)
					isLongAppleMusic = true
					log.Info(
						checkCtx, "AppleMusic检查100次依旧没有播放检查轮训放大到60秒",
						zap.Uint32("共计上传歌曲标记", pushCount.Load()),
					)
				}
				if isLongAppleMusic {
					log.Info(checkCtx, "60秒检查", zap.Uint32("共计上传歌曲标记", pushCount.Load()))
				}
				playing, err := exec.GetMRMediaNowPlaying()
				if err != nil {
					log.Warn(checkCtx, "AppleMusicTrackUpdateNowPlaying", zap.Error(err))
					return
				}
				var appleMusicTrackInfo *exec.MRMediaNowPlaying
				if playing.BundleIdentifier == exec.MRMediaNowPlayingAppMusic {
					appleMusicTrackInfo = nil
					if playing.IsPlaying {
						if tmpCount > checkCount {
							isLongAppleMusic = false
							timer.Reset(time.Second * defaultSleep)
						}
						tmpCount = 0
						appleMusicTrackInfo = playing
					} else {
						if _, ok := currentPlayingCache.Load(cAppleMusic); ok {
							currentPlayingCache.Delete(cAppleMusic)
							_, aok1 := currentPlayingCache.Load(cAudirvana)
							_, aok2 := currentPlayingCache.Load(cRoon)
							if !aok1 && !aok2 {
								websocket.BroadcastMessage(
									checkCtx,
									&websocket.WsTrackInfo{
										Type:   "stop",
										Source: cAppleMusic,
									},
								)
								atomicPlaying.Store(false)
							}
						}
					}
				}

				if appleMusicTrackInfo != nil {
					log.Debug(
						checkCtx, "AppleMusicCheckPlayingTrack 获取播放信息",
						zap.Any("track", appleMusicTrackInfo),
					)
					tmpTrack := appleMusicTrackInfo.Title //
					currentTrack = tmpTrack
					position := appleMusicTrackInfo.ElapsedTime
					duration := appleMusicTrackInfo.Duration

					// 将播放信息写入本地缓存
					wti := &websocket.WsTrackInfo{
						Type:   "now_playing",
						Source: cAppleMusic,
						Data: struct {
							Title  string `json:"title"`
							Album  string `json:"album"`
							Artist string `json:"artist"`
						}{
							appleMusicTrackInfo.Title,
							appleMusicTrackInfo.Album,
							appleMusicTrackInfo.Artist,
						},
					}
					// 向WebSocket客户端广播播放信息
					currentPlayingCache.Store(cAppleMusic, wti)
					atomicPlaying.Store(true)
					websocket.BroadcastMessage(
						checkCtx,
						wti,
					)
					if position/float64(duration) > percentScrobble && !mapedAppleMusic[currentTrack] {
						log.Debug(
							checkCtx, "AppleMusicCheckPlayingTrack 标记听歌完成开始中",
							zap.Any("track", appleMusicTrackInfo),
						)
						// 标记听歌完成
						pushTrackScrobbleReq := &lastfm.PushTrackScrobbleReq{
							Artist:      appleMusicTrackInfo.Artist,
							AlbumArtist: appleMusicTrackInfo.Artist,
							Track:       appleMusicTrackInfo.Title,
							Album:       appleMusicTrackInfo.Album,
							Duration:    int64(appleMusicTrackInfo.Duration),
							Timestamp:   now.UTC().Unix(),
						}
						// Save to database
						record := &model.TrackPlayRecord{
							Artist:        pushTrackScrobbleReq.Artist,
							AlbumArtist:   pushTrackScrobbleReq.AlbumArtist,
							Track:         pushTrackScrobbleReq.Track,
							Album:         pushTrackScrobbleReq.Album,
							Duration:      pushTrackScrobbleReq.Duration,
							PlayTime:      time.Unix(pushTrackScrobbleReq.Timestamp, 0),
							Scrobbled:     true,
							MusicBrainzID: pushTrackScrobbleReq.MusicBrainzTrackID,
							TrackNumber:   pushTrackScrobbleReq.TrackNumber,
							Source:        "Apple Music",
						}
						if err := newTrackService.InsertTrackPlayRecord(checkCtx, record); err != nil {
							log.Warn(checkCtx, "AppleMusic Failed to insert track play record", zap.Error(err))
						}
						// Update track play count
						if err := newTrackService.IncrementTrackPlayCount(
							checkCtx, record.Artist, record.Album, record.Track,
						); err != nil {
							log.Warn(checkCtx, "AppleMusic Failed to increment track play count", zap.Error(err))
						}
						// 说明在听歌存在有效数据的
						_, err := lastfm.PushTrackScrobble(checkCtx, pushTrackScrobbleReq)
						if err != nil {
							log.Warn(checkCtx, "AppleMusicCheckPlayingTrack TrackUpdateNowPlaying err", zap.Error(err))
						}
						mapedAppleMusic[currentTrack] = true
						pushCount.Add(1)
						log.Info(
							checkCtx, "AppleMusicCheckPlayingTrack 标记听歌完成",
							zap.String("track", pushTrackScrobbleReq.Track),
						)
					}
					// 上传听歌ing
					if currentTrack != previousTrack {
						// 产生新歌曲
						delete(mapedAppleMusic, previousTrack)
						now = time.Now()
						playingReq := lastfm.TrackUpdateNowPlayingReq{
							Artist:      appleMusicTrackInfo.Artist,
							AlbumArtist: appleMusicTrackInfo.Artist,
							Track:       appleMusicTrackInfo.Title,
							Album:       appleMusicTrackInfo.Album,
							Duration:    int64(appleMusicTrackInfo.Duration),
						}
						log.Info(
							checkCtx, "AppleMusicCheckPlayingTrack NowPlayingTrackInfo",
							zap.Any("appleMusicTrackInfo", appleMusicTrackInfo),
						)
						err := lastfm.TrackUpdateNowPlaying(checkCtx, &playingReq)
						if err != nil {
							log.Warn(ctx, "AppleMusicCheckPlayingTrack TrackUpdateNowPlaying err", zap.Error(err))
						}
					}
					previousTrack = tmpTrack // 防止cue文件出现问题
				}
			}
			h(ctx)
		case <-stop:
			fmt.Println("AppleMusicCheckPlayingTrack check playing track exit")
			return
		}
	}
}

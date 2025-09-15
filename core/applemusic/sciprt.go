package applemusic

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/vincenty1ung/lastfm-scrobbler/common"
	"github.com/vincenty1ung/lastfm-scrobbler/core/applesciprt"
	alog "github.com/vincenty1ung/lastfm-scrobbler/core/log"
	"github.com/vincenty1ung/lastfm-scrobbler/internal/model"
)

type (
	TrackBase struct {
		// media-control get -h  --now
		TrackID    string
		Title      string `json:"title"`
		Album      string `json:"album"`
		Artist     string `json:"artist"`
		Duration   int64  `json:"duration"`
		Position   float64
		Url        string
		AirfoiLogo string

		ElapsedTimeNow float64 `json:"elapsedTimeNow"` // 当前播放时间
		Genre          string  `json:"genre"`
		TrackNumber    int     `json:"trackNumber"`
		IsMusicApp     bool    `json:"isMusicApp"`
		Playing        bool    `json:"playing"` // 是否正在播放

		Composer         string `json:"composer"`         // 作曲家
		BundleIdentifier string `json:"bundleIdentifier"` // 软件标识

		ElapsedTime           float64   `json:"elapsedTime"`
		ArtworkData           string    `json:"artworkData"`
		UniqueIdentifier      int64     `json:"uniqueIdentifier"`
		ContentItemIdentifier string    `json:"contentItemIdentifier"` // 疑似歌曲id
		RepeatMode            int       `json:"repeatMode"`
		QueueIndex            int       `json:"queueIndex"`
		ArtworkMimeType       string    `json:"artworkMimeType"`
		MediaType             string    `json:"mediaType"`
		Timestamp             time.Time `json:"timestamp"`
		ShuffleMode           int       `json:"shuffleMode"`
		TotalTrackCount       int       `json:"totalTrackCount"`
		ProcessIdentifier     int       `json:"processIdentifier"`
		TotalQueueCount       int       `json:"totalQueueCount"`
		PlaybackRate          int       `json:"playbackRate"`

		// Apple Music specific fields
		AlbumArtist      string    `json:"album_artist"`      // 专辑艺术家
		AlbumDisliked    bool      `json:"album_disliked"`    // 专辑是否被讨厌
		AlbumFavorited   bool      `json:"album_favorited"`   // 专辑是否被收藏
		AlbumRating      int       `json:"album_rating"`      // 专辑评分
		BitRate          int       `json:"bit_rate"`          // 比特率
		Bookmark         float64   `json:"bookmark"`          // 书签时间
		BPM              int       `json:"bpm"`               // 每分钟节拍数
		Category         string    `json:"category"`          // 类别
		Comment          string    `json:"comment"`           // 备注
		Compilation      bool      `json:"compilation"`       // 是否为合辑
		DatabaseID       int       `json:"database_id"`       // 数据库ID
		DateAdded        time.Time `json:"date_added"`        // 添加日期
		Description      string    `json:"description"`       // 描述
		DiscCount        int       `json:"disc_count"`        // 光盘总数
		DiscNumber       int       `json:"disc_number"`       // 光盘编号
		Disliked         bool      `json:"disliked"`          // 是否被讨厌
		DurationString   string    `json:"duration_string"`   // 持续时间字符串格式
		Enabled          bool      `json:"enabled"`           // 是否启用播放
		EQ               string    `json:"eq"`                // 均衡器预设
		Finish           float64   `json:"finish"`            // 结束时间
		Gapless          bool      `json:"gapless"`           // 是否为无缝专辑
		Grouping         string    `json:"grouping"`          // 分组
		Kind             string    `json:"kind"`              // 类型描述
		LongDescription  string    `json:"long_description"`  // 长描述
		Favorited        bool      `json:"favorited"`         // 是否被收藏
		Lyrics           string    `json:"lyrics"`            // 歌词
		ModificationDate time.Time `json:"modification_date"` // 修改日期
		Movement         string    `json:"movement"`          // 运动名称
		MovementCount    int       `json:"movement_count"`    // 运动总数
		MovementNumber   int       `json:"movement_number"`   // 运动编号
		PlayedCount      int       `json:"played_count"`      // 播放次数
		PlayedDate       time.Time `json:"played_date"`       // 最后播放日期
		Rating           int       `json:"rating"`            // 评分
		ReleaseDate      time.Time `json:"release_date"`      // 发布日期
		SampleRate       int       `json:"sample_rate"`       // 采样率
		Shufflable       bool      `json:"shufflable"`        // 是否可随机播放
		SkippedCount     int       `json:"skipped_count"`     // 跳过次数
		SkippedDate      time.Time `json:"skipped_date"`      // 最后跳过日期
		SortAlbum        string    `json:"sort_album"`        // 排序专辑名
		SortArtist       string    `json:"sort_artist"`       // 排序艺术家名
		SortAlbumArtist  string    `json:"sort_album_artist"` // 排序专辑艺术家名
		SortName         string    `json:"sort_name"`         // 排序名称
		SortComposer     string    `json:"sort_composer"`     // 排序作曲家名
		Size             int64     `json:"size"`              // 大小(字节)
		Start            float64   `json:"start"`             // 开始时间
		TrackCount       int       `json:"track_count"`       // 音轨总数
		Unplayed         bool      `json:"unplayed"`          // 是否未播放
		VolumeAdjustment int       `json:"volume_adjustment"` // 音量调整
		Work             string    `json:"work"`              // 作品名
		Year             int       `json:"year"`              // 年份
	}

	TrackInfo struct {
		TrackBase
	}
)

// ToTrackMetadata converts TrackInfo to TrackMetadata for database storage
func (ti *TrackInfo) ToTrackMetadata() model.TrackMetadata {
	return model.TrackMetadata{
		AlbumArtist:   ti.AlbumArtist,
		TrackNumber:   int64(ti.TrackNumber),
		Duration:      ti.Duration,
		Genre:         ti.Genre,
		Composer:      ti.Composer,
		ReleaseDate:   ti.ReleaseDate.Format("2006-01-02"),
		MusicBrainzID: "", // Apple Music doesn't provide this directly
		Source:        "Apple Music",
		BundleID:      ti.BundleIdentifier,
		UniqueID:      fmt.Sprintf("%d", ti.DatabaseID),
	}
}

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
	switch result {
	case "playing":
		return common.PlayerStatePlaying, nil
	case "paused":
		return common.PlayerStatePaused, nil
	default:
		return common.PlayerState(result), nil
	}
}

// GetNowPlayingTrackInfo 使用 AppleScript 从 Apple Music 获取当前正在播放的曲目信息。
// 它返回一个指向 TrackInfo 结构体的指针，包含曲目的详细信息。
// 如果在执行 AppleScript 或解析数据时发生错误，函数会记录警告并返回 nil。
func GetNowPlayingTrackInfo(ctx context.Context) *TrackInfo {
	// 首先检查是否正在播放
	state, err := GetState(ctx)
	if err != nil || state != common.PlayerStatePlaying {
		// 如果没有播放或获取状态出错，返回nil
		return nil
	}

	// 使用更简洁的AppleScript代码获取所有相关信息
	tell, err := applesciprt.Tell(
		"Music",
		`try
			if player state is playing then
				if exists current track then
					set t to current track
					set trackInfo to {name:(name of t), album:(album of t), artist:(artist of t), albumArtist:(album artist of t), duration:(duration of t), playerPosition:(player position), databaseID:(database ID of t), composer:(composer of t), albumDisliked:(album disliked of t), albumFavorited:(album favorited of t), albumRating:(album rating of t), bitRate:(bit rate of t), bookmark:(bookmark of t), bpm:(bpm of t), category:(category of t), comment:(comment of t), compilation:(compilation of t), dateAdded:((date added of t) as string), description:(description of t), discCount:(disc count of t), discNumber:(disc number of t), disliked:(disliked of t), enabled:(enabled of t), eq:(EQ of t), finish:(finish of t), gapless:(gapless of t), genre:(genre of t), grouping:(grouping of t), kind:(kind of t), longDescription:(long description of t), favorited:(favorited of t), lyrics:(lyrics of t), modificationDate:((modification date of t) as string), movement:(movement of t), movementCount:(movement count of t), movementNumber:(movement number of t), playedCount:(played count of t), playedDate:((played date of t) as string), rating:(rating of t), releaseDate:((release date of t) as string), sampleRate:(sample rate of t), shufflable:(shufflable of t), skippedCount:(skipped count of t), skippedDate:((skipped date of t) as string), sortAlbum:(sort album of t), sortArtist:(sort artist of t), sortAlbumArtist:(sort album artist of t), sortName:(sort name of t), sortComposer:(sort composer of t), size:(size of t), start:(start of t), trackCount:(track count of t), trackNumber:(track number of t), unplayed:(unplayed of t), volumeAdjustment:(volume adjustment of t), work:(work of t), year:(year of t)}
					return trackInfo
				end if
			end if
		on error errMsg
			return "error:" & errMsg
		end try`,
	)
	if err != nil {
		alog.Warn(ctx, "AppleScript execution error:", zap.Error(err))
		return nil
	}

	// 检查是否有错误
	if strings.HasPrefix(tell, "error:") {
		alog.Warn(ctx, "AppleScript runtime error:", zap.String("response", tell))
		return nil
	}

	// 如果没有返回数据，说明没有播放曲目
	if tell == "" {
		return nil
	}

	// 解析AppleScript返回的记录格式数据
	// 格式类似于: {name:Song Title, album:Album Name, artist:Artist Name, ...}
	info := &TrackInfo{
		TrackBase: TrackBase{
			IsMusicApp: true,
		},
	}

	// 移除首尾的大括号
	tell = strings.TrimPrefix(tell, "{")
	tell = strings.TrimSuffix(tell, "}")

	// 按逗号分割各个字段
	fields := strings.Split(tell, ", ")

	// 解析每个字段
	for _, field := range fields {
		// 查找冒号位置
		colonIndex := strings.Index(field, ":")
		if colonIndex <= 0 || colonIndex >= len(field)-1 {
			continue
		}

		// 提取键和值
		key := strings.TrimSpace(field[:colonIndex])
		value := strings.TrimSpace(field[colonIndex+1:])

		// 处理"missing value"的特殊情况
		if value == "missing value" {
			continue
		}

		// 根据键设置相应的字段
		switch key {
		case "name":
			info.Title = strings.Trim(value, "\"")
		case "album":
			info.Album = strings.Trim(value, "\"")
		case "artist":
			info.Artist = strings.Trim(value, "\"")
		case "albumArtist":
			info.AlbumArtist = strings.Trim(value, "\"")
		case "duration":
			if num, err := strconv.ParseFloat(value, 64); err == nil {
				info.Duration = int64(num)
			}
		case "playerPosition":
			if num, err := strconv.ParseFloat(value, 64); err == nil {
				info.Position = num
			}
		case "databaseID":
			if num, err := strconv.ParseInt(value, 10, 64); err == nil {
				info.Url = fmt.Sprintf("%d", num)
				info.DatabaseID = int(num)
			}
		case "composer":
			info.Composer = strings.Trim(value, "\"")
		case "albumDisliked":
			if b, err := strconv.ParseBool(value); err == nil {
				info.AlbumDisliked = b
			}
		case "albumFavorited":
			if b, err := strconv.ParseBool(value); err == nil {
				info.AlbumFavorited = b
			}
		case "albumRating":
			if num, err := strconv.Atoi(value); err == nil {
				info.AlbumRating = num
			}
		case "bitRate":
			if num, err := strconv.Atoi(value); err == nil {
				info.BitRate = num
			}
		case "bookmark":
			if num, err := strconv.ParseFloat(value, 64); err == nil {
				info.Bookmark = num
			}
		case "bpm":
			if num, err := strconv.Atoi(value); err == nil {
				info.BPM = num
			}
		case "category":
			info.Category = strings.Trim(value, "\"")
		case "comment":
			info.Comment = strings.Trim(value, "\"")
		case "compilation":
			if b, err := strconv.ParseBool(value); err == nil {
				info.Compilation = b
			}
		case "dateAdded":
			// 尝试不同的日期格式
			if t, err := common.ParseChineseTime(value); err == nil {
				info.DateAdded = t
			} else if t, err := time.Parse("2006-01-02 15:04:05 -0700 MST", strings.Trim(value, "\"")); err == nil {
				info.DateAdded = t
			}
		case "description":
			info.Description = strings.Trim(value, "\"")
		case "discCount":
			if num, err := strconv.Atoi(value); err == nil {
				info.DiscCount = num
			}
		case "discNumber":
			if num, err := strconv.Atoi(value); err == nil {
				info.DiscNumber = num
			}
		case "disliked":
			if b, err := strconv.ParseBool(value); err == nil {
				info.Disliked = b
			}
		case "enabled":
			if b, err := strconv.ParseBool(value); err == nil {
				info.Enabled = b
			}
		case "eq":
			info.EQ = strings.Trim(value, "\"")
		case "finish":
			if num, err := strconv.ParseFloat(value, 64); err == nil {
				info.Finish = num
			}
		case "gapless":
			if b, err := strconv.ParseBool(value); err == nil {
				info.Gapless = b
			}
		case "genre":
			info.Genre = common.ConversionSimplifiedFx(strings.Trim(value, "\""))
		case "grouping":
			info.Grouping = strings.Trim(value, "\"")
		case "kind":
			info.Kind = strings.Trim(value, "\"")
		case "longDescription":
			info.LongDescription = strings.Trim(value, "\"")
		case "favorited":
			if b, err := strconv.ParseBool(value); err == nil {
				info.Favorited = b
			}
		case "lyrics":
			info.Lyrics = strings.Trim(value, "\"")
		case "modificationDate":
			// 尝试不同的日期格式
			if t, err := common.ParseChineseTime(value); err == nil {
				info.ModificationDate = t
			} else if t, err := time.Parse("2006-01-02 15:04:05 -0700 MST", strings.Trim(value, "\"")); err == nil {
				info.ModificationDate = t
			}
		case "movement":
			info.Movement = strings.Trim(value, "\"")
		case "movementCount":
			if num, err := strconv.Atoi(value); err == nil {
				info.MovementCount = num
			}
		case "movementNumber":
			if num, err := strconv.Atoi(value); err == nil {
				info.MovementNumber = num
			}
		case "playedCount":
			if num, err := strconv.Atoi(value); err == nil {
				info.PlayedCount = num
			}
		case "playedDate":
			// 尝试不同的日期格式
			if t, err := common.ParseChineseTime(value); err == nil {
				info.PlayedDate = t
			} else if t, err := time.Parse("2006-01-02 15:04:05 -0700 MST", strings.Trim(value, "\"")); err == nil {
				info.PlayedDate = t
			}
		case "rating":
			if num, err := strconv.Atoi(value); err == nil {
				info.Rating = num
			}
		case "releaseDate":
			// 尝试不同的日期格式
			if t, err := common.ParseChineseTime(value); err == nil {
				info.ReleaseDate = t
			} else if t, err := time.Parse("2006-01-02 15:04:05 -0700 MST", strings.Trim(value, "\"")); err == nil {
				info.ReleaseDate = t
			}
		case "sampleRate":
			if num, err := strconv.Atoi(value); err == nil {
				info.SampleRate = num
			}
		case "shufflable":
			if b, err := strconv.ParseBool(value); err == nil {
				info.Shufflable = b
			}
		case "skippedCount":
			if num, err := strconv.Atoi(value); err == nil {
				info.SkippedCount = num
			}
		case "skippedDate":
			// 尝试不同的日期格式
			if t, err := common.ParseChineseTime(value); err == nil {
				info.SkippedDate = t
			} else if t, err := time.Parse("2006-01-02 15:04:05 -0700 MST", strings.Trim(value, "\"")); err == nil {
				info.SkippedDate = t
			}
		case "sortAlbum":
			info.SortAlbum = strings.Trim(value, "\"")
		case "sortArtist":
			info.SortArtist = strings.Trim(value, "\"")
		case "sortAlbumArtist":
			info.SortAlbumArtist = strings.Trim(value, "\"")
		case "sortName":
			info.SortName = strings.Trim(value, "\"")
		case "sortComposer":
			info.SortComposer = strings.Trim(value, "\"")
		case "size":
			if num, err := strconv.ParseInt(value, 10, 64); err == nil {
				info.Size = num
			}
		case "start":
			if num, err := strconv.ParseFloat(value, 64); err == nil {
				info.Start = num
			}
		case "trackCount":
			if num, err := strconv.Atoi(value); err == nil {
				info.TrackCount = num
			}
		case "trackNumber":
			if num, err := strconv.Atoi(value); err == nil {
				info.TrackNumber = num
			}
		case "unplayed":
			if b, err := strconv.ParseBool(value); err == nil {
				info.Unplayed = b
			}
		case "volumeAdjustment":
			if num, err := strconv.Atoi(value); err == nil {
				info.VolumeAdjustment = num
			}
		case "work":
			info.Work = strings.Trim(value, "\"")
		case "year":
			if num, err := strconv.Atoi(value); err == nil {
				info.Year = num
			}
		}
	}

	return info
}

// IsFavorite checks if the current track is favorited in Apple Music
func IsFavorite(ctx context.Context) (bool, error) {
	tell, err := applesciprt.Tell(
		"Music",
		`if exists current track then
		return favorited of current track
	end if`,
	)
	if err != nil {
		alog.Warn(ctx, "err:", zap.Error(err))
		return false, err
	}

	parseBool, err := strconv.ParseBool(tell)
	if err != nil {
		alog.Warn(ctx, "err:", zap.Error(err))
		return false, err
	}
	return parseBool, nil
}

// SetFavorite sets the favorited status of the current track in Apple Music
func SetFavorite(ctx context.Context, favorited bool) error {
	alog.Debug(ctx, "apple music. Track love status:", zap.Bool("favorited", favorited))
	_, err := applesciprt.Tell(
		"Music",
		fmt.Sprintf(`set favorited of current track to %s`, strconv.FormatBool(favorited)),
	)
	if err != nil {
		alog.Warn(ctx, "err:", zap.Error(err))
		return err
	}
	alog.Info(ctx, "apple music. Track loved successfully")
	return nil
}

package api

import (
	"context"
	"html/template"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"

	"github.com/vincenty1ung/lastfm-scrobbler/config"
	"github.com/vincenty1ung/lastfm-scrobbler/core/log"
	"github.com/vincenty1ung/lastfm-scrobbler/core/websocket"
	"github.com/vincenty1ung/lastfm-scrobbler/internal/logic/analysis"
	"github.com/vincenty1ung/lastfm-scrobbler/internal/logic/genre"
	"github.com/vincenty1ung/lastfm-scrobbler/internal/logic/track"
	"github.com/vincenty1ung/lastfm-scrobbler/internal/model"
)

func setupRouter(name string) *gin.Engine {
	r := gin.Default()

	// Add OpenTelemetry middleware
	r.Use(
		otelgin.Middleware(name),
		func(c *gin.Context) {
			traceID := trace.SpanFromContext(c.Request.Context()).SpanContext().TraceID().String()
			c.Header("Trace-Id", traceID)
			c.Next()
		},
	)

	// 首页
	r.GET(
		"/", func(c *gin.Context) {
			// Load HTML template
			tmplPath := filepath.Join("templates", "dashboard.html")
			tmpl, err := template.New("dashboard.html").ParseFiles(tmplPath)
			if err != nil {
				log.Error(c.Request.Context(), "Failed to parse template", zap.Error(err))
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load template"})
				return
			}

			// Set content type and write HTML response
			c.Header("Content-Type", "text/html; charset=utf-8")
			if err := tmpl.Execute(c.Writer, nil); err != nil {
				log.Error(c.Request.Context(), "Failed to execute template", zap.Error(err))
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to render template"})
				return
			}
		},
	)

	// Get track play counts with pagination
	trackService := track.NewTrackService()
	r.GET(
		"/api/track-play-counts", func(c *gin.Context) {
			limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
			offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

			if limit > 100 {
				limit = 100 // Limit max records per page
			}

			records, err := trackService.GetTrackPlayCounts(c.Request.Context(), limit, offset)
			log.Info(
				c.Request.Context(), "Fetched track play counts", zap.Int("count", len(records)),
				zap.Int("limit", limit), zap.Int("offset", offset),
			)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			// Check if client expects HTML response
			acceptHeader := c.GetHeader("Accept")
			if strings.Contains(acceptHeader, "text/html") || c.Query("format") == "html" {
				// Load HTML template
				tmplPath := filepath.Join("templates", "track_play_counts.html")
				tmpl, err := template.New("track_play_counts.html").Funcs(
					template.FuncMap{
						"addOne": func(i int) int {
							return i + 1
						},
						"add": func(a, b int) int {
							return a + b
						},
						"subtract": func(a, b int) int {
							return a - b
						},
					},
				).ParseFiles(tmplPath)
				if err != nil {
					log.Error(c.Request.Context(), "Failed to parse template", zap.Error(err))
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load template"})
					return
				}

				// Execute template with records data
				data := struct {
					Records     []*model.Track
					Limit       int
					Offset      int
					RecordCount int
				}{
					Records:     records,
					Limit:       limit,
					Offset:      offset,
					RecordCount: len(records),
				}

				// Set content type and write HTML response
				c.Header("Content-Type", "text/html; charset=utf-8")
				if err := tmpl.Execute(c.Writer, data); err != nil {
					log.Error(c.Request.Context(), "Failed to execute template", zap.Error(err))
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to render template"})
					return
				}
			} else {
				// Return JSON response for API clients
				c.JSON(http.StatusOK, records)
			}
		},
	)

	// Get play for a specific track
	r.GET(
		"/api/track", func(c *gin.Context) {
			artist := c.Query("artist")
			album := c.Query("album")
			trackName := c.Query("trackName")

			if artist == "" || album == "" || trackName == "" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "artist, album, and trackName are required"})
				return
			}

			record, err := trackService.GetTrack(c.Request.Context(), artist, album, trackName)
			if err != nil {
				if err.Error() == "record not found" {
					c.JSON(http.StatusOK, gin.H{"play_count": 0})
					return
				}
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusOK, record)
		},
	)

	// Generate music preference report
	musicAnalysisService := analysis.NewMusicAnalysisService()

	// Generate music recommendations

	// 获取仪表板统计数据
	r.GET(
		"/api/dashboard/stats", func(c *gin.Context) {
			ctx := c.Request.Context()

			// 获取总播放次数
			totalPlays, err := trackService.GetTotalPlayCount(ctx)
			if err != nil {
				log.Error(ctx, "Failed to get total play count", zap.Error(err))
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get total play count"})
				return
			}

			// 获取曲目总数
			totalTracks, err := trackService.GetTrackCounts(ctx)
			if err != nil {
				log.Error(ctx, "Failed to get track counts", zap.Error(err))
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get track counts"})
				return
			}

			// 获取艺术家总数
			totalArtists, err := trackService.GetArtistCounts(ctx)
			if err != nil {
				log.Error(ctx, "Failed to get artist counts", zap.Error(err))
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get artist counts"})
				return
			}

			// 获取专辑总数
			totalAlbums, err := trackService.GetAlbumCounts(ctx)
			if err != nil {
				log.Error(ctx, "Failed to get album counts", zap.Error(err))
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get album counts"})
				return
			}

			// 返回统计数据
			stats := gin.H{
				"totalPlays":   totalPlays,
				"totalTracks":  totalTracks,
				"totalArtists": totalArtists,
				"totalAlbums":  totalAlbums,
			}

			c.JSON(http.StatusOK, stats)
		},
	)

	// 获取趋势图数据
	r.GET(
		"/api/dashboard/trend", func(c *gin.Context) {
			ctx := c.Request.Context()

			// 获取时间范围参数，默认7天
			rangeStr := c.DefaultQuery("range", "7")
			rangeDays := 7
			switch rangeStr {
			case "30":
				rangeDays = 30
			case "90":
				rangeDays = 90
			default:
				rangeDays = 7
			}
			fillInTrendCycle := FillInTrendCycle(rangeDays)

			// 获取指定天数的播放记录
			recordMap, err := trackService.GetRecentPlayRecordsByDays(ctx, rangeDays)
			if err != nil {
				log.Error(ctx, "Failed to get recent play records", zap.Error(err))
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get recent play records"})
				return
			}
			// 处理数据以适应趋势图
			// 按日期统计播放次数
			dateTrendData := make(map[string]int)
			// 按日期和小时统计播放次数
			hourlyTrendData := make(map[string]*model.HourlyPlayTrendData)
			for _, trendCycle := range fillInTrendCycle {
				if records, ok := recordMap[trendCycle]; ok {
					for _, record := range records {
						dateStr := record.PlayTime.Format("2006-01-02")
						hour := record.PlayTime.Hour()
						// 统计每日总播放次数
						dateTrendData[dateStr]++
						// 初始化该日期的小时统计数据
						if _, exists := hourlyTrendData[dateStr]; !exists {
							hourlyTrendData[dateStr] = &model.HourlyPlayTrendData{
								Date:   dateStr,
								Total:  0,
								Hourly: make(map[int]int),
							}
						}

						// 统计该小时的播放次数
						hourlyTrendData[dateStr].Hourly[hour]++
						hourlyTrendData[dateStr].Total++
					}
				} else {
					dateTrendData[trendCycle] = 0
					hourlyTrendData[trendCycle] = &model.HourlyPlayTrendData{
						Date:  trendCycle,
						Total: 0,
						Hourly: map[int]int{
							0: 0,
							1: 0,
							2: 0,
						},
					}
				}
			}

			// 构造返回数据
			result := gin.H{
				"daily":  dateTrendData,
				"hourly": hourlyTrendData,
			}

			c.JSON(http.StatusOK, result)
		},
	)

	// 获取热门艺术家数据（按播放次数）
	r.GET(
		"/api/dashboard/top-artists/plays", func(c *gin.Context) {
			ctx := c.Request.Context()

			// 获取限制参数，默认10个
			limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
			if limit > 50 {
				limit = 50 // 限制最大数量
			}

			// 获取按播放次数统计的热门艺术家
			artists, err := trackService.GetTopArtistsByPlayCount(ctx, limit)
			if err != nil {
				log.Error(ctx, "Failed to get top artists by play count", zap.Error(err))
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get top artists by play count"})
				return
			}

			c.JSON(http.StatusOK, artists)
		},
	)

	// 获取热门艺术家数据（按曲目数）
	r.GET(
		"/api/dashboard/top-artists/tracks", func(c *gin.Context) {
			ctx := c.Request.Context()

			// 获取限制参数，默认10个
			limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
			if limit > 50 {
				limit = 50 // 限制最大数量
			}

			// 获取按曲目数统计的热门艺术家
			artists, err := trackService.GetTopArtistsByTrackCount(ctx, limit)
			if err != nil {
				log.Error(ctx, "Failed to get top artists by track count", zap.Error(err))
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get top artists by track count"})
				return
			}

			c.JSON(http.StatusOK, artists)
		},
	)

	// 分析报告页面
	r.GET(
		"/report", func(c *gin.Context) {
			// Create a background context for the report generation
			ctx := c.Request.Context()

			// Generate the report data
			reportData, err := musicAnalysisService.GenerateMusicPreferenceReport(ctx)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			// Load HTML template
			tmplPath := filepath.Join("templates", "report.html")
			tmpl, err := template.New("report.html").Funcs(
				template.FuncMap{
					"addOne": func(i int) int {
						return i + 1
					},
				},
			).ParseFiles(tmplPath)
			if err != nil {
				log.Error(ctx, "Failed to parse template", zap.Error(err))
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load template"})
				return
			}

			// Set content type and write HTML response
			c.Header("Content-Type", "text/html; charset=utf-8")
			if err := tmpl.Execute(c.Writer, reportData); err != nil {
				log.Error(ctx, "Failed to execute template", zap.Error(err))
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to render template"})
				return
			}
		},
	)

	// 播放统计页面
	r.GET(
		"/playCounts", func(c *gin.Context) {
			limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
			offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

			if limit > 100 {
				limit = 100 // Limit max records per page
			}

			records, err := trackService.GetTrackPlayCounts(c.Request.Context(), limit, offset)
			log.Info(
				c.Request.Context(), "Fetched track play counts", zap.Int("count", len(records)),
				zap.Int("limit", limit), zap.Int("offset", offset),
			)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			// Check if client expects HTML response
			acceptHeader := c.GetHeader("Accept")
			if strings.Contains(acceptHeader, "text/html") || c.Query("format") == "html" {
				// Load HTML template
				tmplPath := filepath.Join("templates", "track_play_counts.html")
				tmpl, err := template.New("track_play_counts.html").Funcs(
					template.FuncMap{
						"addOne": func(i int) int {
							return i + 1
						},
						"add": func(a, b int) int {
							return a + b
						},
						"subtract": func(a, b int) int {
							return a - b
						},
					},
				).ParseFiles(tmplPath)
				if err != nil {
					log.Error(c.Request.Context(), "Failed to parse template", zap.Error(err))
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load template"})
					return
				}

				// Execute template with records data
				data := struct {
					Records     []*model.Track
					Limit       int
					Offset      int
					RecordCount int
				}{
					Records:     records,
					Limit:       limit,
					Offset:      offset,
					RecordCount: len(records),
				}

				// Set content type and write HTML response
				c.Header("Content-Type", "text/html; charset=utf-8")
				if err := tmpl.Execute(c.Writer, data); err != nil {
					log.Error(c.Request.Context(), "Failed to execute template", zap.Error(err))
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to render template"})
					return
				}
			} else {
				// Return JSON response for API clients
				c.JSON(http.StatusOK, records)
			}
		},
	)

	// 音乐推荐页面
	r.GET(
		"/recommendations", func(c *gin.Context) {
			// Create a background context for the recommendation generation
			ctx := c.Request.Context()

			// Generate recommendations
			recommendations, err := musicAnalysisService.GenerateRecommendations(ctx)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			// Load HTML template
			tmplPath := filepath.Join("templates", "recommendations.html")
			tmpl, err := template.New("recommendations.html").Funcs(
				template.FuncMap{
					"addOne": func(i int) int {
						return i + 1
					},
				},
			).ParseFiles(tmplPath)
			if err != nil {
				log.Error(ctx, "Failed to parse template", zap.Error(err))
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load template"})
				return
			}

			// Execute template with recommendations data
			data := struct {
				Recommendations []analysis.MusicRecommendation
			}{
				Recommendations: recommendations,
			}

			// Set content type and write HTML response
			c.Header("Content-Type", "text/html; charset=utf-8")
			if err := tmpl.Execute(c.Writer, data); err != nil {
				log.Error(ctx, "Failed to execute template", zap.Error(err))
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to render template"})
				return
			}
		},
	)

	// 最近播放接口
	r.GET(
		"/api/recent-plays", func(c *gin.Context) {
			limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

			if limit > 100 {
				limit = 100 // Limit max records
			}

			records, err := trackService.GetRecentPlayRecords(c.Request.Context(), limit)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusOK, records)
		},
	)

	// 按时间段获取播放排行榜接口
	r.GET(
		"/api/track-play-counts/period", func(c *gin.Context) {
			limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
			offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
			period := c.Query("period") // 支持 week, month

			if limit > 100 {
				limit = 100 // Limit max records per page
			}

			records, err := trackService.GetTrackPlayCountsByPeriod(c.Request.Context(), limit, offset, period)
			log.Info(
				c.Request.Context(), "Fetched track play counts by period", zap.String("period", period),
				zap.Int("count", len(records)), zap.Int("limit", limit), zap.Int("offset", offset),
			)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusOK, records)
		},
	)

	// 获取按来源统计的播放次数
	r.GET(
		"/api/dashboard/play-counts-by-source", func(c *gin.Context) {
			ctx := c.Request.Context()

			sourceCounts, err := trackService.GetPlayCountsBySource(ctx)
			if err != nil {
				log.Error(ctx, "Failed to get play counts by source", zap.Error(err))
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get play counts by source"})
				return
			}

			c.JSON(http.StatusOK, sourceCounts)
		},
	)

	// 获取热门专辑数据（按播放次数）
	r.GET(
		"/api/dashboard/top-albums", func(c *gin.Context) {
			ctx := c.Request.Context()

			// 获取时间范围参数，默认30天
			daysStr := c.DefaultQuery("days", "30")
			days, err := strconv.Atoi(daysStr)
			if err != nil {
				// 如果无法解析天数，默认使用30天
				days = 30
			}

			// 获取限制参数，默认10个
			limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
			if limit > 50 {
				limit = 50 // 限制最大数量
			}

			// 获取按播放次数统计的热门专辑
			albums, err := trackService.GetTopAlbumsByPlayCount(ctx, days, limit)
			if err != nil {
				log.Error(ctx, "Failed to get top albums by play count", zap.Error(err))
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get top albums by play count"})
				return
			}

			c.JSON(http.StatusOK, albums)
		},
	)

	// 获取热门流派数据（按播放次数和曲目数）
	genreService := genre.NewGenreService()
	r.GET(
		"/api/dashboard/top-genres", func(c *gin.Context) {
			ctx := c.Request.Context()

			// 获取限制参数，默认10个
			limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
			if limit > 50 {
				limit = 50 // 限制最大数量
			}

			// 获取热门流派的详细信息
			genres, err := genreService.GetTopGenresWithDetails(ctx, limit)
			if err != nil {
				log.Error(ctx, "Failed to get top genres with details", zap.Error(err))
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get top genres with details"})
				return
			}

			c.JSON(http.StatusOK, genres)
		},
	)

	// 获取未同步到Last.fm的播放记录（分页）
	r.GET(
		"/api/unscrobbled-records", func(c *gin.Context) {
			ctx := c.Request.Context()

			limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
			offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

			if limit > 100 {
				limit = 100 // Limit max records per page
			}

			records, err := trackService.GetUnscrobbledRecordsWithPagination(ctx, limit, offset)
			if err != nil {
				log.Error(ctx, "Failed to get unscrobbled records", zap.Error(err))
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get unscrobbled records"})
				return
			}

			c.JSON(http.StatusOK, records)
		},
	)

	// 获取未同步到Last.fm的播放记录总数
	r.GET(
		"/api/unscrobbled-records/count", func(c *gin.Context) {
			ctx := c.Request.Context()

			count, err := trackService.GetUnscrobbledRecordsCount(ctx)
			if err != nil {
				log.Error(ctx, "Failed to get unscrobbled records count", zap.Error(err))
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get unscrobbled records count"})
				return
			}

			c.JSON(http.StatusOK, gin.H{"count": count})
		},
	)

	// 同步选中的未同步记录到Last.fm
	r.POST(
		"/api/unscrobbled-records/sync", func(c *gin.Context) {
			ctx := c.Request.Context()

			var req struct {
				IDs []uint `json:"ids"`
			}

			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
				return
			}

			if len(req.IDs) == 0 {
				c.JSON(http.StatusBadRequest, gin.H{"error": "No record IDs provided"})
				return
			}

			// 将 req.IDs 从 []uint 转换为 []int64
			ids := make([]int64, len(req.IDs))
			for i, id := range req.IDs {
				ids[i] = int64(id)
			}

			// 调用logic层方法同步选中的记录
			successCount, failedRecords, err := trackService.SyncSelectedUnscrobbledRecords(ctx, ids)
			if err != nil {
				log.Error(ctx, "Failed to sync selected unscrobbled records", zap.Error(err))
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to sync records"})
				return
			}

			c.JSON(
				http.StatusOK, gin.H{
					"success_count":  successCount,
					"failed_count":   len(failedRecords),
					"failed_records": failedRecords,
				},
			)
		},
	)

	// 处理收藏请求
	r.POST(
		"/api/favorite", func(c *gin.Context) {
			ctx := c.Request.Context()

			var req struct {
				Artist   string `json:"artist"`
				Album    string `json:"album"`
				Track    string `json:"track"`
				Source   string `json:"source"`
				Favorite bool   `json:"favorite"`
			}

			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
				return
			}

			// 验证必要参数
			if req.Artist == "" || req.Album == "" || req.Track == "" || req.Source == "" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "artist, album, track, and source are required"})
				return
			}

			// 调用logic层方法处理收藏逻辑
			appleMusicFav, lastFmFav, err := trackService.SetTrackFavorite(
				ctx, req.Artist, req.Album, req.Track, req.Source, req.Favorite, model.TrackMetadata{},
			)
			if err != nil {
				log.Error(ctx, "Failed to set track favorite", zap.Error(err))
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to set track favorite"})
				return
			}

			c.JSON(
				http.StatusOK, gin.H{
					"apple_music": appleMusicFav,
					"lastfm":      lastFmFav,
				},
			)
		},
	)

	// WebSocket endpoint
	r.GET(
		"/ws", func(c *gin.Context) {
			// 升级HTTP连接到WebSocket连接
			conn, err := websocket.UpgradeConnection(c.Writer, c.Request)
			if err != nil {
				log.Error(c.Request.Context(), "Failed to upgrade to WebSocket", zap.Error(err))
				return
			}

			// 添加连接到连接池
			websocket.AddClient(conn)

			// 启动goroutine处理WebSocket消息
			go websocket.HandleWebSocketMessages(conn)
		},
	)

	// Health check endpoint
	r.GET(
		"/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		},
	)

	return r
}

func StartHTTPServer(ctx context.Context, name string) {
	r := setupRouter(name)
	port := config.ConfigObj.HTTP.Port
	if port == "" {
		port = "8080" // Default port
	}
	log.Info(ctx, "Starting HTTP server on port", zap.String("port", port))
	err := r.Run(":" + port)
	if err != nil {
		panic(err)
	}
}

// FillInTrendCycle FillInTrendCycle
func FillInTrendCycle(rangeDays int) []string {
	now := time.Now()
	rangeDayList := make([]string, 0, rangeDays)
	rangeDayList = append(rangeDayList, now.AddDate(0, 0, -rangeDays).Format("2006-01-02"))
	start := now.AddDate(0, 0, -rangeDays)
	for start.Before(now) {
		start = start.AddDate(0, 0, 1)
		rangeDayList = append(rangeDayList, start.Format("2006-01-02"))
	}
	rangeDayList = append(rangeDayList, now.AddDate(0, 0, 1).Format("2006-01-02"))
	return rangeDayList
}

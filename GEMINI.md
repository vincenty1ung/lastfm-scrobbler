# 项目概述

这是一个用 Go 语言编写的音乐播放记录同步工具，主要用于将 Audirvana, Roon 和 Apple Music 播放的音乐曲目同步（Scrobble）到 Last.fm。项目通过定时检查播放状态，获取当前播放曲目的信息，并在满足一定条件（如播放进度达到 55%）时将曲目信息上报到 Last.fm。

除了同步功能，该项目还提供了丰富的功能，包括：

- **本地数据存储**: 使用 SQLite 通过 GORM 实现本地数据持久化，存储播放记录和播放统计。
- **Web 界面**: 提供一个 Web 界面，包括仪表板、报告和推荐功能。
- **播放记录追踪**: 记录每次播放的详细信息，包括艺术家、专辑、曲目、播放时间等。
- **播放统计**: 统计每首曲目的播放次数，使用乐观锁机制保证并发安全。
- **数据同步**: 将未同步到 Last.fm 的播放记录进行同步，并标记同步状态。
- **命令行工具**: 提供多个命令行工具，用于数据同步、音乐分析和内存管理。

## 主要技术栈

- **语言**: Go 1.24
- **依赖管理**: Go Modules
- **Web 框架**: Gin
- **数据库**: SQLite (通过 GORM), MySQL (可选)
- **缓存**: Redis
- **链路跟踪**: OpenTelemetry
- **主要依赖库**:
  - `github.com/spf13/cobra` - 命令行接口
  - `github.com/spf13/viper` - 配置管理
  - `go.uber.org/zap` - 日志记录
  - `github.com/shkh/lastfm-go` - Last.fm API 客户端
  - `github.com/andybrewer/mack` - AppleScript 执行
  - `gorm.io/gorm` - ORM 框架
  - `gorm.io/driver/sqlite` - SQLite 驱动
  - `gorm.io/driver/mysql` - MySQL 驱动
  - `github.com/gorilla/websocket` - WebSocket 支持
  - `github.com/longbridgeapp/opencc` - 简繁体中文转换
  - `go.opentelemetry.io/otel` - OpenTelemetry SDK

## 项目架构

- `main.go`: 程序入口，使用 Cobra 设置命令行参数并启动服务。
- `api/`: 基于 Gin 的 API 服务器，处理 HTTP 请求。
- `cmd/`: 命令行工具的实现。
  - `analysis_cmd.go`: 音乐分析。
  - `memory_tool.go`: 内存管理。
  - `sync_records.go`: 数据同步。
- `common/`: 通用工具函数，如中文转换。
- `config/`: 配置管理模块，使用 Viper 解析 `config.yaml`。
- `core/`: 核心功能模块，如日志、数据库、Redis、遥测等。
- `internal/`: 项目内部逻辑。
  - `cache/`: 缓存逻辑。
  - `logic/`: 业务逻辑，分为 `analysis`, `genre`, `track`。
  - `model/`: 数据模型和数据库操作。
  - `scrobbler/`: 核心同步逻辑。
- `shell/`: 包含用于构建、启动和停止服务的 shell 脚本。
- `static/`: 存放 Web 界面的静态文件（图片等）。
- `templates/`: Web 界面的 HTML 模板。
- `.storage/`: 本地数据存储目录。

## 数据模型

### Track

`Track` 模型存储了关于每个音轨的详细信息，包括播放统计和用户偏好。

- **ID**: 主键
- **Artist**: 艺术家
- **AlbumArtist**: 专辑艺术家
- **Album**: 专辑
- **Track**: 曲目名
- **TrackNumber**: 音轨号
- **Duration**: 持续时间
- **Genre**: 流派
- **Composer**: 作曲家
- **ReleaseDate**: 发布日期
- **MusicBrainzID**: MusicBrainz ID
- **PlayCount**: 播放次数
- **IsAppleMusicFav**: 是否为 Apple Music 喜欢
- **IsLastFmFav**: 是否为 Last.fm 喜欢
- **Source**: 数据来源 (Apple Music, Audirvana, Roon)
- **BundleID**: 应用标识符
- **UniqueID**: 唯一标识符
- **Version**: 乐观锁版本号
- **CreatedAt**: 创建时间
- **UpdatedAt**: 更新时间

### TrackPlayRecord (播放记录)

存储每次播放的详细信息：

- ID: 主键
- Artist: 艺术家
- AlbumArtist: 专辑艺术家
- Track: 曲目名
- Album: 专辑名
- Duration: 持续时间
- PlayTime: 播放时间
- Scrobbled: 是否已同步到 Last.fm
- MusicBrainzID: MusicBrainz ID
- TrackNumber: 音轨号
- Source: 数据来源（Audirvana, Roon, Apple Music）
- CreatedAt: 创建时间
- UpdatedAt: 更新时间

### Genre (流派)

存储音乐流派信息：

- **ID**: 主键
- **Name**: 流派名称 (英文)
- **NameZh**: 流派名称 (中文)
- **Extra**: 额外信息
- **PlayCount**: 播放次数
- **CreatedAt**: 创建时间
- **UpdatedAt**: 更新时间

## Web 界面

项目提供了一个简单的 Web 界面，用于：

- **仪表板**: 显示当前的播放信息。
- **推荐**: 根据播放历史提供音乐推荐。
- **报告**: 生成音乐播放报告。

## 构建和运行

## 配置

在运行程序前，需要配置 `config/config.yaml` 文件，填入 Last.fm 和 Musixmatch 的 API 密钥等信息。

## 构建

```bash
go build
```

## 运行

```bash
./lastfm-scrobbler
```

### 使用脚本运行

项目提供了 shell 脚本来简化构建和运行过程：

```bash
# 构建 launchctl 服务
sh shell/script/build_lastfm-scrobblers_launchctl.sh

# 启动服务
sh shell/script/start_lastfm-scrobblers.sh

# 停止服务
sh shell/script/stop_lastfm-scrobblers.sh
```

## 查看日志

```bash
tail -f .logs/go_lastfm-scrobbler.log
```

# 开发约定

- **代码风格**: 遵循 Go 语言惯用风格。
- **日志**: 使用 `go.uber.org/zap` 进行日志记录，区分不同日志级别。
- **配置**: 使用 `github.com/spf13/viper` 管理配置，配置文件为 `config/config.yaml`。
- **命令行接口**: 使用 `github.com/spf13/cobra` 构建命令行接口。
- **数据库**: 使用 `gorm.io/gorm` 作为 ORM 框架，使用 SQLite 作为本地存储。

## 特性开发与记忆协议

- **特性清单**: 为任何新模块或特性，必须在 `memory/{date}` 目录内创建一个特性清单文件，详细说明其范围、功能和实现要点。文件名应具有描述性，如 `feature_name_feature_manifest.md`。
- **记忆索引**: 创建新特性后，必须在中央 `memory_index.md` 文件中添加一个条目，包含:
  - **日期**: 添加特性的日期。
  - **特性摘要**: 一句话总结新特性。
  - **链接**: 指向该特性的特性清单文件的链接。
- **记忆扩展**: 如果主 `CURSOR.md` 文件变得过于庞大，应为特定领域创建补充的 markdown 文件，并在主文件中链接到它们，以保持清晰。
- **日期分类管理**: 特性清单文件应按创建日期归档到 `memory/{date}` 目录中，以便更好地组织和管理。
package common

import (
	"strings"
)

var (
	AppSystemEvents    = "System Events"
	AppAudirvanaOrigin = "Audirvana Origin"
	FileExtWav1        = ".wav"
	FileExtWav2        = strings.ToUpper(FileExtWav1)
)

type PlayerState string

const (
	PlayerStateDefault = ""
	PlayerStateStopped = "Stopped"
	PlayerStatePlaying = "Playing"
	PlayerStatePaused  = "Paused"
)

// PlayerType 定义播放器类型
type PlayerType string

const (
	PlayerAudirvana  PlayerType = "Audirvana"
	PlayerRoon       PlayerType = "Roon"
	PlayerAppleMusic PlayerType = "Apple Music"
)

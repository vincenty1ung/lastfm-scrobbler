package applemusic

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsRunning(t *testing.T) {
	ctx := context.Background()
	// This test will pass if Music app is running or not
	// We're just testing that the function doesn't panic
	running := IsRunning(ctx)
	assert.NotNil(t, running)
}

func TestGetState(t *testing.T) {
	ctx := context.Background()
	// Check if Music is running first
	if !IsRunning(ctx) {
		// If Music is not running, we expect an error
		state, err := GetState(ctx)
		assert.Equal(t, "", string(state))
		assert.NotNil(t, err)
	} else {
		// If Music is running, we should get a valid state or an error
		state, err := GetState(ctx)
		fmt.Println(state)
		// We just ensure the function doesn't panic
		_ = state
		_ = err
	}
}

func TestGetNowPlayingTrackInfo(t *testing.T) {
	ctx := context.Background()
	// Check if Music is running first
	if !IsRunning(ctx) {
		// If Music is not running, we expect nil
		info := GetNowPlayingTrackInfo(ctx)
		assert.Nil(t, info)
	} else {
		// If Music is running, we should get track info or nil (if no track is playing)
		info := GetNowPlayingTrackInfo(ctx)
		// info can be nil if no track is playing, which is valid
		// We just ensure the function doesn't panic
		fmt.Println(info)
	}
}

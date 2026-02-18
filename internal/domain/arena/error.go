package arena

import "errors"

// ----------------------------
// Errors (domain-level)
// ----------------------------

var (
	ErrInvalidName         = errors.New("invalid arena name")
	ErrArenaNotRunning     = errors.New("arena is not running")
	ErrArenaNotPaused      = errors.New("arena is not paused")
	ErrArenaNotPending     = errors.New("arena is not pending")
	ErrArenaFinished       = errors.New("arena is finished")
	ErrPlayerAlreadyJoined = errors.New("player already joined")
	ErrPlayerNotFound      = errors.New("player not found")
	ErrPermissionDenied    = errors.New("permission denied")
	ErrInvalidAction       = errors.New("invalid action payload")
	ErrApplyAtTickTooOld   = errors.New("applyAtTick must be >= current tick")
)

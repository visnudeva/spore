package audio

import (
	"errors"
	"time"
)

const (
	liveWatchInterval   = 500 * time.Millisecond
	liveReconnectMin    = time.Second
	liveReconnectMax    = 20 * time.Second
	liveReconnectSettle = 2 * time.Second
)

var (
	errPlaybackMoved  = errors.New("playback replaced")
	errPlaybackPaused = errors.New("playback paused")
)

// streamShouldReconnect reports whether a dropped stream should be opened
// again. Live radio and other endless HTTP streams qualify. A local file, a
// seekable track, or anything with a known length is allowed to end.
func streamShouldReconnect(tp *trackPipeline) bool {
	if tp == nil || tp.seekable || !isURL(tp.path) {
		return false
	}
	if tp.knownDuration > 0 || tp.decodedDuration > 0 {
		return false
	}
	return tp.live || tp.livePrefetch != nil
}

// Reconnecting reports whether a dropped live station is being opened again.
func (p *Player) Reconnecting() bool {
	return p.playing.Load() && !p.paused.Load() && p.reconnecting.Load()
}

// watchLive reopens a live station after the connection drops. It stays on
// the generation passed in, so stop and a newer play both end it.
func (p *Player) watchLive(path string, gen uint64) {
	defer func() {
		if p.playGen.Load() == gen {
			p.reconnecting.Store(false)
		}
	}()

	delay := liveWatchInterval
	backoff := liveReconnectMin
	for {
		timer := time.NewTimer(delay)
		<-timer.C
		if p.playGen.Load() != gen || !p.playing.Load() {
			return
		}
		if p.paused.Load() || !p.liveStreamDead() {
			p.reconnecting.Store(false)
			delay = liveWatchInterval
			backoff = liveReconnectMin
			continue
		}

		p.reconnecting.Store(true)
		err := p.replayLive(path, gen)
		if errors.Is(err, errPlaybackMoved) || p.playGen.Load() != gen || !p.playing.Load() {
			return
		}
		if err != nil {
			delay = backoff
			if backoff < liveReconnectMax {
				backoff *= 2
				if backoff > liveReconnectMax {
					backoff = liveReconnectMax
				}
			}
			continue
		}
		p.reconnecting.Store(false)
		delay = liveReconnectSettle
		backoff = liveReconnectMin
	}
}

func (p *Player) replayLive(path string, gen uint64) error {
	tp, err := p.buildPipeline(path)
	if err != nil {
		return err
	}
	return p.commitPlayback(tp, gen, false)
}

func (p *Player) liveStreamDead() bool {
	p.mu.Lock()
	cur := p.current
	p.mu.Unlock()
	if !streamShouldReconnect(cur) {
		return false
	}
	if p.gapless.Drained() {
		return true
	}
	return p.StreamErr() != nil
}

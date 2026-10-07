package audio

import "testing"

func TestStreamShouldReconnect(t *testing.T) {
	prefetch := &livePrefetchStreamer{}
	cases := []struct {
		name string
		tp   *trackPipeline
		want bool
	}{
		{name: "nil"},
		{name: "local file", tp: &trackPipeline{path: "/music/song.mp3"}},
		{name: "seekable http", tp: &trackPipeline{path: "https://example/song.mp3", seekable: true, live: true}},
		{name: "known length", tp: &trackPipeline{path: "https://example/song.mp3", live: true, knownDuration: 1}},
		{name: "decoded length", tp: &trackPipeline{path: "https://example/song.mp3", livePrefetch: prefetch, decodedDuration: 1}},
		{name: "youtube", tp: &trackPipeline{path: "https://example/watch", ytdlSeek: true}},
		{name: "icy radio", tp: &trackPipeline{path: "https://example/stream", live: true}, want: true},
		{name: "endless http", tp: &trackPipeline{path: "http://example/live", livePrefetch: prefetch}, want: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := streamShouldReconnect(tc.tp); got != tc.want {
				t.Fatalf("streamShouldReconnect() = %v, want %v", got, tc.want)
			}
		})
	}
}

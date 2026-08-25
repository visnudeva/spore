<table>
  <tr>
    <td>
      <strong>Spore, a terminal web radio and local music player with FFT visualizers.
<br>
    </td>
    <td>
  <img src="Logo.png" alt="spore" width="220">
</td>
  </tr>
</table>

[![spore](https://github.com/visnudeva/spore/blob/main/spore.png)](https://github.com/visnudeva/spore)



Built from the best of [tera](https://github.com/shinokada/tera) (radio browsing) and [cliamp](https://github.com/bjarneo/cliamp) (audio engine + visualizers).

## Features

- **Web radio** – browse top stations, search by name, explore tags ([radio-browser.info](https://www.radio-browser.info/))
- **Local files** – browse and play MP3, FLAC, OGG, WAV, M4A, Opus, and more
- **FFT visualizer** – 18 modes (starfield, plasma, lava, orbit, iris, …)
- **Favorites** – save stations and play them later
- **Session restore** – resumes the last radio station and visualizer on launch
- **Pure Go audio** – no mpv dependency; uses [beep](https://github.com/gopxl/beep) (+ ffmpeg for some codecs)

## Requirements

- Go 1.23+
- A working audio output (ALSA / PulseAudio / PipeWire)
- Optional: `ffmpeg` for AAC/M4A/Opus and some radio streams

## Install

```bash
git clone https://github.com/visnudeva/spore.git
cd spore
go build -o spore ./cmd/spore
./spore
```

## Keybindings

| Key | Action |
|-----|--------|
| `tab` | Files / Radio / Favorites |
| `enter` | Play |
| `space` / `p` | Pause / resume |
| `s` | Stop |
| `+` / `-` | Volume |
| `m` | Mute |
| `v` / `V` | Cycle / toggle visualizer |
| `/` | New radio search |
| `f` | Favorite |
| `d` | Remove favorite |
| `q` | Quit |

## Config

Favorites, history, and session state (`session.json`) live in `~/.config/spore/`.

On launch, spore restores the last visualizer mode and auto-plays the last radio station. If `~/.config/spore` is missing but `~/.config/wavr` exists, that directory is migrated once.

Override the config location with `SPORE_CONFIG_DIR`.

## Credits

- Audio engine and visualizers adapted from [cliamp](https://github.com/bjarneo/cliamp) by bjarneo
- Radio UX inspired by [tera](https://github.com/shinokada/tera) by shinokada
- Station data from [radio-browser.info](https://www.radio-browser.info/)

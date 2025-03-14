![wrtag](https://share.senan.xyz/wrtag.svg)

_Fast automated music tagging and organisation based on MusicBrainz_

---

**wrtag** is similar to music tagging and organisation tools such as [Beets](https://beets.io/) and [MusicBrainz Picard](https://picard.musicbrainz.org/) but aims to be simpler, more composable, and faster.

To describe the general workflow:

1. Input files are provided (existing or new).
2. They are matched with releases in the [MusicBrainz database](https://musicbrainz.org/). This may be done using existing tags or filenames.
3. The files are moved or copied to a user-defined filesystem layout. For example, `My music/Artist name/(Release year) Release name/1. Track title.ext`.
4. The file's tags are updated to match the tags in the MusicBrainz database.
5. Enjoy a clean, consistent, and accurate browsing experience in your music player/server of choice.

# Table of Contents

1. [Features](#features)
2. [Included tools](#included-tools)
   - [Tool `wrtag`](#tool-wrtag)
     - [Importing new music](#importing-new-music)
     - [Re-tagging already imported music](#re-tagging-already-imported-music)
   - [Tool `wrtagweb`](#tool-wrtagweb)
     - [API](#api)
     - [Configuration](#configuration)
     - [Options](#options)
   - [Tool `metadata`](#tool-metadata)
3. [Installation](#installation)
4. [Global configuration](#global-configuration)
   - [Options](#options-1)
   - [Format](#format)
     - [CLI arguments](#cli-arguments)
     - [Environment variables](#environment-variables)
     - [Config file](#config-file)
5. [Path format](#path-format)
   - [Basic structure](#basic-structure)
   - [Available template data](#available-template-data)
   - [Helper functions](#helper-functions)
   - [Example formats](#example-formats)
6. [Addons](#addons)
   - [Addon Lyrics](#addon-lyrics)
   - [Addon ReplayGain](#addon-replaygain)
   - [Addon Subprocess](#addon-subprocess)
7. [Notifications](#notifications)
8. [Goals and non-goals](#goals-and-non-goals)

# Features

- Unix-style suite of tools for different use cases, using the same core **wrtag** functionality.
- **Fast** tagging thanks to [Go](https://go.dev/).
- Filesystem organisation of music files, covers, and configurable extra files.
- **Cover fetching** or upgrades from the [Cover Art Archive](https://coverartarchive.org/).
- Care taken to ensure **no orphan** folders are left in the library when moves or copies occur.
- Validation to ensure your library is **always consistent** with no duplicates or unrecognised paths.
- Safe **concurrent** processing with tree-style filesystem locking.
- Addons for fetching lyrics, calculating [ReplayGain](https://wiki.hydrogenaud.io/index.php?title=ReplayGain_2.0_specification), or any user-defined subprocess.
- Rescanning the library and processing it for new changes in MusicBrainz (`wrtag sync`).
- An optional **web interface** for importing new releases over the network. Allows the user to be notified and confirm details if there is no 100% match found.
- Support for [gazelle-origin](https://github.com/x1ppy/gazelle-origin) files to improve matching from certain sources.
- Support for **Linux**, **macOS**, and **Windows** with static/portable [binaries available](https://github.com/sentriz/wrtag/releases) for each.

# Included tools

## Tool `wrtag`

<img align="right" width="300" src=".github/screenshot-wrtag-copy.png">

The `wrtag` tool is the main command line tool for music organisation. It has two subcommands for working on individual folders (`copy`, `move`), and one for bulk processing (`sync`).

### Importing new music

#### Moving from source

`wrtag` is the main command line tool that works on a single folder and requires an operation such as `move` or `copy`.

For example:

```console
$ wrtag move "Example"                 # tags and moves `Example` into the library defined by the path-format config option
$ wrtag move -dry-run "Example"        # shows move and tag operations without applying them
$ wrtag move -yes "Example"            # use anyway even if low match
$ wrtag move -mbid "abc" -yes "Example" # overwrite matched MusicBrainz release UUID
```

#### Copying from source

If the source files should be left alone, `wrtag` also provides a `copy` operation:

```console
$ wrtag copy -yes "Example"      # copies and tags `Example` into the library, use anyway even if low match
$ wrtag copy -mbid "abc" -yes    # overwrite matched MusicBrainz release UUID even if low match
```

### Re-tagging already imported music

Re-tagging your music can be useful for a few reasons. For example, if your [path-format](#path-format) configuration has changed, or the metadata in the MusicBrainz database has changed.

#### Re-tagging a single release

Since a `move` operation always cleans up the source directory - and is smart about the music already being in place - a re-tag is just a move pointed at some music that is already in place:

```console
# path-format is /my/music/XXX
$ wrtag move "/my/music/Tame Impala/(2010) Innerspeaker"
# now has updated tags, and moved again if needed
```

#### Re-tagging in bulk

Bulk operations are done with the `sync` subcommand. Unlike the `copy` and `move` commands which operate on single releases, the `sync` command works on your already imported and tagged library.

> [!WARNING]
> As the `sync` command is non-interactive, when used incorrectly it can be destructive. Only use `sync` on a library whose contents have been populated by `copy` or `move`.

By default, `sync` recurses through all directories rooted in `path-format` and finds leaf directories. A leaf directory is one that has no sub-folders, and therefore looks like a release. The tracks are read, and if they have a `MUSICBRAINZ_ALBUMID` (e.g., from `copy` or `move`), the release info is fetched from MusicBrainz and the release is re-tagged.

If no `MUSICBRAINZ_ALBUMID` is present, the release is matched as it usually would, and only re-tagged if a high match score is calculated.

```console
$ wrtag sync                          # recurse all releases and re-tag
$ wrtag sync -dry-run                 # show what above would do
$ wrtag sync "/my/music/Tame Impala"  # find all releases in "Tame Impala/" and re-tag those
$ wrtag sync -age-older 24h           # find all releases older than 1 day and re-tag
$ wrtag sync -num-workers 16          # process a maximum of 16 releases at a time
```

## Tool `wrtagweb`

<img align="right" width="300" src=".github/screenshot-wrtagweb.png">

`wrtagweb` is based on the same core functionality as `wrtag`, except it's web-based instead of command line. Instead of importing releases from the command line arguments, new releases are imported over HTTP.

For example, an HTTP client (a custom script, a BitTorrent client "on complete" script, or Soulseek) sends an HTTP request to `wrtagweb`, giving it a new path to import. `wrtagweb` imports it. If there isn't a 100% match, the user is notified. Then, the user can correct the match, reject, or accept anyway.

```mermaid
%%{init: {'theme': 'base', 'themeVariables': { 'fontSize': '26px' }}}%%
flowchart LR
    a["BitTorrent / Soulseek script"] -->
    b["Release added to queue"] -->
    c["User notified if input needed"] -->
    d["Release imported"]
```

### API

Jobs are added to the queue with an HTTP request like `POST <wrtag.host>/op/<copy|move>` with form value `path=<absolute path to directory>`. The form value can be an `application/x-www-form-urlencoded` form body, or URL query parameter.

Authentication is via HTTP basic auth password, configured by `web-api-key`.

<details>
<summary><b>Example with <i>cURL</i></b></summary>

```console
curl \
    --request POST \
    --data-urlencode "path=/path/to/the/release" \
    "https://:my-api-key@wrtag.hostname/op/copy"
```

</details>

<details>
<summary><b>Example with <i>Transmission</i></b></summary>

Create a script named `done.sh` or anything you like, and make it executable:
`chmod +x done.sh`

Update your Transmission `settings.json` to reference the new script:

```json
...
"script-torrent-done-enabled": true,
"script-torrent-done-filename": "/scripts/done.sh",
...
```

Edit the script to send a `copy` job with the newly finished torrent. Transmission will set `TR_TORRENT_NAME` to the name/path of the torrent. [See all variables](https://github.com/transmission/transmission/blob/main/docs/Scripts.md#on-torrent-completion)

```bash
#!/bin/sh

curl \
    --request POST \
    --data-urlencode "path=<path to downloads>/$TR_TORRENT_NAME" \
    "http://:<wrtag api key>@<wrtag host>/op/copy"
```

</details>

<details>
<summary><b>Example with <i>qBittorrent</i></b></summary>

> TODO

</details>

<details>
<summary><b>Example with <i>Deluge</i></b></summary>

> TODO

</details>

<details>
<summary><b>Example with <i>sldkd</i></b></summary>

> TODO

</details>

### Configuration

Configuration for `wrtagweb` works the same as [Global configuration](#global-configuration). For example, `wrtagweb -web-arg`, `WRTAG_WEB_ARG`, and the global config file is also read.

### Options

<!-- gen with `go run ./cmd/wrtagweb -h 2>&1 | ./gen-docs | wl-copy` -->

| CLI argument     | Environment variable  | Config file key | Description                                                  |
| ---------------- | --------------------- | --------------- | ------------------------------------------------------------ |
| -web-api-key     | WRTAG_WEB_API_KEY     | web-api-key     | API key for web interface                                    |
| -web-db-path     | WRTAG_WEB_DB_PATH     | web-db-path     | Path to database path for web interface (default "wrtag.db") |
| -web-listen-addr | WRTAG_WEB_LISTEN_ADDR | web-listen-addr | Listen address for web interface                             |
| -web-public-url  | WRTAG_WEB_PUBLIC_URL  | web-public-url  | Public URL for web interface (optional)                      |

## Tool `metadata`

> TODO

# Installation

- You can find static/portable binaries (wrtag, wrtagweb, metadata) on the [releases page](https://github.com/sentriz/wrtag/releases) for Windows, macOS, and Linux.
- Docker images for many architectures are available on [Docker Hub](https://hub.docker.com/r/sentriz/wrtag). The Docker image by default starts `wrtagweb`, but has the `wrtag` tools included too.
- To install from source, install a recent [Go](https://go.dev/) toolchain, clone the repo, and run `go install ./cmd/...` from inside.

# Global configuration

Global configuration is used by all tools. Any option can be provided with a CLI argument, environment variable, or config file key. See [Format](#format) for more technical details.

### Options

<!-- gen with `go run ./cmd/wrtag -h 2>&1 | ./gen-docs | wl-copy` -->

| CLI argument      | Environment variable   | Config file key  | Description                                                                                    |
| ----------------- | ---------------------- | ---------------- | ---------------------------------------------------------------------------------------------- |
| -addon            | WRTAG_ADDON            | addon            | Define an addon for extra metadata writing (see [Addons](#addons)) (stackable)                 |
| -caa-base-url     | WRTAG_CAA_BASE_URL     | caa-base-url     | CoverArtArchive base URL (default "<https://coverartarchive.org/>")                            |
| -caa-rate-limit   | WRTAG_CAA_RATE_LIMIT   | caa-rate-limit   | CoverArtArchive rate limit duration                                                            |
| -config           | WRTAG_CONFIG           | config           | Print the parsed config and exit                                                               |
| -config-path      | WRTAG_CONFIG_PATH      | config-path      | Path to config file (default "$XDG_CONFIG_HOME/wrtag/config")                                  |
| -cover-upgrade    | WRTAG_COVER_UPGRADE    | cover-upgrade    | Fetch new cover art even if it exists locally                                                  |
| -keep-file        | WRTAG_KEEP_FILE        | keep-file        | Define an extra file path to kept when moving/copying to root dir (stackable)                  |
| -log-level        | WRTAG_LOG_LEVEL        | log-level        | set the logging level (default INFO)                                                           |
| -mb-base-url      | WRTAG_MB_BASE_URL      | mb-base-url      | MusicBrainz base URL (default "<https://musicbrainz.org/ws/2/>")                               |
| -mb-rate-limit    | WRTAG_MB_RATE_LIMIT    | mb-rate-limit    | MusicBrainz rate limit duration (default 1s)                                                   |
| -notification-uri | WRTAG_NOTIFICATION_URI | notification-uri | Add a shoutrrr notification URI for an event (see [Notifications](#notifications)) (stackable) |
| -path-format      | WRTAG_PATH_FORMAT      | path-format      | Path to root music directory including path format rules (see [Path format](#path-format))     |
| -research-link    | WRTAG_RESEARCH_LINK    | research-link    | Define a helper URL to help find information about an unmatched release (stackable)            |
| -tag-weight       | WRTAG_TAG_WEIGHT       | tag-weight       | Adjust distance weighting for a tag (0 to ignore) (stackable)                                  |
| -version          | WRTAG_VERSION          | version          | Print the version and exit                                                                     |

### Format

### CLI arguments

Just call the command with the CLI argument. For example, `wrtag -some-key "some value"`. For stackable (repeatable) arguments, pass them multiple times. For example, `wrtag -some-key "value 1" -some-key "value 2"`.

> [!NOTE]
> Be aware of the position of global vs command arguments. For example, `wrtag <global options> cmd <cmd options>`. Check `-h` when in doubt.

### Environment variables

Environment variables are prefixed with `WRTAG_` usually. For example, `WRTAG_LOG_LEVEL=info wrtag`. For stackable (repeatable) arguments, join them with a comma (`,`). For example, `WRTAG_ADDON="replaygain,lyrics genius musixmatch"`. If the value of the variable should have a comma, it can be escaped with a backslash. For example, `\,`.

### Config file

The config file can be used instead of CLI arguments or environment variables. The format follows ([flagconf](https://pkg.go.dev/go.senan.xyz/flagconf?utm_source=godoc#hdr-usage)), which works like:

```
some-key argument
other-key argument
```

For stackable (repeatable) arguments, provide the same key multiple times. For example:

```
addon replaygain
addon lyrics genius musixmatch
```

See the [example config](./config.example) for more.

# Path format

The `path-format` configuration option defines both the root music directory and the template for organizing your music files. This template uses Go's text/template syntax and is populated with MusicBrainz release data.

## Basic structure

In order to minimise potential release conflict, the path format should include at least three path segments:

1. The root music directory (where all your music will be stored).
2. Artist/release organization (typically artist name and album details).
3. Track naming format (including track numbers and titles).

For example:

```
path-format /music/library/{{ <some artist format> }}/({{ <some release format> }}/{{ <track format> }}
```

This could format a release like:

```
/music/library/Tame Impala/(2010) Innerspeaker/01. It Is Not Meant to Be.flac
/music/library/Tame Impala/(2010) Innerspeaker/02. Desire Be Desire Go.flac
/music/library/Tame Impala/(2010) Innerspeaker/03. Alter Ego.flac
...
```

On Windows, you can use drive letters and backslashes:

```
path-format C:\User\John\Music\{{ <some artist format> }}\({{ <some release format> }}\{{ <track format> }}
```

## Available template data

The template has access to the following data:

- `.Release` - The full MusicBrainz release object (see [`type Release struct {`](https://github.com/sentriz/wrtag/blob/master/musicbrainz/musicbrainz.go))
- `.Track` - The current track being processed (see [`type Track struct {`](https://github.com/sentriz/wrtag/blob/master/musicbrainz/musicbrainz.go))
- `.TrackNum` - The track number (integer)
- `.IsCompilation` - Boolean indicating if this is a compilation album
- `.Ext` - The file extension including the dot (e.g., ".flac")

## Helper functions

Several helper functions are available to format your paths:

| Function              | Description                            | Example                                       |
| --------------------- | -------------------------------------- | --------------------------------------------- |
| `join`                | Joins strings with a delimiter         | `{{ artists .Release.Artists \| join "; " }}` |
| `pad0`                | Zero-pads a number to specified width  | `{{ pad0 2 .TrackNum }}` → "01"               |
| `sort`                | Sorts a string array                   | `{{ artists .Release.Artists \| sort }}`      |
| `safepath`            | Makes a string safe for filesystem use | `{{ .Track.Title \| safepath }}`              |
| `flatTracks`          | Flattens tracks from all media         | `{{ flatTracks .Release.Media \| len }}`      |
| `artists`             | Gets artist names from artist credits  | `{{ artists .Release.Artists }}`              |
| `artistsString`       | Formats artists as a string            | `{{ artistsString .Track.Artists }}`          |
| `artistsCredit`       | Gets credit names from artist credits  | `{{ artistsCredit .Release.Artists }}`        |
| `artistsCreditString` | Formats artist credits as a string     | `{{ artistsCreditString .Release.Artists }}`  |
| `disambig`            | Gets disambiguation text for a release | `{{ disambig .Release }}`                     |

## Example formats

### The recommended format

Including multi album artist support, release group year, release group and release disambiguations, track numbers, total track numbers, artist names the release is a compilation album

```
/music/{{ artists .Release.Artists | sort | join "; " | safepath }}/({{ .Release.ReleaseGroup.FirstReleaseDate.Year }}) {{ .Release.Title | safepath }}{{ if not (eq (disambig .Release) "") }} ({{ disambig .Release | safepath }}){{ end }}/{{ pad0 2 .TrackNum }}.{{ flatTracks .Release.Media | len | pad0 2 }} {{ if .IsCompilation }}{{ artistsString .Track.Artists | safepath }} - {{ end }}{{ .Track.Title | safepath }}{{ .Ext }}
```

### A basic format

```
/music/{{ artists .Release.Artists | join "; " | safepath }}/{{ .Release.Title | safepath }}/{{ pad0 2 .TrackNum }} {{ .Track.Title | safepath }}{{ .Ext }}
```

### With year and disambiguation

```
/music/{{ artists .Release.Artists | sort | join "; " | safepath }}/({{ .Release.ReleaseGroup.FirstReleaseDate.Year }}) {{ .Release.Title | safepath }}{{ if not (eq (disambig .Release) "") }} ({{ disambig .Release | safepath }}){{ end }}/{{ pad0 2 .TrackNum }} {{ .Track.Title | safepath }}{{ .Ext }}
```

### With compilation handling

```
/music/{{ artists .Release.Artists | sort | join "; " | safepath }}/({{ .Release.ReleaseGroup.FirstReleaseDate.Year }}) {{ .Release.Title | safepath }}/{{ pad0 2 .TrackNum }} {{ if .IsCompilation }}{{ artistsString .Track.Artists | safepath }} - {{ end }}{{ .Track.Title | safepath }}{{ .Ext }}
```

### With disc and track numbers

```
/music/{{ artists .Release.Artists | sort | join "; " | safepath }}/({{ .Release.ReleaseGroup.FirstReleaseDate.Year }}) {{ .Release.Title | safepath }}/{{ pad0 2 .TrackNum }}.{{ flatTracks .Release.Media | len | pad0 2 }} {{ .Track.Title | safepath }}{{ .Ext }}
```

## Validation

The path format is validated to ensure:

1. It contains at least three path segments
2. Different tracks from the same release get different paths
3. Tracks with the same name from different releases get different paths
4. No paths end with a trailing slash or contain double slashes

This validation helps prevent file conflicts and ensures a consistent library structure.

The path format is validated to ensure:

# Addons

Addons can be used to fetch/compute additional metadata after the MusicBrainz match has been applied and the files have been tagged.

They are configured as part of the [global configuration](#global-configuration) using a [config format](#format).

For example:

- `$ wrtag -addon "lyrics a b c" -addon "replaygain x y z"`
- `$ WRTAG_ADDON="lyrics a b c,replaygain" wrtag`
- or repeating the `addon` clause in the config file.

## Addon Lyrics

The `lyrics` addon can fetch and embed lyric information from [Genius](https://genius.com/) and [Musixmatch](https://www.musixmatch.com/) in your tracks.

The format of the addon config is `lyrics <source>...` where the source is one of `genius` or `musixmatch`. For example, `"lyrics genius musixmatch"`. Note that sources will be tried in the order they are specified.

## Addon ReplayGain

The `replaygain` addon computes and adds [ReplayGain 2.0](https://wiki.hydrogenaud.io/index.php?title=ReplayGain_2.0_specification) information to your files. It is great for normalising the perceived loudness of audio in your tracks.

The format of the addon config is `replaygain <opts>...` where opts can be `true-peak` and `force`. If the force option is passed, ReplayGain information is recomputed even if it’s already present in the files.

## Addon Subprocess

The subprocess addon is for running a user-provided program.

The format of the addon config is `subproc <path> <args>...`, where `path` is the path to the program, or the program name itself if it’s in your `$PATH`. `args` are extra command line arguments to pass to the program. One of the `args` should be a special placeholder named `<files>`. This will be expanded to the paths to the files that were just processed by `wrtag`.

For example, the addon `"subproc my-program a --b 'c d' <files>"` might call `my-program` with arguments `["a", "--b", "c d", "track 1.flac", "track 2.flac", "track 3.flac"]` after importing a release with 3 tracks.

# Notifications

Notifications can be used to notify you or another system of events such as importing or syncing. For example, sending an email when user input is needed to import a release. Or notifying your [music server](https://github.com/sentriz/gonic) after a sync has completed.

The possible events are:

| Tool       | Name            | Description                                    |
| ---------- | --------------- | ---------------------------------------------- |
| `wrtagweb` | `complete`      | Executed when a release is imported            |
| `wrtagweb` | `needs-input`   | Executed when a release requires input         |
| `wrtag`    | `sync-complete` | Executed when a sync has completed             |
| `wrtag`    | `sync-error`    | Executed when a sync has completed with errors |

`wrtag` uses [shoutrrr](https://github.com/containrrr/shoutrrr) to provide upstream notifications over SMTP, HTTP, etc.

For example:

- `smtp://username:password@host:port/?from=from@example.com&to=recipient@example.com`
- `generic+https://my.subsonic.com/rest/startScan.view?c=wrtag&v=1.16&u=user&p=password`

See [the shoutrrr documentation](https://containrrr.dev/shoutrrr/v0.8/services/overview/) for the list of providers.

Notifications are configured as part of the [global configuration](#global-configuration). The format is `<eventspec> <shoutrrr uri>`. `eventspec` is a comma-separated list of event names.

For example, `"complete,sync-complete smtp://example.com"`. Multiple eventspec and URIs can be configured by stacking the config option according to the [config format](#format).

# Goals and non-goals

> TODO

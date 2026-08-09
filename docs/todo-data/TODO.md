# TODO

---
## New Features

- Consider aetherengine
- Allow uncached
- Transcoding
- Migrate settings to frontend
- https://check.snzb.stream/docs/
- AIOmetadata support
- Watch together

---

## Bugs

- Web playback not using nvenc
- MPV subtitles
- No online subs on Android mobile
- DV supports in mpv player
- Scrob support
- stream codes not auto populating EPG
- MacOS sizing issues (likely iPad as well)
- Voice search not working on android

---

## Testing

- In Progress Testing — Manual “Prequeue forever”: add movies and series from Details → More Options, Home/index long-press, and Watchlist long-press; confirm an already-active details prequeue is adopted without duplicate resolution, the item survives its normal TTL and a backend restart, playback reuses it, and the admin Active Prequeues page labels it `MANUAL · FOREVER` and can remove it individually or via Clear All
- In Progress Testing — Admin maintenance template navigation: open Server → Maintenance with and without an enabled prewarm automation, verify Prequeue Management is always visible and opens Active Prequeues, then verify Hidden Items, Bad Streams, Resolved NZBs, and Share Links open and each subpage can return to Maintenance
- In Progress Testing — Released episode searches retry empty text-based providers with a season-scoped query: verify Captain Star S01E01 discovers the `S01-S02 Complete` Jackett result, providers that return an exact episode do not make a fallback request, and future/unreleased episodes do not run a season fallback
- In Progress Testing — Scrob integration: link a writable Scrob account in Profile Scrobbling and confirm newly completed movies and episodes are pushed with their watched timestamps. Confirm active playback appears in Scrob Now Playing, pause/resume and progress update, and stop/completion clears the session. Then run Scrob → Local, Local → Scrob, and bidirectional history automations; confirm TMDB/show coordinates survive, repeat runs do not create duplicate remote plays, and a recent local unwatch removes the corresponding Scrob item. For 2FA accounts, verify live profile pushes and scheduled sync generate valid rotating codes.
- In Progress Testing — Discord partial-progress notifications are deleted after unfinished playback stops, including overlapping sessions, temporary Discord failures, and backend restarts; verify no 0% or partial episode message remains after the two-minute stale timeout
- In Progress Testing — Android Live TV no longer trusts the redacted client settings shape to decide whether channels exist: verify the affected non-master profile loads channels, and collect a fresh log package if it still fails so the new `[live-channels]` request/settings-shape diagnostics can be compared
- In Progress Testing — Dashboard uses explicit player source metadata: verify AIOStreams/Comet HTTP playback shows Debrid, Usenet WebDAV playback shows Usenet, and legacy clients still fall back to path classification
- In Progress Testing — Dashboard shelf live TV cards use the active channel logo and selecting one starts that channel through the same playback flow as Favorite Channels
- In Progress Testing — Fire TV/Android TV press-and-hold voice search diagnostics added; capture `VoiceSearchDiag`, `VoiceSearch`, and `ReactNativeJS` with ADB while reproducing on the Search page
- In Progress Testing — Apple TV screensaver/background exit saves the latest playback position before returning to details; also verify resuming from the saved position on multiple Apple TV devices
- MPV migration crash
- DV/HDR support in MPV
- Little bug, on my firestick, if I click the select button during the stream instead of pause it, it pause a fraction of seconds and the bottom menu appear

---

## Un-Testable

- Premiumize support added
- In Progress Testing — Trakt VIP Smart Lists and personal-list pagination: connect a Trakt VIP account that owns a dynamic/Smart List, open the Trakt custom-list picker, and verify both personal lists and Smart Lists appear, Smart Lists are labeled distinctly, and selecting one successfully imports/syncs all resolved movie/show items across multiple pages. Also verify a non-VIP account still lists and syncs ordinary personal lists without errors.
- Android crashing on playback https://paste.c-net.org/f65cd509-d2a2-9d49-c47d-63cfc8c3ace9 Potentially related to Atmos
- In Progress Testing — Library scan survives adding libraries: start a scan on a large local library, then add another library (and set access) while progress is still moving; confirm the first scan keeps progressing, Scan stays disabled with Working…, completion toast still fires, and adding the second library does not clear or fail the first scan. Also confirm a second Scan click while scanning errors with conflict, and scheduled local-media scans still complete.
- Infuse playback death

---

## One Day

- Chromecasting still has issues with subtitles
- YouTube playback not working with Gluetun proxy (need logs)
- Alternate App Icons
- Apple TV profile linking

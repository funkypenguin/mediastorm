# TODO

---
## New Features

- Consider aetherengine
- Allow uncached
- Transcoding
- Migrate settings to frontend
- https://check.snzb.stream/docs/
- Manual prequeue
- AIOmetadata support
- Watch together

---

## Bugs

- Episode search needs a broad title/season-pack fallback when exact SxxExx candidates fail, so valid complete-series releases are considered
- Web playback not using nvenc
- MPV subtitles
- No online subs on Android mobile
- DV supports in mpv player
- Scrob support
- stream codes not auto populating EPG
- MacOS sizing issues (likely iPad as well)
- Notifications lingering in discord
- Voice search not working on android

---

## Testing

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

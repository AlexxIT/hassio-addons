This add-on now access `go2rtc.yaml` from `/addon_configs/a889bffc_go2rtc/go2rtc.yaml` instead of from `/config/go2rtc.yaml`.

- You need to move your `go2rtc.yaml` file to the new location: `mv -v /config/go2rtc.yaml /addon_configs/a889bffc_go2rtc/go2rtc.yaml`.
- References to the Home Assistant `/config/` directory in the `go2rtc.yaml` file should be updated to `/homeassistant/`.

The good news is that when this add-on is backed up, the `go2rtc.yaml` file will now be included in the backup, so you can easily restore it if needed.

For go2rtc release notes, see https://github.com/AlexxIT/go2rtc/releases.

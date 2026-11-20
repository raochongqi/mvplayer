# Linux packaging

The binary links the system libmpv (`libmpv.so.2`), so the package must depend
on `libmpv2` (Debian/Ubuntu: `apt install libmpv2`).

## AppImage

```sh
make build-mpv
mkdir -p AppDir/usr/bin AppDir/usr/share/applications AppDir/usr/share/icons/hicolor/256x256/apps
cp build/mvplayer AppDir/usr/bin/
cp packaging/linux/mvplayer.desktop AppDir/usr/share/applications/
cp packaging/linux/mvplayer.png AppDir/usr/share/icons/hicolor/256x256/apps/
linuxdeploy --appdir AppDir --output appimage
```

`linuxdeploy` is not vendored; download it from
https://github.com/linuxdeploy/linuxdeploy/releases. Bundling libmpv into the
AppImage adds ~25–30 MiB compressed and removes the system dependency — only do
that if the 50 MiB installer budget allows it.

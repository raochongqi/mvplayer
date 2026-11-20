; mvPlayer Windows installer (NSIS 3, LZMA solid).
; Build: makensis packaging/windows/installer.nsi   (after `make package-windows`)
Unicode true
!include "MUI2.nsh"

!define APPNAME "mvPlayer"
!define VERSION "0.1.0"
Name "${APPNAME} ${VERSION}"
OutFile "..\..\build\mvPlayer-${VERSION}-setup.exe"
InstallDir "$LOCALAPPDATA\mvPlayer"
RequestExecutionLevel user
SetCompressor /SOLID lzma

!insertmacro MUI_PAGE_DIRECTORY
!insertmacro MUI_PAGE_INSTFILES
!insertmacro MUI_LANGUAGE "English"

Section "mvPlayer"
  SetOutPath "$INSTDIR"
  File "..\..\build\mvplayer.exe"
  File "..\..\build\libmpv-2.dll"
  CreateShortcut "$SMPROGRAMS\mvPlayer.lnk" "$INSTDIR\mvplayer.exe"
  WriteUninstaller "$INSTDIR\uninstall.exe"
SectionEnd

Section "Uninstall"
  Delete "$INSTDIR\mvplayer.exe"
  Delete "$INSTDIR\libmpv-2.dll"
  Delete "$INSTDIR\uninstall.exe"
  Delete "$SMPROGRAMS\mvPlayer.lnk"
  RMDir "$INSTDIR"
SectionEnd

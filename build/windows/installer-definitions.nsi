!include "MUI2.nsh"

!define NAME "Wahay"

Name "${NAME}"
OutFile "${NAME} installer.exe"
Unicode True

!define MUI_ICON "wahay.ico"
!define MUI_UNICON "wahay.ico"

Caption "Wahay Installer"
BrandingText " "

InstallDir "$ProgramFiles\${Name}"

!define MUI_WELCOMEPAGE_TITLE "Welcome to Wahay Installer"
!define MUI_WELCOMEPAGE_TEXT "This installer will guide you through the installation of Wahay.$\r$\n$\r$\n$\r$\n$\r$\n$_CLICK"
!define MUI_LICENSEPAGE_TEXT_BOTTOM "If you accept the terms of the agreement, click I Agree to continue."
!define MUI_FINISHPAGE_NOREBOOTSUPPORT

!insertmacro MUI_PAGE_WELCOME
!insertmacro MUI_PAGE_LICENSE "LICENSE"
!define MUI_COMPONENTSPAGE_NODESC
!insertmacro MUI_PAGE_COMPONENTS
!insertmacro MUI_PAGE_DIRECTORY
!insertmacro MUI_PAGE_INSTFILES
!insertmacro MUI_PAGE_FINISH

!insertmacro MUI_UNPAGE_CONFIRM
!insertmacro MUI_UNPAGE_INSTFILES

!insertmacro MUI_LANGUAGE "English"

Section "Wahay"
  SetOutPath "$INSTDIR"

  SectionIn 1 RO

  File /oname=wahay.exe "wahay.exe"
  File "dll\*.dll"
  File /r "share"
  File /r "lib"
  File /oname=wahay.ico "wahay-256x256.ico"

  WriteUninstaller "$INSTDIR\Uninstall.exe"
  WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\${NAME}"   "DisplayName" "${NAME}"
  WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\${NAME}"   "UninstallString" "$INSTDIR\Uninstall.exe"
SectionEnd

Section "Start Menu shortcut"
  CreateShortCut "$SMPROGRAMS\${NAME}.lnk" "$INSTDIR\wahay.exe" "" "$INSTDIR\wahay.ico"
SectionEnd

Section "Desktop shortcut"
  CreateShortCut "$DESKTOP\${NAME}.lnk" "$INSTDIR\wahay.exe" "" "$INSTDIR\wahay.ico"
SectionEnd

Section "Uninstall"
  Delete "$INSTDIR\wahay.exe"
  Delete "$INSTDIR\*.dll"
  Delete "$INSTDIR\wahay.ico"
  RMDir /r "$INSTDIR\lib"
  RMDir /r "$INSTDIR\share"

  Delete "$SMPROGRAMS\${NAME}.lnk"
  Delete "$DESKTOP\${NAME}.lnk"
  DeleteRegKey HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\${NAME}"
  Delete "$INSTDIR\Uninstall.exe"
  RMDir "$INSTDIR"
SectionEnd

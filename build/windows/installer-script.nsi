!include "MUI2.nsh"

!define NAME "Wahay"

Name "${NAME}"
OutFile "${NAME} Installer.exe"
Unicode True

!define MUI_ICON "wahay.ico"
!define MUI_UNICON "wahay.ico"

Caption "Wahay (${VERSION}) Installer"
BrandingText " "

InstallDir "$ProgramFiles\${Name}"

!define MUI_WELCOMEPAGE_TITLE "Welcome to Wahay Installer"
!define MUI_WELCOMEPAGE_TEXT "This installer will guide you through the installation of Wahay.$\r$\n$\r$\n$\r$\n$\r$\n$_CLICK"
!define MUI_LICENSEPAGE_TEXT_BOTTOM "If you accept the terms of the agreement, click I Agree to continue."
!define MUI_FINISHPAGE_NOREBOOTSUPPORT

!insertmacro MUI_PAGE_WELCOME
!insertmacro MUI_PAGE_LICENSE "LICENSE"
!define MUI_COMPONENTSPAGE_NODESC
!define MUI_COMPONENTSPAGE_TEXT_TOP "Select the components you want to include in your installation."
!define MUI_COMPONENTSPAGE_TEXT_COMPLIST "Choose the features you want to install. Uncheck any components you don't need."
!insertmacro MUI_PAGE_COMPONENTS
!insertmacro MUI_PAGE_DIRECTORY
!insertmacro MUI_PAGE_INSTFILES

!define MUI_FINISHPAGE_LINK "https://wahay.org/"
!define MUI_FINISHPAGE_LINK_LOCATION "https://wahay.org/"
!insertmacro MUI_PAGE_FINISH

!insertmacro MUI_UNPAGE_CONFIRM
!insertmacro MUI_UNPAGE_INSTFILES

!insertmacro MUI_LANGUAGE "English"

Section "Wahay App"
  SetOutPath "$INSTDIR"

  SectionIn 1 RO

  File /oname=wahay.exe "wahay.exe"
  File "dll\*.dll"
  File /r "share"
  File /r "lib"
  File "gdbus.exe"

  File /oname=wahay.ico "wahay-256x256.ico"

  WriteUninstaller "$INSTDIR\Uninstall.exe"
  WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\${NAME}"   "DisplayName" "${NAME}"
  WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\${NAME}"   "UninstallString" "$INSTDIR\Uninstall.exe"
SectionEnd

SectionGroup /e "Shortcuts"
  Section "Start Menu"
    CreateShortCut "$SMPROGRAMS\${NAME}.lnk" "$INSTDIR\wahay.exe" "" "$INSTDIR\wahay.ico"
  SectionEnd

  Section "Desktop"
    CreateShortCut "$DESKTOP\${NAME}.lnk" "$INSTDIR\wahay.exe" "" "$INSTDIR\wahay.ico"
  SectionEnd
SectionGroupEnd


Section "Uninstall"
  Delete "$INSTDIR\wahay.exe"
  Delete "$INSTDIR\*.dll"
  Delete "$INSTDIR\wahay.ico"
  RMDir /r "$INSTDIR\lib"
  RMDir /r "$INSTDIR\share"
  Delete "$INSTDIR\gdbus.exe"

  Delete "$SMPROGRAMS\${NAME}.lnk"
  Delete "$DESKTOP\${NAME}.lnk"
  DeleteRegKey HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\${NAME}"
  Delete "$INSTDIR\Uninstall.exe"
  RMDir "$INSTDIR"
SectionEnd

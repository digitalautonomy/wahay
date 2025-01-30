!include "MUI2.nsh"

!define NAME "Wahay"

Name "${NAME}"
OutFile "${NAME} Installer.exe"
Unicode True

!define MUI_ICON "wahay-256x256.ico"
!define MUI_UNICON "wahay-256x256.ico"

Caption "$(Caption)"
BrandingText " "

InstallDir "$PROGRAMFILES64\${Name}"

!define MUI_WELCOMEPAGE_TITLE "$(WelcomeTitle)"
!define MUI_WELCOMEPAGE_TEXT "$(WelcomeText)"
!define MUI_LICENSEPAGE_TEXT_BOTTOM "$(LicenseText)"
!define MUI_FINISHPAGE_NOREBOOTSUPPORT

!insertmacro MUI_PAGE_WELCOME
!insertmacro MUI_PAGE_LICENSE "LICENSE"
!define MUI_COMPONENTSPAGE_NODESC
!define MUI_COMPONENTSPAGE_TEXT_TOP "$(ComponentsText)"
!define MUI_COMPONENTSPAGE_TEXT_COMPLIST "$(ComponentsList)"
!insertmacro MUI_PAGE_COMPONENTS
!insertmacro MUI_PAGE_DIRECTORY
!insertmacro MUI_PAGE_INSTFILES

!define MUI_FINISHPAGE_LINK "https://wahay.org/"
!define MUI_FINISHPAGE_LINK_LOCATION "https://wahay.org/"
!insertmacro MUI_PAGE_FINISH

!insertmacro MUI_UNPAGE_CONFIRM
!insertmacro MUI_UNPAGE_INSTFILES

!insertmacro MUI_LANGUAGE "English"
!insertmacro MUI_LANGUAGE "Spanish"

LangString Caption ${LANG_ENGLISH} "Wahay (${VERSION}) Installer (64 bits)"
LangString Caption ${LANG_SPANISH} "Instalador de Wahay (${VERSION})"

LangString WelcomeTitle ${LANG_ENGLISH} "Welcome to Wahay Installer"
LangString WelcomeTitle ${LANG_SPANISH} "Bienvenido al Instalador de Wahay"

LangString WelcomeText ${LANG_ENGLISH} "This installer will guide you through the installation of Wahay.$\r$\n$\r$\n$\r$\n$\r$\n$_CLICK"
LangString WelcomeText ${LANG_SPANISH} "Este instalador le guiará a través de la instalación de Wahay.$\r$\n$\r$\n$\r$\n$\r$\n$_CLICK"

LangString LicenseText ${LANG_ENGLISH} "If you accept the terms of the agreement, click I Agree to continue."
LangString LicenseText ${LANG_SPANISH} "Si acepta los términos del acuerdo, haga clic en Acepto para continuar."

LangString ComponentsText ${LANG_ENGLISH} "Select the components you want to include in your installation. You can install Wahay along with additional tools like Tor and Mumble if you don't have them yet."
LangString ComponentsText ${LANG_SPANISH} "Seleccione los componentes que desea incluir en su instalación. Puede instalar Wahay junto con herramientas adicionales como Tor y Mumble si aún no las tiene."

LangString ComponentsList ${LANG_ENGLISH} "Choose the features you want to install. Uncheck any components you don't need."
LangString ComponentsList ${LANG_SPANISH} "Elija las características que desea instalar. Desmarque los componentes que no necesite."

LangString SecWahayName ${LANG_ENGLISH} "Wahay App"
LangString SecWahayName ${LANG_SPANISH} "Aplicación Wahay"

LangString SecShortcutsName ${LANG_ENGLISH} "Shortcuts"
LangString SecShortcutsName ${LANG_SPANISH} "Accesos Directos"

LangString SecStartMenuName ${LANG_ENGLISH} "Start Menu"
LangString SecStartMenuName ${LANG_SPANISH} "Menú de Inicio"

LangString SecDesktopName ${LANG_ENGLISH} "Desktop"
LangString SecDesktopName ${LANG_SPANISH} "Escritorio"

Function .onInit
  !insertmacro MUI_LANGDLL_DISPLAY
  WriteRegDWORD HKLM "Software\${NAME}" "Installer Language" $LANGUAGE
FunctionEnd

Section "$(SecWahayName)"
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

SectionGroup /e "$(SecShortcutsName)"
  Section "$(SecStartMenuName)"
    CreateShortCut "$SMPROGRAMS\${NAME}.lnk" "$INSTDIR\wahay.exe" "" "$INSTDIR\wahay.ico"
  SectionEnd

  Section "$(SecDesktopName)"
    CreateShortCut "$DESKTOP\${NAME}.lnk" "$INSTDIR\wahay.exe" "" "$INSTDIR\wahay.ico"
  SectionEnd
SectionGroupEnd

Function un.onInit
    ReadRegDWORD $LANGUAGE HKLM "Software\${NAME}" "Installer Language"
FunctionEnd

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
  DeleteRegKey HKLM "SOFTWARE\WOW6432Node\Wahay"
  Delete "$INSTDIR\Uninstall.exe"
  RMDir "$INSTDIR"
SectionEnd

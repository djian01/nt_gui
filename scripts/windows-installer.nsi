Unicode True
!include "MUI2.nsh"
!include "LogicLib.nsh"
!include "x64.nsh"
Name "NET-Test"
OutFile "${OUTPUT_FILE}"
InstallDir "$LOCALAPPDATA\Programs\net-test"
RequestExecutionLevel user
SetCompressor /SOLID lzma
!define UNINSTALL_KEY "Software\Microsoft\Windows\CurrentVersion\Uninstall\NETTest"
!define MUI_ICON "${REPO_DIR}\scripts\installer-assets\installer.ico"
!insertmacro MUI_PAGE_WELCOME
!insertmacro MUI_PAGE_LICENSE "${REPO_DIR}\LICENSE"
!insertmacro MUI_PAGE_INSTFILES
!insertmacro MUI_PAGE_FINISH
!insertmacro MUI_UNPAGE_CONFIRM
!insertmacro MUI_UNPAGE_INSTFILES
!insertmacro MUI_LANGUAGE "English"

Function .onInit
  ${IfNot} ${RunningX64}
    MessageBox MB_OK|MB_ICONSTOP "NET-Test requires 64-bit Windows."
    Abort
  ${EndIf}
  !if "${ARCH}" == "arm64"
    ReadRegStr $0 HKLM "SYSTEM\CurrentControlSet\Control\Session Manager\Environment" "PROCESSOR_ARCHITECTURE"
    ${If} $0 != "ARM64"
      MessageBox MB_OK|MB_ICONSTOP "This package requires Windows on ARM64. Download the amd64 package for Intel/AMD PCs."
      Abort
    ${EndIf}
  !endif
  ; Microsoft's documented Evergreen runtime registration, per user or machine.
  SetRegView 32
  ReadRegStr $0 HKCU "Software\Microsoft\EdgeUpdate\Clients\{F3017226-FE2A-4295-8BDF-00C3A9A7E4C5}" "pv"
  ${If} $0 == ""
  ${OrIf} $0 == "0.0.0.0"
    ReadRegStr $0 HKLM "Software\Microsoft\EdgeUpdate\Clients\{F3017226-FE2A-4295-8BDF-00C3A9A7E4C5}" "pv"
  ${EndIf}
  ${If} $0 == ""
  ${OrIf} $0 == "0.0.0.0"
    MessageBox MB_OK|MB_ICONSTOP "Install Microsoft Edge WebView2 Evergreen Runtime, then run this installer again. Download: https://developer.microsoft.com/microsoft-edge/webview2/"
    Abort
  ${EndIf}
  MessageBox MB_OKCANCEL "Close NET-Test before installing or upgrading." IDOK ready
  Abort
  ready:
FunctionEnd

Section "Install"
  SetShellVarContext current
  SetOutPath "$INSTDIR"
  File "${SOURCE_DIR}\net-test.exe"
  File "${REPO_DIR}\LICENSE"
  WriteUninstaller "$INSTDIR\Uninstall.exe"
  CreateShortcut "$SMPROGRAMS\NET-Test.lnk" "$INSTDIR\net-test.exe"
  WriteRegStr HKCU "${UNINSTALL_KEY}" "DisplayName" "NET-Test"
  WriteRegStr HKCU "${UNINSTALL_KEY}" "DisplayVersion" "${VERSION}"
  WriteRegStr HKCU "${UNINSTALL_KEY}" "Publisher" "Packet Streams"
  WriteRegStr HKCU "${UNINSTALL_KEY}" "DisplayIcon" "$INSTDIR\net-test.exe"
  WriteRegStr HKCU "${UNINSTALL_KEY}" "UninstallString" '"$INSTDIR\Uninstall.exe"'
  WriteRegDWORD HKCU "${UNINSTALL_KEY}" "NoModify" 1
  WriteRegDWORD HKCU "${UNINSTALL_KEY}" "NoRepair" 1
SectionEnd

Section "Uninstall"
  MessageBox MB_OKCANCEL "Close NET-Test before uninstalling." IDOK uninstall_ready
  Abort
  uninstall_ready:
  SetShellVarContext current
  ClearErrors
  Delete "$INSTDIR\net-test.exe"
  ${If} ${Errors}
    MessageBox MB_OK|MB_ICONSTOP "Could not remove NET-Test. Close the app and run uninstall again."
    Abort
  ${EndIf}
  Delete "$SMPROGRAMS\NET-Test.lnk"
  Delete "$INSTDIR\LICENSE"
  Delete "$INSTDIR\Uninstall.exe"
  RMDir "$INSTDIR"
  DeleteRegKey HKCU "${UNINSTALL_KEY}"
SectionEnd

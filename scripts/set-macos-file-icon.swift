import AppKit

// The DMG file icon is separate from the mounted volume's .VolumeIcon.icns.
guard CommandLine.arguments.count == 3,
      let icon = NSImage(contentsOfFile: CommandLine.arguments[1]),
      NSWorkspace.shared.setIcon(icon, forFile: CommandLine.arguments[2], options: []) else {
    fputs("Unable to set the installer file icon.\n", stderr)
    exit(1)
}

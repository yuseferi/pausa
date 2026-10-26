cask "pausa" do
  version "1.9.4"

  on_arm do
    sha256 "c50732b050b9abe3897ef9560a51b2c841f89a36dbdfc282767b2b11319139dd"
    url "https://github.com/yuseferi/pausa/releases/download/v#{version}/pausa-#{version}-arm64-macos.zip"
  end

  on_intel do
    sha256 "5a4819b115146b025b36b6e3a00a30910cb8025ff84ba1a51b16cd5bccaa0e0e"
    url "https://github.com/yuseferi/pausa/releases/download/v#{version}/pausa-#{version}-amd64-macos.zip"
  end

  name "Pausa"
  desc "Native-feeling macOS break reminder with fullscreen-space overlays"
  homepage "https://yuseferi.github.io/pausa/"

  app "pausa.app"

  # Pausa is distributed unsigned (no paid Apple Developer
  # account), so a downloaded bundle carries the
  # com.apple.quarantine attribute and macOS refuses the first
  # launch. Strip it so plain `brew install` works.
  postflight do
    system_command "/usr/bin/xattr",
                   args: ["-dr", "com.apple.quarantine", "#{appdir}/pausa.app"],
                   must_succeed: true
  end

  zap trash: [
    "~/Library/Application Support/Pausa",
    "~/Library/Logs/Pausa",
  ]
end

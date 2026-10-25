cask "pausa" do
  version "1.9.0"

  on_arm do
    sha256 "7559c2dcfc27881c72078782f3f8289a7bb0a97f687eb13639bac0333522b532"
    url "https://github.com/yuseferi/pausa/releases/download/v#{version}/pausa-#{version}-arm64-macos.zip"
  end

  on_intel do
    sha256 "39f0e83b7c43b72bafb5c3bc58fa59e576e1a872c46768b8764f09081b405190"
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

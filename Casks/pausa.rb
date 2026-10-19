cask "pausa" do
  version "1.8.0"

  on_arm do
    sha256 "dbddefa01c0aa438fea83d766de71df24819b4882b41e4ccfdb9e347a794614b"
    url "https://github.com/yuseferi/pausa/releases/download/v#{version}/pausa-#{version}-arm64-macos.zip"
  end

  on_intel do
    sha256 "6aa387146fba8e0df1bfbc566ca7582b44e513d5be289c31d7cc233af188711a"
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

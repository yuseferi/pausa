cask "pausa" do
  version "1.9.0"

  on_arm do
    sha256 "79aa384bcf176a9e3bf94f5604ed8a5dea2a6f4685038b6a1284cd8f5ec17197"
    url "https://github.com/yuseferi/pausa/releases/download/v#{version}/pausa-#{version}-arm64-macos.zip"
  end

  on_intel do
    sha256 "eaf54133dd8acd58fc3e0a645cabba5aeb3e917e42caad56f533e7360bae2dfc"
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

cask "pausa" do
  version "1.9.5"

  on_arm do
    sha256 "3e293bac97c0236ad614802d30ab5412f88acfb66f4ea53f30d1b2a6aefe6bd2"
    url "https://github.com/yuseferi/pausa/releases/download/v#{version}/pausa-#{version}-arm64-macos.zip"
  end

  on_intel do
    sha256 "3b006aab3eda3dbba0f19816317e3e4dcad04c7ebcd4979b734b9bb29174600d"
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
  postflight_steps do
    run "/usr/bin/xattr",
        args:         ["-dr", "com.apple.quarantine", "{{appdir}}/pausa.app"],
        must_succeed: true
  end

  zap trash: [
    "~/Library/Application Support/Pausa",
    "~/Library/Logs/Pausa",
  ]
end

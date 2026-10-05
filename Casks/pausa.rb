cask "pausa" do
  version "1.3.1"

  on_arm do
    sha256 "8ff7d26806838abba504136789e5cab8ea35d8f1b8f7fdc32c838823fc307259"
    url "https://github.com/yuseferi/pausa/releases/download/v#{version}/pausa-#{version}-arm64-macos.zip"
  end

  on_intel do
    sha256 "6650847dc4ef5b11041dbd1bbd85640f1506bb35a6e2530c7496041c23d87b6e"
    url "https://github.com/yuseferi/pausa/releases/download/v#{version}/pausa-#{version}-amd64-macos.zip"
  end

  name "Pausa"
  desc "Native-feeling macOS break reminder with fullscreen-space overlays"
  homepage "https://yuseferi.github.io/pausa/"

  app "pausa.app"

  zap trash: [
    "~/Library/Application Support/Pausa",
    "~/Library/Logs/Pausa",
  ]
end

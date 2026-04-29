cask "pausa" do
  version "1.0.0"
  sha256 "443a27cb85bbd11f8fd2a9fb28c5964d38f29f83c3525db3cf8a538c77c9c6bd"

  url "https://github.com/yuseferi/pausa/releases/download/v#{version}/pausa-#{version}-arm64-macos.zip"
  name "Pausa"
  desc "Native-feeling macOS break reminder with fullscreen-space overlays"
  homepage "https://github.com/yuseferi/pausa"

  app "pausa.app"

  zap trash: [
    "~/Library/Application Support/Pausa",
    "~/Library/Logs/Pausa",
  ]
end

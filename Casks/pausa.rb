cask "pausa" do
  version "1.0.4"

  on_arm do
    sha256 "8a976b00f1859399410949d9c23c511f7948dd3f9c5516f67ac5009952931e15"
    url "https://github.com/yuseferi/pausa/releases/download/v#{version}/pausa-#{version}-arm64-macos.zip"
  end

  on_intel do
    sha256 "c96d03f031d404225ea12b39e9d5ed8d26f44edfcf0f9d6f09463dfd0260be41"
    url "https://github.com/yuseferi/pausa/releases/download/v#{version}/pausa-#{version}-amd64-macos.zip"
  end

  name "Pausa"
  desc "Native-feeling macOS break reminder with fullscreen-space overlays"
  homepage "https://github.com/yuseferi/pausa"

  app "pausa.app"

  zap trash: [
    "~/Library/Application Support/Pausa",
    "~/Library/Logs/Pausa",
  ]
end

cask "pausa" do
  version "1.5.0"

  on_arm do
    sha256 "394454d1ee485de6596755f39bbbfd21603a29672fa4a405b109e356b7760e02"
    url "https://github.com/yuseferi/pausa/releases/download/v#{version}/pausa-#{version}-arm64-macos.zip"
  end

  on_intel do
    sha256 "5c2dd9f00dc9a70597a252618943098fcaadd7fd280794cd2b9ac42f5c24b904"
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

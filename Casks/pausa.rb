cask "pausa" do
  version "1.3.0"

  on_arm do
    sha256 "bd450be117296f18989c29d155515afe6a0d5428925f15018e3adf9f28d14c8b"
    url "https://github.com/yuseferi/pausa/releases/download/v#{version}/pausa-#{version}-arm64-macos.zip"
  end

  on_intel do
    sha256 "39118c25c649a01745c13057ca322313d83cb348261fc2e67da5f7d1192cb693"
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

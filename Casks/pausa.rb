cask "pausa" do
  version "1.3.1"

  on_arm do
    sha256 "1c94d1e59860dd602d5e6518b89ee7acb3dc6c5f05545fed5c67fd1991a0fb5e"
    url "https://github.com/yuseferi/pausa/releases/download/v#{version}/pausa-#{version}-arm64-macos.zip"
  end

  on_intel do
    sha256 "a8f03c9d8ce87035cc5df564a1c6e88f1fddea69ea92a27a684c33bf8f693b5f"
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

cask "pausa" do
  version "1.4.0"

  on_arm do
    sha256 "3c6e37bbea563c1454f35eb01c1e5f80a7a98824119bdd79f44487fe31c78789"
    url "https://github.com/yuseferi/pausa/releases/download/v#{version}/pausa-#{version}-arm64-macos.zip"
  end

  on_intel do
    sha256 "f8e1495618e5e4982481124dc88fe573d7b304332247a85774aa26178d251e44"
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

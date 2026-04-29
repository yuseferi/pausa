cask "pausa" do
  version "1.0.1"

  on_arm do
    sha256 "16220b2e70dc9abff51c5e2f3cf01cace6b18e88a2fe14d1c889f046e477ad6b"
    url "https://github.com/yuseferi/pausa/releases/download/v#{version}/pausa-#{version}-arm64-macos.zip"
  end

  on_intel do
    sha256 "528c89e7e01f09cb932531dab789d71185291615a1f9a8ea2534884c530c357a"
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

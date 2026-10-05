cask "pausa" do
  version "1.4.1"

  on_arm do
    sha256 "c9ca0046445067bc1f623d40d0f16f07f33f16e140b8bb8fd05d7fc7cf2bf307"
    url "https://github.com/yuseferi/pausa/releases/download/v#{version}/pausa-#{version}-arm64-macos.zip"
  end

  on_intel do
    sha256 "aca10dc4103b62bb794764f45188e7c0bebe251e3f8dd6d1c52cf6e896bb7cfe"
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

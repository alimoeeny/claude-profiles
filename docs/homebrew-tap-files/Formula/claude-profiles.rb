class ClaudeProfiles < Formula
  desc "Manage multiple Claude Code profiles"
  homepage "https://github.com/alimoeeny/claude-profiles"
  version "0.3.0"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/alimoeeny/claude-profiles/releases/download/v#{version}/claude-profiles_Darwin_arm64.tar.gz"
      sha256 "efed07c623d57c4fca5419bdafdfff822b5e51e08bd2f647fc6ff3db640821b5"
    elsif Hardware::CPU.intel?
      url "https://github.com/alimoeeny/claude-profiles/releases/download/v#{version}/claude-profiles_Darwin_amd64.tar.gz"
      sha256 "eccb04442399b3aca877fec2e375c3d1acef53bd057f1546fe46edebe7ab688c"
    end
  end

  license "MIT"

  def install
    bin.install "claude-profiles"
  end

  test do
    assert_match version.to_s, shell_output("#{bin}/claude-profiles --version")
  end
end

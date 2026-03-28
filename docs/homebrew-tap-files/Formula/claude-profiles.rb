class ClaudeProfileManager < Formula
  desc "Manage multiple Claude Code profiles"
  homepage "https://github.com/alimoeeny/claude-profiles"
  version "0.1.0"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/alimoeeny/claude-profiles/releases/download/v#{version}/claude-profiles_Darwin_arm64.tar.gz"
      sha256 "PLACEHOLDER_ARM64"
    elsif Hardware::CPU.intel?
      url "https://github.com/alimoeeny/claude-profiles/releases/download/v#{version}/claude-profiles_Darwin_amd64.tar.gz"
      sha256 "PLACEHOLDER_AMD64"
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

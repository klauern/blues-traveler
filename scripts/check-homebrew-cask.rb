# Run with `brew ruby scripts/check-homebrew-cask.rb <generated-cask.rb>`.
# Load and exercise the real Homebrew DSL/runner, but never execute commands.
require "cask/cask_loader"
require "simulate_system"

path = ARGV.fetch(0)
source = File.read(path)
raise "Deprecated postflight stanza" if source.match?(/^\s*postflight do/)

class RecordingCommand
  class << self
    attr_accessor :calls

    def run(command, **options)
      calls << [command.to_s, options]
    end
  end
end

[:macos, :linux].product([:arm, :intel]).each do |os, arch|
  Homebrew::SimulateSystem.with(os:, arch:) do
    cask = Cask::CaskLoader::FromContentLoader.new(source).load(config: nil)
    artifacts = cask.artifacts.select { |artifact| artifact.class.dsl_key == :postflight_steps }
    raise "Expected one postflight_steps artifact" unless artifacts.length == 1

    RecordingCommand.calls = []
    Homebrew::InstallSteps::Runner.new(context: cask, command: RecordingCommand).run(artifacts.first.steps)
    calls = RecordingCommand.calls
    if os == :macos
      raise "Expected one macOS command" unless calls.length == 1

      command, options = calls.first
      expected_args = ["-dr", "com.apple.quarantine", "#{cask.staged_path}/blues-traveler"]
      raise "Changed quarantine command" unless command == "/usr/bin/xattr" && options[:args] == expected_args
      raise "Changed failure behavior" unless options[:must_succeed] == true
      raise "Unexpected sudo" unless options[:sudo] == false
    else
      raise "Command must not run on Linux" unless calls.empty?
    end
    puts "PASS #{os}/#{arch}: parsed cask and verified recorded commands"
  end
end

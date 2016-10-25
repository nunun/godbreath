task "default" do
  system "rake -sT"
end

desc "run debug"
task "debug" do
  sh "go get -v"
  sh "go run main.go -v debug"
end

namespace "debug" do
  desc "run help debug"
  task "help" do
    sh "go run main.go -h"
  end

  desc "clean generated files."
  task "clean" do
    sh "rm -f debug/*_gen.go"
  end
end

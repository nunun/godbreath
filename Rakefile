task "default" do
  system "rake -sT"
end

desc "run test."
task "test" do
  sh "go get -v"
  sh "go run main.go test"
end

namespace "test" do
  desc "run help test."
  task "help" do
    sh "go run main.go -h"
  end

  desc "clean generated files."
  task "clean" do
    sh "rm -f test/*_gen.go"
  end
end

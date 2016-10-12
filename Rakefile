task "default" do
  system "rake -sT"
end

desc "clean generated files."
task "clean" do
  sh "rm -f test/*_gen.go"
end

desc "build godbreath and run test."
task "test" do
  sh "(cd cmd/godbreath; go get -v)"
  sh "go run cmd/godbreath/godbreath.go test"
end

namespace "test" do
  desc "run help test."
  task "help" do
    sh "go run cmd/godbreath/godbreath.go -h"
  end
end

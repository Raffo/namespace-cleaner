workflow "New workflow 1" {
  on = "push"
  resolves = ["docker://golang"]
}

action "build binary" {
  uses = "./build-action-1"
}

action "docker://golang" {
  uses = "docker://golang"
  runs = "bash build.sh"
}

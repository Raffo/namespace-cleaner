workflow "New workflow 1" {
  on = "push"
}

action "build binary" {
  uses = "./build-action-1"
}


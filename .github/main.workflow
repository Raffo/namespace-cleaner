workflow "the one that does everything" {
  on = "push"
  resolves = ["docker push"]
}

action "build binary" {
  uses = "./build-action-1"
}

action "Makefiles are the best thing ever" {
  uses = "docker://golang"
  runs = "bash build.sh"
}

action "docker build" {
  uses = "docker://docker:stable"
  args = "build -t x0rg/namespace-cleaner ."
  needs = ["Makefiles are the best thing ever"]
}

action "docker login" {
  needs = ["docker build"]
  uses = "actions/docker/login@master"
  secrets = ["DOCKER_USERNAME", "DOCKER_PASSWORD"]
}

action "docker push" {
  uses = "docker://docker:stable"
  needs = ["docker login"]
  args = "push x0rg/namespace-cleaner"
}

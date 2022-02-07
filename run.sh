#!/bin/zsh
source ~/.zshrc
PATH="/usr/local/bin:$PATH"
/usr/local/bin/env-cmd /Users/ddaniels/.gimme/versions/go1.18beta1.darwin.arm64/bin/go run main.go

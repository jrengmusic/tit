package main

import (
"fmt"
"os"
"tit/internal/git"
)

func main() {
if len(os.Args) < 2 {
fmt.Println("Usage: test_setup <repo_path>")
return
}

if err := os.Chdir(os.Args[1]); err != nil {
fmt.Printf("Failed to cd: %v\n", err)
return
}

state, err := git.DetectState()
if err != nil {
fmt.Printf("Error: %v\n", err)
return
}

fmt.Printf("State: %+v\n", state)
}

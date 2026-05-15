---
description: Install Forma and set up your first Go project.
title: Installation
---

# Installation

## Prerequisites

Forma requires Go 1.22 or newer, so install that first. You'll also want some kind of text editor or IDE to write code and a terminal to run commands.

## Project Setup

Next, open a terminal and create a new Go project, then go get the Forma dependency so it's ready to be imported:

```bash
# Set up a new Go project
$ mkdir my-app
$ cd my-app

# Initialize your project go.mod file
$ go mod init github.com/my-user/my-app
go: creating new go.mod: module github.com/my-user/my-app

# Install Forma
$ go get github.com/tobiasgleiter/forma
...
```

You should now have a directory structure like this:

```
my-app/
  |-- go.mod
  |-- go.sum
```

That's it! Now you are ready to build your first Forma page!

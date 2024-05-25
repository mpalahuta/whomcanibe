# whomcanibe

Dead-simple terminal program that allows you to list and filter AWS profiles available to your user (based on `~/.aws/config` file).
Built using [Bubble Tea](https://github.com/charmbracelet/bubbletea) and [Bubbles](https://github.com/charmbracelet/bubbles).

## Install

Set up the binary by running:
```sh
go install
```

Now you should be able to run the program from any directory:
```sh
whomcanibe
```

If the command doesn't work, make sure your GOBIN is configured correctly.

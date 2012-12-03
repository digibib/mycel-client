# Mycel-client
A rewrite of the [Mycel] client in Go.

## Why?
Go is a compiled lanugage, and thus it allows us to distribute the client as a small executable and discard the Ruby environment. The clients are distributed as Live images over network, and therefore we want them to be as small as possible.

## How to compile
You need the Go plattform. See the official golang site for [installation instructions]

In addition, you need the GTK development headers:

    sudo apt-get install libgtk2.0-dev libgtksourceview2.0-dev

Then fetch and compile the client using the go command:

    go get github.com/digibib/mycel-client
    go build client.go

[Mycel]: https://github.com/digibib/mycel
[installation instructions]: http://golang.org/doc/install
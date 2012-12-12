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
    go build client.go config.go

The go tool can output target binaries to all major platforms. You set the target architecture by modifying the Go environment variables. Note that you may need additional libraries if you are compiling to another platform than your existing environment. For example, to compile a 32-bit linux binary on a 64-bit system, you may need the following:

    sudo apt-get install libc6-dev-i386 ia32-libs-gtk gcc-multilib

Cross-compiling can be quite complicated. If you can't make it work, just compile it on the target platform.

[Mycel]: https://github.com/digibib/mycel
[installation instructions]: http://golang.org/doc/install
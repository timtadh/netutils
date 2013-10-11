netutils
========

By Tim Henderson (tim.tadh@gmail.com)

This is a little library that was part of a larger project that I decided to
pull out and make public. It is kinda a replacement for `chan" in the sense
that it allows you to treat a TCP connection as a pair of `chan`s. However,
it is not as ambitious as `netchan` was. It doesn't attempt to solve any sort of
serialization for you instead it is a `[]byte` oriented interface.


How It Works
============

It has two important methods which can take a TCP connection and turn it into a
pair of channels. The first is `TCPReader`:

    func TCPReader(con *net.TCPConn) (recv <-chan byte) 

TCPReader emits a `chan byte` which you can use to read the TCP connection in a
byte oriented manner. Internally, it is maintaining a buffer but externally it
presents it as a smooth stream. If you want to process the connection in a line
or other delimiter oriented way there are utility methods to assist you.

The second method is `TCPWriter` which is almost the inverse of `TCPReader`:

    func TCPWriter(con *net.TCPConn) (send chan<- []byte)

Instead of creating a byte oriented channel it creates a `[]byte` oriented
channel. This maps more smoothly to how sending is usually done, by sending
preconstructed packets of data. If you are interested in sending a free form
stream of bytes you probably want to use the TCP connection object natively.


Docs
====

```
PACKAGE DOCUMENTATION

package netgrid
    import "github.com/timtadh/netutils"



VARIABLES

var DEBUG bool = false


FUNCTIONS

func IsEOF(err error) bool
    A utility function which tests if an error returned from a TCPConnection
    or TCPListener is actually an EOF. In some edge cases this which should
    be treated as EOFs are not returned as one.

func ReadDelim(recv <-chan byte, delim byte) (line []byte, EOF bool)
    A blocking read operation on a byte channel. It reads up to a delimiter
    and returns a byte slice of the bytes read (not including the delim). If
    it encounters an EOF (a closed channel) it returns true on EOF and
    whatever was read before the EOF.

func ReadDelims(recv <-chan byte, delim byte) (linechan <-chan []byte)
    ReadDelims is a nonblocking reader on a byte channel. It consumes the
    channel (eg. will read until the recv channel closes). You should not do
    other reads on the recv channel if using ReadDelims.

    ReadDelims reads until a deliminter it encountered. It then emits a byte
    slice which does not include the deliminter. The sending channel will be
    closed when the recieving channel is closed.

func Readline(recv <-chan byte) (line []byte, EOF bool)
    A block read operation which reads one line from the byte channel. It
    used ReadDelim(recv, '\n') under the hood.

func Readlines(recv <-chan byte) (linechan <-chan []byte)
    Readlines is a nonblock reader on a byte channel. It reads all the lines
    from the recieving channel. See the ReadDelims documentation for caveats
    on usage (as it uses ReadDelims under the hood).

func TCPReader(con *net.TCPConn) (recv <-chan byte)
    A TCPReader takes a tcp connection and transforms it into a readable
    byte channel. It doesn't read []byte since it doesn't know what
    delimiters you want to use. If you want to consume the byte channel as a
    series of lines try "Readlines" if you have another delimiter ('\0' or
    similar) try ReadDelims.

func TCPWriter(con *net.TCPConn) (send chan<- []byte)
    A TCPWriter takes a tcp connection and transforms it into a writable
    []byte channel. You can use this to build more "goish" tcp servers.
```


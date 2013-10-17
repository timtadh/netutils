package netutils

/* Author: Tim Henderson
 * Email: tadh@case.edu
 * Copyright 2013 All Right Reserved
 *
 * Redistribution and use in source and binary forms, with or without
 * modification, are permitted provided that the following conditions are met:
 * 
 *  * Redistributions of source code must retain the above copyright notice,
 *    this list of conditions and the following disclaimer.
 *
 *  * Redistributions in binary form must reproduce the above copyright notice,
 *    this list of conditions and the following disclaimer in the documentation
 *    and/or other materials provided with the distribution.
 *
 *  * Neither the name of the netutils nor the names of its contributors may be
 *    used to endorse or promote products derived from this software without
 *    specific prior written permission.
 * 
 * THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
 * AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
 * IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE
 * ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE
 * LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR
 * CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF
 * SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS
 * INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN
 * CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE)
 * ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE
 * POSSIBILITY OF SUCH DAMAGE.
 */


import (
  "os"
  "io"
  "net"
  logpkg "log"
)


var DEBUG bool = false
var log *logpkg.Logger

func init() {
    log = logpkg.New(os.Stderr, "netutils>", logpkg.Ltime | logpkg.Lshortfile)
}

/*
A utility function which tests if an error returned from a TCPConnection or
TCPListener is actually an EOF. In some edge cases this which should be treated
as EOFs are not returned as one.
*/
func IsEOF(err error) bool {
    if err == nil {
        return false
    } else if err == io.EOF {
        return true
    } else if oerr, ok := err.(*net.OpError); ok {
        /* this hack happens because the error is returned when the
         * network socket is closing and instead of returning a
         * io.EOF it returns this error.New(...) struct. */
        if oerr.Err.Error() == "use of closed network connection" {
            return true
        } else if oerr.Err.Error() == "connection reset by peer" {
            return true
        }
    } else if err.Error() == "use of closed network connection" {
        return true
    }
    return false
}

/*
A TCPWriter takes a tcp connection and transforms it into a writable []byte
channel. You can use this to build more "goish" tcp servers.
*/
func TCPWriter(con *net.TCPConn) (send chan<- []byte) {
    comm := make(chan []byte)
    go func(recv <-chan []byte) {
        for block := range recv {
            if DEBUG {
                log.Println("TCPWriter got a block", string(block))
            }
            for n, err := con.Write(block);
                n < len(block) || err != nil;
                n, err = con.Write(block[n:]) {
                  if err != nil {
                      log.Panic(err)
                  }
            }
            if DEBUG {
                log.Println("TCPWriter sent the block", string(block))
            }
        }
        if err := con.Close(); err != nil {
            log.Println(err)
        }
    }(comm)
    return comm
}

/*
A TCPReader takes a tcp connection and transforms it into a readable byte
channel. It doesn't read []byte since it doesn't know what delimiters you want
to use. If you want to consume the byte channel as a series of lines try
"Readlines" if you have another delimiter ('\0' or similar) try ReadDelims.
*/
func TCPReader(con *net.TCPConn) (recv <-chan byte) {
    comm := make(chan byte)
    go func(send chan<- byte) {
        memblock := make([]byte, 256)
        var EOF bool
        for !EOF {
            n, err := con.Read(memblock)
            if IsEOF(err) {
                EOF = true
            } else if err != nil {
                log.Panic(err)
            }
            if DEBUG && n > 0 {
                log.Println("TCPReader got block", string(memblock[:n]))
            }
            for i := 0; i < n; i++ {
                send<-memblock[i]
            }
        }
        close(send)
    }(comm)
    return comm
}

/*
A blocking read operation on a byte channel. It reads up to a delimiter and
returns a byte slice of the bytes read (not including the delim). If it
encounters an EOF (a closed channel) it returns true on EOF and whatever was
read before the EOF.
*/
func ReadDelim(recv <-chan byte, delim byte) (line []byte, EOF bool) {
    for char := range recv {
        if char == delim {
            return line, false
        } else {
            line = append(line, char)
        }
    }
    return line, true
}

/*
A block read operation which reads one line from the byte channel. It used
ReadDelim(recv, '\n') under the hood.
*/
func Readline(recv <-chan byte) (line []byte, EOF bool) {
    return ReadDelim(recv, '\n')
}

/*
ReadDelims is a nonblocking reader on a byte channel. It consumes the channel
(eg. will read until the recv channel closes). You should not do other reads on
the recv channel if using ReadDelims.

ReadDelims reads until a deliminter it encountered. It then emits a byte slice
which does not include the deliminter. The sending channel will be closed when
the recieving channel is closed.
*/
func ReadDelims(recv <-chan byte, delim byte) (linechan <-chan []byte) {
    lines := make(chan []byte)
    go func() {
        var line []byte
        for char := range recv {
            if char == delim {
                lines<-line
                line = make([]byte, 0, 80)
            } else {
                line = append(line, char)
            }
        }
        close(lines)
    }()
    return lines
}

/*
Readlines is a nonblock reader on a byte channel. It reads all the lines from
the recieving channel. See the ReadDelims documentation for caveats on usage (as
it uses ReadDelims under the hood).
*/
func Readlines(recv <-chan byte) (linechan <-chan []byte) {
    return ReadDelims(recv, '\n')
}


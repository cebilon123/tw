Hey! Nice to e-meet you 😁

So, this is simple application that observes given blockchain
address and saves transactions related to this address.

I have not used context for a lot of the stuff, becuase
the parser interface was not letting me to do that.

I have also used interface for storage, and currently its
only implemented by the MemoryStorage (which means it can
be in the future replaced with other storage).

There is only one unit test because it was told to be done
with the task in 4h (In normal scenario I would write
a lot more of those)

The application can be seen in the `example/sample.go`, it
logs some stuff to console but it can be treated as some
kind of playground.

I have created the PKG directory because it was told that
this Parser should be used as the public interface, and 
the internal one isn't available i.e. if you will get the
package to your own GO project.

Also created simple command line logger, but it can be
replaced with other one as long as it will implement the
`io.Writer` interface.

Also code abstractions are implementing the `io.Closer`
to safely close the stuff behind the scene. (especially 
the observer as well as `ethereum.go` parser)

### Why 1 unit test?
Basically I didn't have a time, I was forced to leave home 😂

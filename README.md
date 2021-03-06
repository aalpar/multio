# MultiplexReader

MultiplexReader duplicates read operations through to many readers.

This library is designed to distribute one read stream into many duplicated and independent read streams.  Its application is memory efficient data replication.  MultiplexReader will only consume as much buffer memory as is necessary for its slowest stream.  Lagging streams can be closed allowing their unshared buffers to be reclaimed.

# Problem Description

When replicating data it's tempting to copy the entire item into memory and then distribute it among replicas.  This has the benefit of a simple implementation and few failure modes.  The approach works well when objects are known to have an upper limit in size and fit well within memory constraint but the approach has several limitations:

- Memory consumption and timing varies with the size of the object being replicated.
- Failures may not be detected until the item is sunk, in which case dependent prior resources were consumed for a failure that could have possibly been detected earlier.

The solution presented in this library breaks the same replication operation into pieces.  Breaking the replication down has some benefits over copy of the entire item as a unit:

- An upper bound on the memory used for replication can be set independent of the size of the object being replicated.
- The problem can be unitized - and resource costs calculated for a unit (memory, network, time).

# The Library

This library uses the io.Reader interface to model replication as a tree of readers - one reader is supplied as the source of data; other readers are created as sinks for the replicated data.

Any `Read([]byte)` call on one of the sinks may produce a `Read([]byte)` call on the source.  Each source read (independent of the sink that initiated it) is replicated to all sinks using channels.

The process of replication to sinks is single threaded and may block waiting for channel slots to become free.  The process is single threaded to support an upper bound on resource consumption and predictable operation.

Blocking channels can be closed to free resources associated with the sink reader.  This also unblocks any waiters on the channel - thus freeing stalls caused by channels reaching capacity.  All sink readers should be closed after use so that channels without consumers are not left dangling, possibly blocking other read operations.



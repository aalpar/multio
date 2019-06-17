# DuplexMultiReader

DuplexMultiReader duplicates read operations through to many readers.

This library is designed to distribute one read stream into many duplicated and independant read streams.  Its application is memory effiecient data replication.  DuplexMultiReader will only consume as much buffer memory as is necessary for its slowest stream.  Lagging streams can be closed allowing their unshared buffers to be reclaimed.


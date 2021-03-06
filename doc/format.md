# Conserve format

## Generalities

All metadata is stored as json dictionaries.


## Software version

Conserve archives include the version of the software that wrote them, which is
an _x.y.z_ tuple.  See [versioning.md](versioning.md) for the semantics.


## Filenames

Files have names in the source and restore directories, and within the archive.

In source and restore directories, file naming is defined by the OS: on Windows as UTF-16,
on OS X as UTF-8 and on Linux as an arbitrary 8-bit encoding.

(Linux filenames are very commonly UTF-8, but there are important exceptions: users who
choose to use different encodings for whole filesystems; network or USB filesystems
using different encodings; files is source trees that are intentionally in odd encodings; and
files that accidentally have anomalous names.  It is useful to include the occasionally
oddly-named file in the backup, and also for users with non-UTF-8 encodings to be able to
configure this. The filename encoding is not easily detectable.  Linux does require that the
separator `/` have the same byte value.)

In the archive, Conserve uses "apaths" as a platform-independent path format.

apaths are stored as UTF-8 byte strings. UTF-8 filenames are
stored as received from the OS with no normalization.

apaths always have `/` separators and start with a `/`.

None of the apath components can be `.`, `..`, or empty.

The apath `/` within the archive is the top source directory.

Filenames are treated as case-sensitive.

## Filename sorting

There is a total order between filenames.  Within a band, entries are
stored in this order.

This ordering allows binary search for files within the index, and to
efficiently list the direct or recursive contents of a directory.

The ordering puts all the direct contents of a directory together, followed
by those of each of its children.

The order is defined as: split the filenames into a directory part
and a non-empty tail part.  Compare by the directory first using a
byte-by-byte comparison of their (typically UTF-8) byte string form.
Then, similarly compare the filenames.

Note that this is not the same as a simple comparison of the strings.

Rationale: This ordering makes several important read operations on the index
efficient.  Since the index overall is ordered, we can binary search into it.
Since all the direct children of a directory are grouped together, a
non-recursive listing of a single directory (e.g. from a web ui or FUSE) is a
single contiguous read.  And, a recursive listing of a single directory (e.g.,
to restore that directory) is also a single contiguous read of the directories
and then all their subtrees.

## Archive

A backup *archive* is a directory, containing archive files.

Archives can be stored on cloud or other remote storage.
The archive makes minimal assumptions about the filesystem it's stored on: in
particular, it need not support case sensitivity, it need not store times or
other metadata, and it only needs to support 8.3 characters.  It must supported
nested subdirectories with a total path length up to 100 characters.

Archive filesystems must allow many files per directory.

## Archive header

In the root directory of the archive there is a file called `CONSERVE`,
which is contains a json dict:

    {"conserve_archive_version":"0.3"}

## Tiers

Within an archive there are multiple *tiers* for incremental/hierarchical
backups.  (For example, they might be annual, monthly, weekly, daily, and
hourly backups.)  Tiers are not directly represented on disk; they're
implicitly all the bands whose names identify them as being in the same tier.

## Bands

Within each tier, there are multiple *bands*.  (For example, "the monthly
backup made on 2012-10-01.")  A band that is not in the base tier has a
*parent band* in the immediately lower tier.

A band can be *open*, while it is receiving data, or *closed* when
everything from the source has been written.  Bands may remain open
indefinitely, across multiple Conserve invocations, until they are finished.
Once the band is finished, it will not be changed.

Bands are numbered hierarchically across tiers and sequentially within
a tier, starting at 0.  So the first base tier band in the whole archive
is `0000`, the first incremental band on top of it is `0000-0000`,
and so on.  The numbers are zero-padded to four digits in each
area, so that they will be grouped conveniently for humans looking at
naively sorted listings of the directory.  (Conserve does not rely on them
being less than five digits, or on the transport returning any particular
ordering; bands numbered over 9999 are supported.)

Band directories contain a description of files that changed, or were deleted,
relative to their ancestor bands.  A copy of the source directory at a
particular time can be extracted by reading the closest band, plus all of its
parents.

Bands are represented as a subdirectory within the archive directory,
as `b` followed by the number.  All bands are in the top-level directory.

    my-archive/
      b0000/
      b0000-0000/
      b0000-0001/
      b0000-0001-0000/

## Band head

A band head is a file `BANDHEAD` containing a json dictionary.

The head file is written when the band is first opened and then it is
not changed again.

The head file contains:

 - `start_time`: the Unix time the band was started

## Band tail

A band tail is a file `BANDTAIL` containing a json dictionary, only for
finished bands: it is the presence of this file that defines the band as
closed.

Band footer contains:

 - `end_time`: the Unix time the band started and ended


## Data blocks

Data blocks contain parts of the full text of stored files.

One data block may contain data for a whole file, the concatenated
text for several files, or part of a file.  One data block
may be referenced from the index block of any number of files
from the current or any descendent band.

The writer can choose the data block size, except that both the uncompressed
and compressed blocks must be <1GB, so they can reasonably fit in memory.
The writer might choose to break the file not at a fixed size but instead at
some boundary it thinks will be stable as the file changes, for example using
an rsync-like rolling checksum.

All the data block for a band are stored within a `d/` subdirectory
of the band, and then within a directory for the first three characters
of their name.

Data block are compressed in the Brotli format.  (Note: This is likely to
change to a different compression algorithm post 0.3 because Brotli is slow to
compress.)

The name of the data block file is the BLAKE2 hash of the uncompressed
contents.


## Index hunks

Index hunks contain the name and metadata of a stored file, plus a
reference to the data hunks holding its full text.

Index hunks are named with decimal sequence numbers padded to 9 digits.

Index hunks are stored in an `i/` subdirectory of the band, and then
in a subdirectory for the sequence number divided by 10000 and
padded to five digits.  So, the first block is `i/00000/000000000`.

Index hunks are stored in json and gzip compressed.

Stored files are in order by filename across all of the index hunks
within a band.

The number of files described within a single index hunk file is
arbitrary and may be chosen to control the number of outstanding data
blocks or the length of the index hunk.

The uncompressed index hunk contains a json list each element of
which is a dict of

   - `apath`: the name of the file
   - `mtime`: in seconds past the unix epoch
   - ownership, permissions, and other filesystem metadata
   - `kind`: one of `"File"`, `"Dir"`, `"Symlink"`
   - `deleted`: true if it was present in a parent band and was
     deleted in this band
   - `blake2`: the BLAKE2 hash in hex of the full text of the file
   - `addrs`: a list of tuples of:
     - `hash`: data block hash: from the current or any
       parent directory
     - `start`: the offset within the uncompressed content of the
       block for the start of this file
     - `length`: the number of bytes of uncompressed data block
       content to store in this file
     `target`: For symlinks, the string target of the symlink.

So, the length of any file is the sum of the `length` entries for all
its `blocks`.

Index blocks can reference any section of any data block in the current
or any ancestor tier band, but not sibling, descendent, or unrelated bands.
The writer may deduplicate blocks or partial blocks against any of these data
blocks using any algorithm, including referencing a different length or offset
into the same block.

Rationale: Constraining which indexes can reference a data block allows
Conserve to purge a band and its children without needing to consider whether
its data is used elsewhere.  This means that the purge operation takes time
proportional only to the data being purged, and does not need to walk the whole
archive or any other index.

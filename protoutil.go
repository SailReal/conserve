// Conserve - robust backup system
// Copyright 2012-2013 Martin Pool
//
// This program is free software; you can redistribute it and/or
// modify it under the terms of the GNU General Public License
// as published by the Free Software Foundation; either version 2
// of the License, or (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.

package conserve

import (
    "os"

    "code.google.com/p/goprotobuf/proto"
)

func ReadProtoFromFile(message proto.Message, filename string) (err error) {
    f, err := os.Open(filename)
    if err != nil {
        return
    }
    fi, err := f.Stat()
    if err != nil {
        return
    }
    buf := make([]byte, fi.Size())
    _, err = f.Read(buf)
    // NB: This doesn't check that we read the whole file, but probably the
    // decode will fail in that case.
    if err != nil {
        return
    }
    err = proto.Unmarshal(buf, message)
    return
}

func WriteProtoToFile(message proto.Message, filename string) (err error) {
    bytes, err := proto.Marshal(message)
    if err != nil {
        return
    }

    f, err := os.Create(filename)
    if err != nil {
        return
    }

    _, err = f.Write(bytes)
    if err != nil {
        f.Close()
        return
    }

    err = f.Close()
    return
}
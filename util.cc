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

#include <sys/types.h>
#include <sys/stat.h>
#include <unistd.h>
#include <fcntl.h>

#include <boost/filesystem.hpp>

#include <glog/logging.h>

#include <google/protobuf/text_format.h>
#include <google/protobuf/io/zero_copy_stream_impl.h>

#include "util.h"

namespace conserve {

using namespace std;

using namespace boost;

using namespace google::protobuf::io;
using namespace google::protobuf;

void write_proto_to_file(
        const Message& message,
	const filesystem::path& path) {
    int fd = open(path.string().c_str(),
	    O_CREAT|O_EXCL|O_WRONLY,
	    0666);
    PCHECK(fd > 0);
    CHECK(message.SerializeToFileDescriptor(fd));
    int ret = close(fd);
    PCHECK(ret == 0);
}


} // namespace conserve

// vim: sw=4 et

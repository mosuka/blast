// Copyright 2014 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

#ifndef CLD2_INTERNAL_CLD2_DYNAMIC_COMPAT_H_
#define CLD2_INTERNAL_CLD2_DYNAMIC_COMPAT_H_

// open(), close(), mmap() and munmap() are not available in vanilla win32.
// This header provides compatibility for different operating systems using
// standard preprocessor definitions.
// Note that _WIN32 is also defined on 64-bit platforms :)
//
// For more information see https://code.google.com/p/cld2/issues/detail?id=19

#ifdef _WIN32
  #include <io.h>
  #define OPEN _open
  #define CLOSE _close
#else // E.g., POSIX. We don't try to support Mac versions prior to OSX.
  #include <sys/mman.h>
  #include <fcntl.h>
  #include <unistd.h>
  #define OPEN open
  #define CLOSE close
#endif

#endif  // CLD2_INTERNAL_CLD2_DYNAMIC_COMPAT_H_

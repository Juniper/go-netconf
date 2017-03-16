# libyang

[libyang](https://github.com/CESNET/libyang) is YANG data modeling language parser and toolkit written (and
providing API) in C. The library is used e.g. in [libnetconf2](https://github.com/CESNET/libnetconf2),
[Netopeer2](https://github.com/CESNET/Netopeer2) or [sysrepo](https://github.com/sysrepo/sysrepo) projects.

## Building libyang

You can simply install locally libyang with executing the following steps.

```
$ git clone https://github.com/CESNET/libyang.git
$ cd libyang
$ mkdir build; cd build
$ cmake ..
$ make
# make install
```

If you want to create a static libyang binary too, add "-DSTATIC=true" to cmake.

```
$ cmake -DSTATIC=true ..
```

## Include libyang into your project

To use the libyang library in your go project add these lines of code under your import statement.

```
/*
#cgo LDFLAGS: -lyang
#include <libyang/libyang.h>
*/
import "C"

```

For a static build also add pcre or you will get errors.

```
#cgo LDFLAGS: -lpcre
```

## Compile your code

To create a binary that is dynamically linked to libyang simply run:

```
go build
```

For a static build execute "go build" with additional flags:
```
go build --ldflags '-extldflags "-static"'
```

## Compile the example with docker

If you don't want to install locally libyang you can use docker for a static or dynamic build. You have two bash scripts which will build a docker container, compile the code and close docker.

```
$ ./build_static.sh
# or
$ ./build_dynamic.sh
```


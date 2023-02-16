# Integration Testing

To validate the changes as well as test against different vendor hardware there is some integration testing contained in this directory.   These tests are supposed to be somewhat vendor agnostic and should be able to run against any vendor.

> **Note**
> Currently these are basic tests and no write operations, however this may change and will be controlled with another enviroment variable to do testing.

To run against an existing device (also known as Device Under Test or DUT) you can specify the following enviroment variables and run the test enviroment:

| Variable              | Use                                                  |
| --------------------- | ---------------------------------------------------- |
| `NETCONF_DUT_SSHHOST` | The hostname to test against                         |
| `NETCONF_DUT_SSHPORT` | The port number of the ssh server (default to 830)   |
| `NETCONF_DUT_SSHUSER` | The username for authentication                      |
| `NETCONF_DUT_SSHPASS` | The password for authentication                      |
| `NETCONF_DUT_FLAVOR`  | **NOT YET USED** The operating system flavor to be used for advanced testing.  One of `junos`, or `eos` |

```plain
$ cd inttest
$ NETCONF_DUT_SSHHOST=mydut \
  NETCONF_DUT_SSHPORT=22 \
  NETCONF_DUT_SSHUSER=root \
  NETCONF_DUT_SSHPASS=juniper123 \
  NETCONF_DUT_FLAVOR=junos \
  go test .
```

## Container tests

In addition to one off testing there are container tests to run against containerized DUTs.  This allows for continuos integration testing on PRs and pushes to the main branch.  Unfortunatly we still live in a world where vendors hide their container images behind login walls and pay walls which means leagly obtaining these requires jumping through hoops.

### Juniper cSRX

Juniper's cSRX supports very basic netconf support.  It only supports SSH (no TLS), runs on port 22 only and doesn't support any advanced features like call home, etc.   For testing there is a demo image that can be downloaded and used.

> **Note**
> Juniper also has a cRPD but this image doesn't support netconf at all.

To install the image use the following steps:

1. Download the docker cSRX image from <https://support.juniper.net/support/downloads/?p=csrxeval> (login requiered).  *At the time of writing the latest eval version is 20.2R1*)
2. Run `docker image load -i junos-csrx-docker-20.3.R1.8.img`  This will add the image `csrx:20.3R1.8` to your local images.

You can now run the csrx integration tests using Docker:

```plain
cd inttest
CSRX_IMAGE=csrx:20.3R1.8 just csrx
```

### Arista cEOS-lab

Arista provides cEOS-lab images to be used for lab testing.  Don't conuse with cEOS which is similar but made for production hardware enviroments.

Like cSRX the image is locked away behind a login.  You can find all downloades at <https://www.arista.com/en/support/software-download>. *At the time of this writing the latest M release is cEOS64-lab-4.28.3M.  Other versions may or may not work similarly.  Pleae feel free to test!*

> **Note**
> You may have to be an Arista customer to be able to download the image?

1. Download desired EOS lab image from <https://www.arista.com/en/support/software-download> (login required).  You can download either the 64bit or non-64bit images.
2. Run the following command to import the image locally: `docker import --platform linux/am64 cEOS64-lab-4.28.3M.tar.xz ceos64-lab:4.28.3M`

> Note this will fail with M1 Macs (or non-amd64 systems): See https://github.com/docker/for-mac/issues/6361.

You can then run the ceos integration tests using Docker:

```plain
cd inttest
CEOS_IMAGE=ceos64-lab:4.28.3M just ceos
```

### tail-f (Cisco) confD

tail-f confd is a service/framework for adding NETCONF and other protocols to existing systems (like network devices).  There is a basic version available for download with a Cisco CCO loging that can be used for testing.  This is not already in a Docker container so `Dockerfile.confd' is available to create a docker image

1. Download confd from <https://developer.cisco.com/site/confD/> (login reqired).  You will want the linux x86_64 images.  At the time of writing the latest is `confd-basic-7.8.3.linux.x86_64.zip`.  Place this file in the same folder as Dockerfild.confd.
2. Create the docker image `docker build -f Dockerfile.confd . -t confd-basic:7.8.`

Now the integration tests will work

```plain
cd inttest
CONFD_IMAGE=conf-basic:7.8 just ceos
```

### netopeer2

Netopeer2 is an opensource NETCONF server.  A docker image is automaticall built when running the tests so no additional work is needed.

```plain
cd inttent
just netopeer2
```

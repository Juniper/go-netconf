FROM sysrepo/sysrepo-netopeer2:latest

MAINTAINER mislav.novakovic@sartura.hr

# install wget
RUN \
      apt-get install -y wget

# install golang
RUN \
      wget https://storage.googleapis.com/golang/go1.7.3.linux-amd64.tar.gz && \
      tar -zxvf go1.7.*.linux-amd64.tar.gz -C /usr/local

#set golang environment variables
ENV GOROOT /usr/local/go
ENV GOPATH /opt/dev/go
ENV PATH $PATH:$GOROOT/bin

# get go-netconf
RUN \
      go get github.com/sartura/go-netconf/netconf


# add static build to libyang
RUN \
      cd /opt/dev/libyang && \
      cd build && \
      cmake -DSTATIC=true .. && \
      make -j2 && \
      make install && \
      ldconfig

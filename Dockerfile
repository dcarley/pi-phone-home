FROM dcarley/golang-rpi

RUN apt-get -y install software-properties-common && \
  add-apt-repository ppa:snappy-dev/beta && \
  apt-get -y update
RUN apt-get -y install snappy-tools

FROM docker.mirror.hashicorp.services/ubuntu:16.04

ENV DEBIAN_FRONTEND noninteractive

RUN apt-get update && apt-get install -y \
    locales \
    openssh-server \
    sudo

RUN locale-gen en_US.UTF-8

RUN if ! getent passwd vagrant; then useradd -d /home/vagrant -m -s /bin/bash vagrant; fi \
    && echo 'vagrant ALL=(ALL) NOPASSWD:ALL' >> /etc/sudoers \
    && mkdir -p /etc/sudoers.d \
    && echo 'vagrant ALL=(ALL) NOPASSWD:ALL' >> /etc/sudoers.d/vagrant \
    && chmod 0440 /etc/sudoers.d/vagrant

RUN mkdir -p /home/vagrant/.ssh \
    && chmod 0700 /home/vagrant/.ssh \
    && wget --no-check-certificate \
      https://raw.github.com/hashicorp/vagrant/master/keys/vagrant.pub \
      -O /home/vagrant/.ssh/authorized_keys \
    && chmod 0600 /home/vagrant/.ssh/authorized_keys \
    && chown -R vagrant /home/vagrant/.ssh

RUN mkdir -p /run/sshd

CMD /usr/sbin/sshd -D \
    -o UseDNS=no \
    -o PidFile=/tmp/sshd.pid

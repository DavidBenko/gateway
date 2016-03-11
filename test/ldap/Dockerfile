FROM ubuntu:14.04

RUN apt-get -y update && apt-get install -y \
      slapd \
      ldap-utils

ADD ./slapd.conf /root
ADD ./security/servercrt.pem /root
ADD ./security/serverkey.pem /root
ADD ./security/cacert.pem /root

EXPOSE 389
ENTRYPOINT /usr/sbin/slapd -f /root/slapd.conf -d 255

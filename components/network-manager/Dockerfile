#
# Copyright (C) 2018 Nalej Group - All Rights Reserved
#

# iron/go is the alpine image with only ca-certificates added
FROM zerotier/zerotier-containerized:1.2.4 as zt_container

ADD dockerenv/startup.sh /
RUN chmod 0755 /startup.sh
EXPOSE 9993/udp

# Now just add the binary\
RUN mkdir /nalej

# Now just add the binary
COPY network-manager /nalej/

ENTRYPOINT ["./startup.sh"]
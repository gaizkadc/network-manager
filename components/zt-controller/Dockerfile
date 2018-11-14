#
# Copyright (C) 2018 Nalej Group - All Rights Reserved
#

FROM zerotier/zerotier-containerized:1.2.4 as zt_container

ADD dockerenv/startup.sh /
RUN chmod 0755 /startup.sh
EXPOSE 9993/udp

# RUN mkdir -p /var/lib/zerotier-one/ \
#    && chown 2000:2000 /var/lib/zerotier-one/ \
#    && chmod 0755 /var/lib/zerotier-one/

ENTRYPOINT ["./startup.sh"]
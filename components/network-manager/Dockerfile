FROM debian:stretch

RUN apt-get update \
    && apt-get install -y curl gpg \
    && curl -s https://install.zerotier.com/ | bash \
    && apt-get autoremove --purge -y \
    && rm -rf /var/lib/apt/lists/*

ADD dockerenv/local.conf /var/lib/zerotier-one
ADD dockerenv/startup.sh /
RUN chmod 0755 /startup.sh
EXPOSE 9993/udp

COPY network-manager /nalej/

ENTRYPOINT ["./startup.sh"]
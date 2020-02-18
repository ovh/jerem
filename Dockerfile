FROM debian:9
RUN apt update
RUN apt install -y ca-certificates
ADD ./bin/jerem /jerem
COPY ./config.yml /config.yml
CMD ["/jerem", "--config", "/config.yml"]
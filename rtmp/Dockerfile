FROM alpine:3.13.4 AS builder
RUN apk add --update build-base git bash gcc make g++ zlib-dev linux-headers pcre-dev openssl-dev
RUN git clone https://github.com/arut/nginx-rtmp-module.git && \
    git clone https://github.com/nginx/nginx.git
RUN cd nginx && ./auto/configure --add-module=../nginx-rtmp-module && make && make install

FROM alpine:3.13.4 AS nginx
RUN apk add --update pcre ffmpeg
COPY --from=builder /usr/local/nginx /usr/local/nginx

COPY ./nginx.conf /usr/local/nginx/conf/nginx.conf

COPY ./run.sh /usr/local/nginx/conf/run.sh
RUN chmod +x /usr/local/nginx/conf/run.sh

CMD /usr/local/nginx/conf/run.sh
# ENTRYPOINT ["/usr/local/nginx/sbin/nginx"]
# CMD ["-g", "daemon off;"]
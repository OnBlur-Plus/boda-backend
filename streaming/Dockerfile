FROM ossrs/srs:5 AS srs

RUN mv /usr/local/srs/objs/ffmpeg/bin/ffmpeg /usr/local/bin/ffmpeg && \
    ln -sf /usr/local/bin/ffmpeg /usr/local/srs/objs/ffmpeg/bin/ffmpeg

RUN rm -rf /usr/local/srs/objs/nginx/html/console \
    /usr/local/srs/objs/nginx/html/players

FROM ossrs/srs:ubuntu20 AS build

COPY --from=srs /usr/local/srs /usr/local/srs

ADD ./ /usr/local/streaming

WORKDIR /usr/local/streaming

RUN make clean && make

# Use UPX to compress the binary.
ENV DEBIAN_FRONTEND=noninteractive
RUN apt-get update -y && apt-get install -y upx

RUN echo "Before UPX" && \
    ls -lh /usr/local/srs/objs/srs /usr/local/streaming/streaming && \
    upx --best --lzma /usr/local/srs/objs/srs && \
    upx --best --lzma /usr/local/streaming/streaming && \
    echo "After UPX for" && \
    ls -lh /usr/local/srs/objs/srs /usr/local/streaming/streaming

#FROM ubuntu:focal AS dist
FROM ossrs/oryx:focal-1 AS dist

EXPOSE 2022 2443 1935 8080 5060 9000 8000/udp 10080/udp

# Copy files from build.
COPY --from=build /usr/local/streaming /usr/local/streaming
COPY --from=build /usr/local/srs /usr/local/srs

# Prepare data directory.
RUN mkdir -p /data && \
    cd /usr/local/streaming/containers && \
    rm -rf data && ln -sf /data .

CMD ["./bootstrap"]
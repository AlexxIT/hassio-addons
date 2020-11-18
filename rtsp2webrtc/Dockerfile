ARG BUILD_FROM
FROM $BUILD_FROM

RUN apk add --no-cache unzip go jq

WORKDIR /app

COPY app /app

RUN go build

# RUN curl -L https://github.com/deepch/RTSPtoWebRTC/archive/master.zip -o master.zip \
#  && unzip master.zip \
#  && mv RTSPtoWebRTC-master/* . \
#  && go build \
#  && rm -R RTSPtoWebRTC-master master.zip

COPY run.sh /
RUN chmod a+x /run.sh

CMD [ "/run.sh" ]
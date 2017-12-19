FROM busybox

COPY ./http_logger /http_logger

EXPOSE 8080
ENTRYPOINT [ "/http_logger" ]
CMD []

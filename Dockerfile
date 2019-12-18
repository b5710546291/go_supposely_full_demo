FROM alpine:3.7
COPY go_supposely_full_demo /usr/local/

EXPOSE 9001

RUN ["chmod", "+x", "usr/local/go_supposely_full_demo"]
CMD usr/local/go_supposely_full_demo


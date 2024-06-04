FROM ghcr.io/linuxsuren/api-testing:master

WORKDIR /workspace
COPY e2e/* .
RUN chmod +x entrypoint.sh

CMD [ "/workspace/entrypoint.sh" ]

# Dockerfile for manifest tool query action as wsk action
FROM openwhisk/dockerskeleton
 
RUN apk --no-cache add ca-certificates jq && update-ca-certificates

### Add manifest tool static binary
ADD manifest-tool-linux-amd64 /bin/manifest-tool
ADD exec /action/exec

CMD ["/bin/bash", "-c", "cd actionProxy && python -u actionproxy.py"]

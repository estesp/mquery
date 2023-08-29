.PHONY: function clean

PREFIX ?= ${DESTDIR}/usr
INSTALLDIR=${PREFIX}/bin

# Lambda "provider.al2" Go support requires the executable name
# to be "bootstrap"
function:
	GOARCH=amd64 GOOS=linux go build -tags lambda.norpc -o bootstrap inspect.go

clean:
	rm -f bootstrap

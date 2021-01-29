.PHONY: function clean

PREFIX ?= ${DESTDIR}/usr
INSTALLDIR=${PREFIX}/bin

function:
	GOOS=linux go build -o function inspect.go

clean:
	rm -f function

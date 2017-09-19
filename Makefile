.PHONY: binary clean install shell test-integration

PREFIX ?= ${DESTDIR}/usr
INSTALLDIR=${PREFIX}/bin

binary:
	go build mquery.go

static:
	go build -ldflags "-linkmode external -extldflags -static" -a -installsuffix cgo -o mquery .

clean:
	rm -f mquery

cross:
	packaging/cross.sh

cross-clean:
	rm -f mquery-*

install:
	install -d -m 0755 ${INSTALLDIR}
	install -m 755 mquery ${INSTALLDIR}


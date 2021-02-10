.PHONY: binary clean install cross-clean cross static

PREFIX ?= ${DESTDIR}/usr
INSTALLDIR=${PREFIX}/bin
ARCHFLAGS=
EXE=

ifeq ($(TARGETARCH),arm)
	ARCHFLAGS=GOARM=$(TARGETVARIANT)
endif
ifeq ($(TARGETOS),windows)
	EXE=.exe
endif

binary:
	go build -o mquery$(EXE) .

static:
	CGO_ENABLED=0 GOOS=linux GOARCH=$(TARGETARCH) $(ARCHFLAGS) go build -ldflags "-extldflags -static" -a -tags netgo -installsuffix netgo -o mquery 
clean:
	rm -f mquery

cross:
	packaging/cross.sh

cross-clean:
	rm -f mquery-*

install:
	install -d -m 0755 ${INSTALLDIR}
	install -m 755 mquery ${INSTALLDIR}


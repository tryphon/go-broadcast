TMPGOPATH := build
GOPKG := projects.tryphon.eu/go-broadcast

build:
	mkdir -p "${TMPGOPATH}/src/${GOPKG}"
	find . -path ./debian -prune -o -type f -name "*.go" -exec tar cf - {} + | (cd "${TMPGOPATH}/src/${GOPKG}" && tar xvf -)
	export GOPATH=${PWD}/${TMPGOPATH} && go get ${GOPKG}

install:
	cp ${TMPGOPATH}/bin/go-broadcast ${DESTDIR}/usr/bin

clean:
	rm -rf build
	rm Makefile

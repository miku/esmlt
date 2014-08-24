SHELL := /bin/bash
TARGETS = esmlt

# http://docs.travis-ci.com/user/languages/go/#Default-Test-Script
test:
	go get -d && go test -v

imports:
	goimports -w .

fmt:
	go fmt ./...

all: fmt test
	go build

install:
	go install

clean:
	go clean
	rm -f coverage.out
	rm -f $(TARGETS)
	rm -f esmlt-*.x86_64.rpm
	rm -f debian/esmlt*.deb
	rm -rf debian/esmlt/usr

cover:
	go get -d && go test -v	-coverprofile=coverage.out
	go tool cover -html=coverage.out

esmlt:
	go build cmd/esmlt/esmlt.go

# ==== packaging

deb: $(TARGETS)
	mkdir -p debian/esmlt/usr/sbin
	cp $(TARGETS) debian/esmlt/usr/sbin
	cd debian && fakeroot dpkg-deb --build esmlt .

REPOPATH = /usr/share/nginx/html/repo/CentOS/6/x86_64

publish: rpm
	cp esmlt-*.rpm $(REPOPATH)
	createrepo $(REPOPATH)

rpm: $(TARGETS)
	mkdir -p $(HOME)/rpmbuild/{BUILD,SOURCES,SPECS,RPMS}
	cp ./packaging/esmlt.spec $(HOME)/rpmbuild/SPECS
	cp $(TARGETS) $(HOME)/rpmbuild/BUILD
	./packaging/buildrpm.sh esmlt
	cp $(HOME)/rpmbuild/RPMS/x86_64/esmlt*.rpm .

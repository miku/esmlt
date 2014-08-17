SHELL := /bin/bash
TARGETS = duppool dupfilter

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
	rm -f dupsquash-*.x86_64.rpm
	rm -f debian/dupsquash*.deb
	rm -rf debian/dupsquash/usr

cover:
	go get -d && go test -v	-coverprofile=coverage.out
	go tool cover -html=coverage.out

duppool:
	go build cmd/duppool/duppool.go

dupfilter:
	go build cmd/dupfilter/dupfilter.go

# ==== packaging

deb: $(TARGETS)
	mkdir -p debian/dupsquash/usr/sbin
	cp $(TARGETS) debian/dupsquash/usr/sbin
	cd debian && fakeroot dpkg-deb --build dupsquash .

REPOPATH = /usr/share/nginx/html/repo/CentOS/6/x86_64

publish: rpm
	cp dupsquash-*.rpm $(REPOPATH)
	createrepo $(REPOPATH)

rpm: $(TARGETS)
	mkdir -p $(HOME)/rpmbuild/{BUILD,SOURCES,SPECS,RPMS}
	cp ./packaging/dupsquash.spec $(HOME)/rpmbuild/SPECS
	cp $(TARGETS) $(HOME)/rpmbuild/BUILD
	./packaging/buildrpm.sh dupsquash
	cp $(HOME)/rpmbuild/RPMS/x86_64/dupsquash*.rpm .

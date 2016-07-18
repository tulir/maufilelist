build:
	go build -o maufilelist

package-prep: build
	mkdir -p package/usr/bin
	mkdir -p package/etc/mfl
	mkdir -p package/var/log/mfl
	cp maufilelist package/usr/bin/
	cp example/config.json package/etc/mfl/
	cp example/format.gohtml package/etc/mfl/example-format.gohtml
	cp example/mfl.json package/etc/mfl/example-dirconf.json

package: package-prep
	dpkg-deb --build package maufilelist.deb > /dev/null

clean:
	rm -f maufilelist maufilelist.deb package/usr/bin/maufilelist package/etc/mfl/*

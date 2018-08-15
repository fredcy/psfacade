DESTDIR = /var/tmp/pscal

install:
	install -D pscal_service $(DESTDIR)/usr/local/bin/pscal_service
	install -D --mode=640 pscal.conf $(DESTDIR)/etc/init/pscal.conf

VERSION = 1
ITERATION = 7

fpm:
	fpm -s dir -t rpm -n pscal -v $(VERSION) --iteration $(ITERATION) -C $(DESTDIR) --config-files /etc .

DESTDIR = /var/tmp/pscal

install:
	install -D pscal_service $(DESTDIR)/usr/local/bin/pscal_service
	install -D --mode=640 pscal.conf $(DESTDIR)/etc/init/pscal.conf

fpm:
	fpm -s dir -t rpm -n pscal -v 1 -C $(DESTDIR) --config-files /etc .

FROM aporeto/promtail:master

ADD docktail /usr/bin/docktail
ADD entrypoint.sh /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]

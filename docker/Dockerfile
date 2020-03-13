FROM aporeto/promtail:v1.3.0

ADD docktail /usr/bin/docktail
ADD entrypoint.sh /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]

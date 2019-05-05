#!/bin/sh
# shellcheck disable=SC2068

# Start our extractor
/usr/bin/docktail &
# Start promtail
/usr/bin/promtail $@

#!/usr/bin/env bash
export GOFLAGS="-tags=sqlite_fts5 ${GOFLAGS}"
export CGO_ENABLED=${CGO_ENABLED:-1}
exec /usr/bin/env go "$@"

#!/usr/bin/env bash
set -Eeuo pipefail

BINARY='mquery'

LDFLAGS="-w -extldflags -static"
LDFLAGS_OTHER=

# List of platforms we build binaries for at this time:
PLATFORMS=(
  # format: GOOS/GOARCH[/GOARM]

  # OSX, Windows, Linux x86_64/i386
  darwin/amd64 windows/amd64 linux/amd64 linux/386

  # IBM POWER and z Systems
  linux/ppc64le linux/s390x

  # ARM; 32bit and 64bit
  linux/arm/5 linux/arm/6 linux/arm/7 linux/arm64

  # MIPS
  linux/mips64le

  # Risc-V
  linux/riscv64
)

FAILURES=()

for PLATFORM in "${PLATFORMS[@]}"; do
  GOOS="${PLATFORM%%/*}"
  GOARM="${PLATFORM#$GOOS/}"
  GOARCH="${GOARM%%/*}"
  GOARM="${GOARM#$GOARCH/}"

  BIN_FILENAME="${BINARY}-${GOOS}-${GOARCH}"
  ARCH_ENV="GOOS=${GOOS} GOARCH=${GOARCH}"
  if [ "${GOARCH}" = 'arm' ]; then
    [ -n "${GOARM}" ] || echo >&2 "WARNING: missing GOARM value for $PLATFORM in ${BASH_SOURCE[0]}"
    # "manifest-tool-linux-armv7", etc
    BIN_FILENAME="${BIN_FILENAME}v${GOARM}"
    ARCH_ENV="${ARCH_ENV} GOARM=${GOARM}"
  fi
  if [ "${GOOS}" = 'windows' ]; then
    BIN_FILENAME="${BIN_FILENAME}.exe"
  fi

  [ "${GOOS}" = 'linux' ] && _LDFLAGS="${LDFLAGS}" || _LDFLAGS="${LDFLAGS_OTHER}"

  CMD="${ARCH_ENV} CGO_ENABLED=0 GO_EXTLINK_ENABLED=0 go build -ldflags \"${_LDFLAGS}\" -o ${BIN_FILENAME} -tags netgo -installsuffix netgo ."
  echo "${CMD}"
  eval "${CMD}" || FAILURES=( "${FAILURES[@]}" "${PLATFORM}" )
done

# eval errors
if [ "${#FAILURES[@]}" -gt 0 ]; then
  echo >&2
  echo >&2 "ERROR: ${BINARY} build failed on: ${FAILURES[*]}"
  echo >&2
  exit 1
fi

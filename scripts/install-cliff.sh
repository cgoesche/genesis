#!/bin/bash

# Logs messages to STDOUT using the same message severities as described in RFC 5424 in 6.2.1
# Table 2: Syslog Message Severities - https://datatracker.ietf.org/doc/html/rfc5424#section-6.2.1
function log(){
  local log_lvl="${1}"
  local msg="${2}"

  case ${log_lvl} in
    0)
      printf "%s\n" "[$(tput setaf 1)$(tput bold)emergency$(tput sgr0)] ${msg}"
      ;;
    1)
      printf "%s\n" "[$(tput setaf 1)$(tput bold)alert$(tput sgr0)] ${msg}"
      ;;
    2)
      printf "%s\n" "[$(tput setaf 1)$(tput bold)critical$(tput sgr0)] ${msg}"
      ;;
    3)
      printf "%s\n" "[$(tput setaf 1)$(tput bold)error$(tput sgr0)] ${msg}"
      ;;
    4)
      printf "%s\n" "[$(tput setaf 3)$(tput bold)warning$(tput sgr0)] ${msg}"
      ;;
    5)
      printf "%s\n" "[$(tput setaf 2)$(tput bold)notice$(tput sgr0)] ${msg}"
      ;;
    6)
      printf "%s\n" "[$(tput setaf 2)$(tput bold)info$(tput sgr0)] ${msg}"
      ;;
    7)
      printf "%s\n" "[$(tput setaf 5)$(tput bold)debug$(tput sgr0)] ${msg}"
      ;;
    *)
      printf "%s\n" "${@}"
      ;;
  esac
}

BIN_INSTALL_DIR="/usr/local/bin"

log 6 "Querying Github API for latest git-cliff release download URL..."
RELEASE_TAR_DOWNLOAD_URL="$(curl -s https://api.github.com/repos/orhun/git-cliff/releases/latest \
	| tr -d \" \
	| awk -F" " '/.*-x86_64-unknown-linux-gnu.tar.gz$/ && /browser_download_url/ {print $2}')"

RELEASE_TAR_FILENAME="$(echo "${RELEASE_TAR_DOWNLOAD_URL}" | awk -F'/' '{print $9}')"
EXTRACTED_TAR_DIR_NAME="$(echo "${RELEASE_TAR_FILENAME}" | awk -F'-x86_64' '{print $1}')"

if [[ -n "${RELEASE_TAR_DOWNLOAD_URL}" && -n "${RELEASE_TAR_FILENAME}" ]]; then
  CHECKSUM_DOWNLOAD_URL="${RELEASE_TAR_DOWNLOAD_URL}.sha512"
  CHECKSUM_FILENAME="${RELEASE_TAR_FILENAME}.sha512"
else
  log 3 "Could not get latest git-cliff release download URL"
  exit 1
fi

log 6 "Downloading git-cliff latest binary release for x86_64"
if ! wget "${RELEASE_TAR_DOWNLOAD_URL}"; then
  log 3 "Could not download git-cliff"
fi

log 6 "Downloading git-cliff SHA512 checksum for ${RELEASE_TAR_FILENAME}"
if ! wget "${CHECKSUM_DOWNLOAD_URL}"; then
  log 3 "Could not download git-cliff checksum"
fi

log 6 "Verifying checksum"
if ! sha512sum -c "${CHECKSUM_FILENAME}"; then
  log 3 "The downloaded tar archive is corrupted"
  exit 1
fi
rm -f "${CHECKSUM_FILENAME}"

log 6 "Extracting ${PWD}/${RELEASE_TAR_FILENAME}"
if ! tar -xvf "${RELEASE_TAR_FILENAME}"; then
  log 3 "Failed to extract ${PWD}/${RELEASE_TAR_FILENAME}"
  exit 1
fi

log 6 "Copying binary ${EXTRACTED_TAR_DIR_NAME}/git-cliff to ${BIN_INSTALL_DIR}"
if ! cp "${EXTRACTED_TAR_DIR_NAME}/git-cliff" "${BIN_INSTALL_DIR}"; then
  log 3 "Failed to copy git-cliff to ${BIN_INSTALL_DIR}"
fi

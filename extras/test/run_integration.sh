#!/usr/bin/env bash
# // Copyright (c) 2022 Cisco and/or its affiliates.
# //
# // Licensed under the Apache License, Version 2.0 (the "License");
# // you may not use this file except in compliance with the License.
# // You may obtain a copy of the License at:
# //
# //     http://www.apache.org/licenses/LICENSE-2.0
# //
# // Unless required by applicable law or agreed to in writing, software
# // distributed under the License is distributed on an "AS IS" BASIS,
# // WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# // See the License for the specific language governing permissions and
# // limitations under the License.

set -euo pipefail

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd -P )"

args=($*)

echo "Preparing extras tests.."

VPP_REPO=${VPP_REPO:-master}

export CGO_ENABLED=0
export DOCKER_BUILDKIT=1
export GOTESTSUM_FORMAT="${GOTESTSUM_FORMAT:-testname}"

imgtag="govpp-extras-integration"

go test -c -o test/extras.test \
    -tags 'osusergo netgo e2e' \
    -ldflags '-w -s -extldflags "-static"' \
    -trimpath \
    "${SCRIPT_DIR}/memif"

docker build --tag "${imgtag}" \
    -f "${SCRIPT_DIR}"/build/Dockerfile.extras \
    --build-arg VPP_REPO="${VPP_REPO}" \
    "${SCRIPT_DIR}"/build

vppver=$(docker run --rm -i "${imgtag}" dpkg-query -f '${Version}' -W vpp)

if [ -n "${GITHUB_STEP_SUMMARY:-}" ]; then
    echo "**VPP version**: \`${vppver}\`" >> $GITHUB_STEP_SUMMARY
    echo "" >> $GITHUB_STEP_SUMMARY
fi

echo "=========================================================================="
echo " GOVPP EXTRAS INTEGRATION TEST - $(date) "
echo "=========================================================================="
echo "-     VPP_REPO: $VPP_REPO"
echo "-  VPP version: $vppver"
echo "--------------------------------------------------------------------------"

if docker run -i --privileged \
    -e CGO_ENABLED=0 \
    -e DEBUG_GOVPP \
    -e GOTESTSUM_FORMAT \
    -e CLICOLOR_FORCE \
    -v "$(cd "${SCRIPT_DIR}/.." && pwd)":/src \
    -w /src \
    "${imgtag}" gotestsum --raw-command -- go tool test2json -t -p extras ./test/extras.test -test.v ${args[@]:-}
then
	echo >&2 "-------------------------------------------------------------"
	echo >&2 -e " \e[32mPASSED\e[0m (took: ${SECONDS}s)"
	echo >&2 "-------------------------------------------------------------"
	exit 0
else
	res=$?
	echo >&2 "-------------------------------------------------------------"
	echo >&2 -e " \e[31mFAILED!\e[0m (exit code: $res)"
	echo >&2 "-------------------------------------------------------------"
	exit $res
fi

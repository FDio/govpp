#!/usr/bin/env bash
# // Copyright (c) 2025 Cisco and/or its affiliates.
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
VPP_REPO=${VPP_REPO:-release}
IMGTAG="govpp-integration"

# Build Docker image
docker build -t "$IMGTAG" -f "$SCRIPT_DIR/build/Dockerfile.integration" --build-arg VPP_REPO="$VPP_REPO" "$SCRIPT_DIR/build"

# Get VPP version
VPP_VERSION=$(docker run --rm -i "$IMGTAG" dpkg-query -f '${Version}' -W vpp)

# Display test information
echo "=========================================================================="
echo " GOVPP MEMORY BENCHMARK TEST - $(date) "
echo "=========================================================================="
echo "-     VPP_REPO: $VPP_REPO"
echo "-  VPP version: $VPP_VERSION"
echo "--------------------------------------------------------------------------"

# Run benchmark tests
if docker run -i --privileged -v "$(cd "$SCRIPT_DIR/.." && pwd)":/src -w /src/test/memory "$IMGTAG" go test -bench=.; then
    echo -e "\e[32mPASSED\e[0m (took: ${SECONDS}s)"
    exit 0
else
    echo -e "\e[31mFAILED!\e[0m (exit code: $?)"
    exit $?
fi

#! /usr/bin/env bash

poetryExe="$(which poetry)"
rv="$?"
if [ $rv -ne 0 ]; then
    >&2 echo "Poetry is not installed. Please install it first." 
    exit 1
fi
if [ "$poetryExe" = "" ]; then
    >&2 echo "No poetry executable found in PATH. Please install it first."
    exit 1
fi

CURDIR="$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"

REPOSITORY_ROOT="$(realpath ${CURDIR}/../..)"

PACKAGE_ROOT="${REPOSITORY_ROOT}/test"

venv_path="${REPOSITORY_ROOT}/.venv"

expectedRobotLibArtifact="$(realpath ${PACKAGE_ROOT}/dist/stackql_test_tooling-0.1.0-py3-none-any.whl)"

rm -f "${expectedRobotLibArtifact}" || true

cd "${PACKAGE_ROOT}"

poetry install

poetry build

if [ ! -f "${expectedRobotLibArtifact}" ]; then
    >&2 echo "Expected artifact not found: ${expectedRobotLibArtifact}"
    exit 1
fi


>&2 echo "Artifact built successfully: ${expectedRobotLibArtifact}"







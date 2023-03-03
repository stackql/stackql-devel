#!/usr/bin/env python3


import argparse
import os


def build_stackql(verbose :bool) -> int:
    os.environ['BuildMajorVersion'] = os.environ.get('BuildMajorVersion', '1')
    os.environ['BuildMinorVersion'] = os.environ.get('BuildMinorVersion', '1')
    os.environ['BuildPatchVersion'] = os.environ.get('BuildPatchVersion', '1')
    os.environ['CGO_ENABLED'] = os.environ.get('CGO_ENABLED', '1')
    return os.system(
        f'go build {"-x -v" if verbose else ""} --tags "json1 sqleanall" -ldflags "-X github.com/stackql/stackql/internal/stackql/cmd.BuildMajorVersion=$BuildMajorVersion '
        '-X github.com/stackql/stackql/internal/stackql/cmd.BuildMinorVersion=$BuildMinorVersion '
        '-X github.com/stackql/stackql/internal/stackql/cmd.BuildPatchVersion=$BuildPatchVersion '
        '-X github.com/stackql/stackql/internal/stackql/cmd.BuildCommitSHA=$BuildCommitSHA '
        '-X github.com/stackql/stackql/internal/stackql/cmd.BuildShortCommitSHA=$BuildShortCommitSHA '
        "-X 'github.com/stackql/stackql/internal/stackql/cmd.BuildDate=$BuildDate' "
        "-X 'stackql/internal/stackql/planbuilder.PlanCacheEnabled=$PlanCacheEnabled' "
        '-X github.com/stackql/stackql/internal/stackql/cmd.BuildPlatform=$BuildPlatform" '
        '-o build/ ./stackql'
    )


def unit_test_stackql(verbose :bool) -> int:
    return os.system(f'go test -timeout 1200s {"-v" if verbose else ""} --tags "json1 sqleanall"  ./...')


def run_robot_mocked_functional_tests_stackql(verbose :bool) -> int:
    return os.system(
        'robot '
        f'--variable SHOULD_RUN_DOCKER_EXTERNAL_TESTS:true '
        f'--variable CONCURRENCY_LIMIT:-1 ' 
        '-d test/robot/functional '
        'test/robot/functional'
    )


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument('--verbose', action='store_true')
    parser.add_argument('--build', action='store_true')
    parser.add_argument('--test', action='store_true')
    args = parser.parse_args()
    ret_code = 0
    if args.build:
        ret_code = build_stackql(args.verbose)
        if ret_code != 0:
            exit(ret_code)
    if args.test:
        ret_code = unit_test_stackql(args.verbose)
        if ret_code != 0:
            exit(ret_code)
    exit(ret_code)


if __name__ == '__main__':
    main()

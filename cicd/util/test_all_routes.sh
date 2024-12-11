#!/bin/bash

BASE_URL="${1:-'http://localhost:1080'}"

echo 'Testing /v1/projects/testing-project-three/locations/global/keyRings/testing-three/cryptoKeys'
curl -X GET $BASE_URL/v1/projects/testing-project-three/locations/global/keyRings/testing-three/cryptoKeys -w '\n\n'

echo 'Testing /v1/projects/testing-project-three/locations/australia-southeast2/keyRings/big-m-testing-three/cryptoKeys'
curl -X GET $BASE_URL/v1/projects/testing-project-three/locations/australia-southeast2/keyRings/big-m-testing-three/cryptoKeys -w '\n\n'

echo 'Testing /v1/projects/testing-project-three/locations/australia-southeast1/keyRings'
curl -X GET $BASE_URL/v1/projects/testing-project-three/locations/australia-southeast1/keyRings -w '\n\n'

echo 'Testing /v1/projects/testing-project-three/locations/global/keyRings'
curl -X GET $BASE_URL/v1/projects/testing-project-three/locations/global/keyRings -w '\n\n'

echo 'Testing /v1/projects/testing-project-three/locations/australia-southeast2/keyRings'
curl -X GET $BASE_URL/v1/projects/testing-project-three/locations/australia-southeast2/keyRings -w '\n\n'

echo 'Testing /v1/projects/testing-project-two/locations/global/keyRings/testing-two/cryptoKeys'
curl -X GET $BASE_URL/v1/projects/testing-project-two/locations/global/keyRings/testing-two/cryptoKeys -w '\n\n'

echo 'Testing /v1/projects/testing-project-two/locations/australia-southeast2/keyRings/big-m-testing-two/cryptoKeys'
curl -X GET $BASE_URL/v1/projects/testing-project-two/locations/australia-southeast2/keyRings/big-m-testing-two/cryptoKeys -w '\n\n'

echo 'Testing /v1/projects/testing-project-two/locations/australia-southeast1/keyRings'
curl -X GET $BASE_URL/v1/projects/testing-project-two/locations/australia-southeast1/keyRings -w '\n\n'

echo 'Testing /v1/projects/testing-project-two/locations/global/keyRings'
curl -X GET $BASE_URL/v1/projects/testing-project-two/locations/global/keyRings -w '\n\n'

echo 'Testing /v1/projects/testing-project-two/locations/australia-southeast2/keyRings'
curl -X GET $BASE_URL/v1/projects/testing-project-two/locations/australia-southeast2/keyRings -w '\n\n'

echo 'Testing /v1/projects/testing-project/locations/global/keyRings/testing/cryptoKeys'
curl -X GET $BASE_URL/v1/projects/testing-project/locations/global/keyRings/testing/cryptoKeys -w '\n\n'

echo 'Testing /v1/projects/testing-project/locations/australia-southeast1/keyRings'
curl -X GET $BASE_URL/v1/projects/testing-project/locations/australia-southeast1/keyRings -w '\n\n'

echo 'Testing /v1/projects/testing-project/locations/global/keyRings'
curl -X GET $BASE_URL/v1/projects/testing-project/locations/global/keyRings -w '\n\n'

echo 'Testing /projects/testing-project/zones/australia-southeast1-a/instances/000000001/getIamPolicy'
curl -X GET $BASE_URL/projects/testing-project/zones/australia-southeast1-a/instances/000000001/getIamPolicy -w '\n\n'

echo 'Testing /token'
curl -X GET $BASE_URL/token -w '\n\n'

echo 'Testing /v1/projects/testing-project/aggregated/usableSubnetworks'
curl -X GET $BASE_URL/v1/projects/testing-project/aggregated/usableSubnetworks -w '\n\n'

echo 'Testing /v1/projects/another-project/aggregated/usableSubnetworks'
curl -X GET $BASE_URL/v1/projects/another-project/aggregated/usableSubnetworks -w '\n\n'

echo 'Testing /v1/projects/yet-another-project/aggregated/usableSubnetworks'
curl -X GET $BASE_URL/v1/projects/yet-another-project/aggregated/usableSubnetworks -w '\n\n'

echo 'Testing /v1/projects/empty-project/aggregated/usableSubnetworks'
curl -X GET $BASE_URL/v1/projects/empty-project/aggregated/usableSubnetworks -w '\n\n'

echo 'Testing /projects/testing-project/zones/australia-southeast1-a/acceleratorTypes'
curl -X GET $BASE_URL/projects/testing-project/zones/australia-southeast1-a/acceleratorTypes -w '\n\n'

echo 'Testing /projects/another-project/zones/australia-southeast1-a/acceleratorTypes'
curl -X GET $BASE_URL/projects/another-project/zones/australia-southeast1-a/acceleratorTypes -w '\n\n'

echo 'Testing /v3/projects/testproject:getIamPolicy'
curl -X GET $BASE_URL/v3/projects/testproject:getIamPolicy -w '\n\n'


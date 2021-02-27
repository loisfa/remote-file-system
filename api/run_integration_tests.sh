#!/bin/bash
# To be run at the root of /api module
# Start the API prior to running these tests

echo "Running integration tests"
pipenv run python3 integration_tests

#!/bin/bash

echo "Running mypy type check"
cd integration_tests
# could not find a better way without throwing: error: Cannot find implementation or library stub for module named 'model.dto' 
pipenv run mypy __main__.py
cd ..
#! /bin/bash

PORT=8080
HEALTHCHECK_ENDPOINT="localhost:${PORT}/health-check"
RESPONSE_CODE=$(curl --silent --output /dev/null --write-out "%{http_code}" ${HEALTHCHECK_ENDPOINT})

if [ $RESPONSE_CODE != 200 ]
then
    echo "Starting the API..."
    go run main.go &
    GO_API_PID=$!
    echo "PID: ${GO_API_PID}"

    HEALTHCHECK_OK=false
    MAX_ATTEMPTS_COUNT=5
    ATTEMPTS_COUNT=0
    while [ $HEALTHCHECK_OK = false] && [$ATTEMPTS_COUNT -lt $MAX_ATTEMPTS_COUNT ];
    do
        sleep 1
        ATTEMPTS_COUNT=ATTEMPTS_COUNT+1
        RESPONSE_CODE=$(curl --silent --output /dev/null --write-out "%{http_code}" ${HEALTHCHECK_ENDPOINT})
        if [ $RESPONSE_CODE != 200 ]
        then
            HEALTHCHECK_OK
        fi
    done
    
    if [ $HEALTHCHECK_OK = false ]
    then 
        echo "API started"
    else
        echo "Could not start the API after ${MAX_ATTEMPTS_COUNT} attempts"
    fi
    
else 
    echo "API already started"
fi

echo "Running python integration tests"
python3 integration_tests

echo "PID: $GO_API_PID"
if [ -n "$GO_API_PID" ]
then
    echo "Stopping the API"
    # kill $GO_API_PID
    echo "Should kill PID: ${GO_API_PID}"
fi
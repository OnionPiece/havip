## About setup.sh

1. it supposes you have a local gitlab container supplies all projects needed as cache
2. it supposes you have a local file server & yum repo container supplies all files and .rpm packages needed as cache

## To run tests
    docker run -a stdout --net havipv2-test-net havipv2-testr

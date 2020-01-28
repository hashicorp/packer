docker pull openapitools/openapi-generator-cli

docker run --rm \
    -v ${PWD}:/local openapitools/openapi-generator-cli generate \
    -i https://api.hyperone.com/openapi.json \
    --git-user-id hyperonecom \
    --git-repo-id h1-client-go \
    -g go \
    -o /local

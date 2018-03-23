#!/bin/bash
NAME="registry.digitorus.com/pdfsigner"
TAGS=()

if [ -d .git ]; then
  GIT_COMMIT="`git rev-parse HEAD`"
  GIT_BRANCH="`git rev-parse --abbrev-ref HEAD`"
else
  GIT_COMMIT="${BITBUCKET_COMMIT}"
  GIT_BRANCH="${BITBUCKET_BRANCH}"
fi;

COMMIT="${NAME}:${GIT_COMMIT}"
BRANCH="${NAME}:${GIT_BRANCH}"

echo " ---> Installing dependencies..."
git config --global url.ssh://git@bitbucket.org/.insteadOf https://bitbucket.org/
go get

echo
echo " ---> Compiling server..."
echo

CGO_ENABLED=0 GOOS=linux go build -tags="docker" -a -installsuffix cgo \
  -ldflags "-w -s -X main.BuildDate=`date -u +%Y%m%d%H%M%S` -X main.Version=${GIT_COMMIT:0:7} -X main.GitCommit=${GIT_COMMIT} -X main.GitBranch=${GIT_BRANCH}"

if [ $? -ne 0 ]
then
  exit 1
fi

# add user (don't run as root)
echo "root:x:0:0:root:/root:/dev/null" > passwd
echo "user:x:10001:10001:user:/dev/null:/dev/null" >> passwd

curl -o ca-certificates.crt https://curl.haxx.se/ca/cacert.pem
tar cfz zoneinfo.tar.gz /usr/share/zoneinfo

# clean docker environment
docker system prune -a -f

# build new docker image
docker build --no-cache -t ${COMMIT} .

# tag image with the current branch name
TAGS+=(${GIT_BRANCH})

# the master branch is the latest prodcution version
if [ "${GIT_BRANCH}" == "master" ]; then
    TAGS+=("latest")
fi

echo

# add tags
for t in "${TAGS[@]}"; do
    docker tag ${COMMIT} ${NAME}:$t
    echo Added tag $t
done

if [ "${DOCKER_USERNAME}" != "" ]; then
  docker login --username $DOCKER_USERNAME --password $DOCKER_PASSWORD $DOCKER_REGISTRY
  docker push ${COMMIT}

  # push tags
  for t in "${TAGS[@]}"; do
   docker push ${NAME}:$t
  done

else
  echo
  echo Done, now run the following command to test your container locally:
  echo docker run ${NAME}
  echo
  echo Run the following command to upload:
  echo docker push ${NAME}
  echo
fi


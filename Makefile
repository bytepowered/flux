# Binary name
BINARY=flux
VERSION=0.20
GITCOMMIT=`git rev-parse --short HEAD`
BUILD_DATE=`date +%FT%T%z`

# Setup the -ldflags option for go build here, interpolate the variable values
LDFLAGS=-ldflags "-w -s -X main.GitCommit=${GITCOMMIT} -X main.Version=${VERSION} -X main.BuildDate=${BUILD_DATE}"
BUFLAGS=CGO_ENABLED=0

# Release
BUILD_DIR=./build
# Binary
OUTPUT="${BUILD_DIR}/${BINARY}"

# Builds the project
build:
		rm -rf ${BUILD_DIR}
		mkdir -p ${BUILD_DIR}

		# Build for linux
		${BUFLAGS} go build ${LDFLAGS} -a -installsuffix cgo -o ${OUTPUT} ./main/main.go

		# Copy configs and scripts
		cp -R ./main/conf.d ${BUILD_DIR}

		# Write version
		echo "${VERSION}" > ${BUILD_DIR}/version
		ls -lSh ${BUILD_DIR}

install:
		go install

clean:
		go clean

.PHONY:  clean build

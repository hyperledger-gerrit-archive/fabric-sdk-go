#!/bin/bash
# Original script from Hyperledger Fabric and Fabric CA projects

set -e

if [ ! -f $GOPATH/bin/golint ]; then
   echo "Installing golint ..."
   go get -u github.com/golang/lint/golint
fi

declare -a arr=(
"./config"
"./fabric-ca-client"
"./fabric-client"
"./test"
)

echo "Running linters..."
for i in "${arr[@]}"
do
   echo "Checking $i"
   OUTPUT="$(golint $i/...)"
   if [[ $OUTPUT ]]; then
      echo "You should check the following golint suggestions:"
	    printf "$OUTPUT\n"
      echo "end golint suggestions"
   fi

   OUTPUT="$(go vet $i/...)"
   if [[ $OUTPUT ]]; then
      echo "You should check the following govet suggestions:"
	    printf "$OUTPUT\n"
      echo "end govet suggestions"
   fi

   found=`gofmt -l \`find $i -name "*.go" |grep -v "./vendor"\` 2>&1`
   if [ $? -ne 0 ]; then
      echo "The following files need reformatting with 'gofmt -w <file>':"
      printf "$badformat\n"
      exit 1
   fi

   OUTPUT="$(goimports -srcdir $GOPATH/src/github.com/hyperledger/fabric-sdk-go -l $i)"
   if [[ $OUTPUT ]]; then
      echo "YOU MUST FIX THE FOLLOWING GOIMPORTS ERRORS:"
	    printf "$OUTPUT\n"
      echo "END GOIMPORTS ERRORS"
      exit 1
   fi
done

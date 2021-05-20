#!/bin/bash

# Script compiles hyperledger projects which is needed,
# starts the sample PAC-network, creates channels mychannel1, mychannel2
# and runs the client application with private atomic commit

FABRIC_PROJECT_PATH=/home/vano/go/src/github.com/hyperledger/fabric
FABRIC_SAMPLES_PATH=/home/vano/go/src/github.com/hyperledger/fabric-samples
TEST_NETWORK_PATH=/home/vano/go/src/github.com/hyperledger/fabric-samples/test-network
APPLICATION_PATH=/home/vano/go/src/github.com/hyperledger/fabric-samples/asset-transfer-basic/application-pac-go
#FABRIC_SDK_GO_PATH=...

cd $TEST_NETWORK_PATH

set -x
./network_pac.sh down
res=$?
{ set +x; } 2>/dev/null
if [ $res -ne 0 ]; then
  fatalln "Failed to stop running network"
fi

if [[ $1 == fabric ]]; then

  cd $FABRIC_PROJECT_PATH
  
  #compiling fabric
  set -x
  make clean docker-clean peer-docker orderer-docker tools-docker docker-thirdparty docker native
  res=$?
  { set +x; } 2>/dev/null
  if [ $res -ne 0 ]; then
    fatalln "Failed to compile fabric project"
  fi
  
  #copying binaries
  #remember that ca-binaries must be copied there manually as well
  cp  $FABRIC_PROJECT_PATH/build/bin/* $FABRIC_SAMPLES_PATH/bin

fi

cd $TEST_NETWORK_PATH

#run test-network and create channels
set -x
./network_pac.sh up createChannels
res=$?
{ set +x; } 2>/dev/null
if [ $res -ne 0 ]; then
  fatalln "Failed to run network or create channels"
fi

#install chaincodes
set -x
./network_pac.sh deployCC -ccn cc1 -ccp ../asset-transfer-basic/pac-chaincodes-go/cc1-assets -ccl go
res=$?
{ set +x; } 2>/dev/null
if [ $res -ne 0 ]; then
  fatalln "Failed to install cc1"
fi

set -x
./network_pac.sh deployCC -ccn cc2 -ccp ../asset-transfer-basic/pac-chaincodes-go/cc2-balances -ccl go
res=$?
{ set +x; } 2>/dev/null
if [ $res -ne 0 ]; then
  fatalln "Failed to install cc2"
fi


cd $APPLICATION_PATH

#running the client application
set -x
go run assetPACTransfer.go
res=$?
{ set +x; } 2>/dev/null
if [ $res -ne 0 ]; then
  fatalln "Failed to run PAC client application"
fi

#getting docker containers logs
#ORDERER
set -x
ORDERER_ID=`docker inspect --format="{{.Id}}" orderer.example.com`
res=$?
{ set +x; } 2>/dev/null
if [ $res -ne 0 ]; then
  fatalln "Failed to get docker container id"
fi

set -x
docker logs $ORDERER_ID >& /home/vano/logs/orderer.example.com.log
res=$?
{ set +x; } 2>/dev/null
if [ $res -ne 0 ]; then
  fatalln "Failed to save docker container logs"
fi

#PEER0_ORG3
set -x
PEER_ORG3_ID=`docker inspect --format="{{.Id}}" peer0.org3.example.com`
res=$?
{ set +x; } 2>/dev/null
if [ $res -ne 0 ]; then
  fatalln "Failed to get docker container id"
fi

set -x
docker logs $PEER_ORG3_ID >& /home/vano/logs/peer0.org3.example.com.log
res=$?
{ set +x; } 2>/dev/null
if [ $res -ne 0 ]; then
  fatalln "Failed to save docker container logs"
fi

echo "don't forget to stop network using in '${TEST_NETWORK_PATH}' command"
echo "./network_pac.sh down"
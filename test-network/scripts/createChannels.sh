#!/bin/bash

# imports  
. scripts/envVar.sh
. scripts/utils.sh

CHANNEL_NAME="$1"
DELAY="$2"
MAX_RETRY="$3"
VERBOSE="$4"
: ${CHANNEL_NAME:="mychannel"}
: ${DELAY:="3"}
: ${MAX_RETRY:="5"}
: ${VERBOSE:="false"}

if [ ! -d "channel-artifacts" ]; then
	mkdir channel-artifacts
fi

createChannelsGenesisBlock() {
	which configtxgen
	if [ "$?" -ne 0 ]; then
		fatalln "configtxgen tool not found."
	fi
	set -x
	#configtxgen -profile TwoOrgsApplicationGenesis -outputBlock ./channel-artifacts/${CHANNEL_NAME}.block -channelID ${CHANNEL_NAME}
	configtxgen -profile Org1ApplicationGenesis -outputBlock ./channel-artifacts/${CHANNEL_NAME}1.block -channelID ${CHANNEL_NAME}1
	configtxgen -profile Org2ApplicationGenesis -outputBlock ./channel-artifacts/${CHANNEL_NAME}2.block -channelID ${CHANNEL_NAME}2
	res=$?
	{ set +x; } 2>/dev/null
  verifyResult $res "Failed to generate channel configuration transaction..."
}

createChannels() {
	setGlobals 1
	# Poll in case the raft leader is not set yet
	local rc=1
	local COUNTER=1
	while [ $rc -ne 0 -a $COUNTER -lt $MAX_RETRY ] ; do
		sleep $DELAY
		set -x
		#osnadmin channel join --channelID $CHANNEL_NAME --config-block ./channel-artifacts/${CHANNEL_NAME}.block -o localhost:7053 --ca-file "$ORDERER_CA" --client-cert "$ORDERER_ADMIN_TLS_SIGN_CERT" --client-key "$ORDERER_ADMIN_TLS_PRIVATE_KEY" >&log.txt
		osnadmin channel join --channelID ${CHANNEL_NAME}1 --config-block ./channel-artifacts/${CHANNEL_NAME}1.block -o localhost:7053 --ca-file "$ORDERER_CA" --client-cert "$ORDERER_ADMIN_TLS_SIGN_CERT" --client-key "$ORDERER_ADMIN_TLS_PRIVATE_KEY" >&log.txt
		osnadmin channel join --channelID ${CHANNEL_NAME}2 --config-block ./channel-artifacts/${CHANNEL_NAME}2.block -o localhost:7053 --ca-file "$ORDERER_CA" --client-cert "$ORDERER_ADMIN_TLS_SIGN_CERT" --client-key "$ORDERER_ADMIN_TLS_PRIVATE_KEY" >&log.txt
		res=$?
		{ set +x; } 2>/dev/null
		let rc=$res
		COUNTER=$(expr $COUNTER + 1)
	done
	cat log.txt
	verifyResult $res "Channel creation failed"
}

# joinChannel ORG
joinChannel1() {
  FABRIC_CFG_PATH=$PWD/../config/
  ORG=$1
  setGlobals $ORG
	local rc=1
	local COUNTER=1
	## Sometimes Join takes time, hence retry
	while [ $rc -ne 0 -a $COUNTER -lt $MAX_RETRY ] ; do
    sleep $DELAY
    set -x
    peer channel join -b $BLOCKFILE1 >&log.txt
    res=$?
    { set +x; } 2>/dev/null
		let rc=$res
		COUNTER=$(expr $COUNTER + 1)
	done
	cat log.txt
	verifyResult $res "After $MAX_RETRY attempts, peer0.org${ORG} has failed to join channel '$CHANNEL_NAME' "
}

joinChannel2() {
  FABRIC_CFG_PATH=$PWD/../config/
  ORG=$1
  setGlobals $ORG
	local rc=1
	local COUNTER=1
	## Sometimes Join takes time, hence retry
	while [ $rc -ne 0 -a $COUNTER -lt $MAX_RETRY ] ; do
    sleep $DELAY
    set -x
    peer channel join -b $BLOCKFILE2 >&log.txt
    res=$?
    { set +x; } 2>/dev/null
		let rc=$res
		COUNTER=$(expr $COUNTER + 1)
	done
	cat log.txt
	verifyResult $res "After $MAX_RETRY attempts, peer0.org${ORG} has failed to join channel '$CHANNEL_NAME' "
}

setAnchorPeer1() {
  ORG=$1
  docker exec cli ./scripts/setAnchorPeer.sh $ORG ${CHANNEL_NAME}1 
}

setAnchorPeer2() {
  ORG=$1
  docker exec cli ./scripts/setAnchorPeer.sh $ORG ${CHANNEL_NAME}2 
}

FABRIC_CFG_PATH=${PWD}/configtx

## Create channel genesis block
infoln "Generating channel genesis blocks '${CHANNEL_NAME}1.block', '${CHANNEL_NAME}2.block'"
createChannelsGenesisBlock

FABRIC_CFG_PATH=$PWD/../config/
#BLOCKFILE="./channel-artifacts/${CHANNEL_NAME}.block"
BLOCKFILE1="./channel-artifacts/${CHANNEL_NAME}1.block"
BLOCKFILE2="./channel-artifacts/${CHANNEL_NAME}2.block"

## Create channel
infoln "Creating channels ${CHANNEL_NAME}1, ${CHANNEL_NAME}2"
createChannels
successln "Channels '${CHANNEL_NAME}1', '${CHANNEL_NAME}2' created"

## Join all the peers to the channel
infoln "Joining org1 peer to the channel1..."
joinChannel1 1
infoln "Joining org2 peer to the channel2..."
joinChannel2 2

## Set the anchor peers for each org in the channel
infoln "Setting anchor peer for org1..."
setAnchorPeer1 1
infoln "Setting anchor peer for org2..."
setAnchorPeer2 2

successln "Channel '$CHANNEL_NAME' joined"

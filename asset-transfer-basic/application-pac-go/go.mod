module asset-transfer-basic

go 1.14

require (
	github.com/golang/protobuf v1.3.3
	github.com/hyperledger/fabric-protos-go v0.0.0-20200707132912-fee30f3ccd23
	github.com/hyperledger/fabric-sdk-go v1.0.0-rc1
	golang.org/x/net v0.0.0-20201021035429-f5854403a974 // indirect
	golang.org/x/sys v0.0.0-20210119212857-b64e53b001e4 // indirect
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1 // indirect
)

replace github.com/hyperledger/fabric-sdk-go => github.com/pacproj/fabric-sdk-go v1.0.0-beta3.0.20210614131156-50524dcf4a35

replace github.com/hyperledger/fabric-protos-go => github.com/pacproj/fabric-protos-go v0.0.0-20210413120152-f3382d191db5

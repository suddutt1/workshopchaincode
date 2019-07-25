# Car trade chaincode 



### Install libraries 
```sh
cd $GOPATH/src
mkdir -p github.com/hyperledger
cd github.com/hyperledger
git clone https://github.com/hyperledger/fabric.git
cd fabric
git checkout tags/v1.4.0

```
### Install the chaincode 

```sh
cd $GOPATH/src
mkdir -p github.com/blockchain
cd github.com/blockchain
git clone https://github.com/suddutt1/workshopchaincode.git . # Do not miss the DOT(.) at the end
# Complie the code
go builld
```
source dev.env

cd $GOPATH/src/github.com/OutCast-IO/Outcaster
rm -f Outcaster

godep go clean -i

godep go build -x -gcflags "-N -l" -o Outcaster
#-o $GOPATH/bin/Outcaster
#go build -x -gcflags "-N -l" -o Outcaster

echo
./Outcaster
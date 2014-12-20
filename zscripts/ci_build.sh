
echo
echo "=> Build Starting..."
echo
echo bitbucket.org/ardanlabs/cobalt
go clean -i bitbucket.org/ardanlabs/cobalt || { exit 1; }
go build bitbucket.org/ardanlabs/cobalt || { exit 1; }
go vet bitbucket.org/ardanlabs/cobalt || { exit 1; }
echo
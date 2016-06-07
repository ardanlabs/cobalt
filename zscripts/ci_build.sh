
echo
echo "=> Build Starting..."
echo
echo github.com/ardanlabs/cobalt
go clean -i github.com/ardanlabs/cobalt || { exit 1; }
go build github.com/ardanlabs/cobalt || { exit 1; }
go vet github.com/ardanlabs/cobalt || { exit 1; }

echo "=> Build Finished..."
echo

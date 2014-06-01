
echo
echo "=> Build Starting..."
echo
echo "go get -u github.com/OutCast-IO/pat"

go get -u github.com/OutCast-IO/pat || {exit 1;}

echo
echo "=> Compiling"
echo

go build -v || {exit;}
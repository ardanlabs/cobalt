echo '=> Starting Tests...'

#go test -v -bench=. -benchtime=5s || { exit 1; }
#go test -v -bench=. -benchtime=20s
go test -v -bench=. || { exit 1; }

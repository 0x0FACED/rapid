unset CC
unset CXX
export GOOS=linux
export GOARCH=amd64
cd cmd/rapid
go build -o rapid main.go
fyne package -os linux -icon test.png
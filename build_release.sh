echo Building ./cubeintro binary...
go build -ldflags="-s -w" .
ls -al cubeintro
ls -alh cubeintro
echo Superstripping the binary...
sstrip -z cubeintro
ls -al cubeintro
ls -alh cubeintro
echo Maxiumum Best Ultra-brute UPX compressing the binary
nice -19 upx -9 --best --ultra-brute cubeintro
ls -al cubeintro
ls -alh cubeintro
./cubeintro

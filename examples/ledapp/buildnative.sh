# Some programs, great when testing software updates
# native for qemu?
CGO_ENABLED=0 go build -ldflags "-X main.txtOnTime=100ms -X main.txtOffTime=500ms" -o program
CGO_ENABLED=0 go build -ldflags "-X main.txtOnTime=1500ms -X main.txtOffTime=500ms" -o programB
CGO_ENABLED=0 go build -ldflags "-X main.txtOnTime=1s -X main.txtOffTime=1s" -o programC


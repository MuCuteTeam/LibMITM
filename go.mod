module androtun

go 1.19

require (
	golang.org/x/sys v0.2.0
	golang.org/x/time v0.0.0-20191024005414-555d28b269f0
	gvisor.dev/gvisor v0.0.0-20230115200031-42e92cae6e7b
)

require (
	github.com/google/btree v1.0.1 // indirect
	golang.org/x/mobile v0.0.0-20221110043201-43a038452099 // indirect
	golang.org/x/mod v0.6.0-dev.0.20220419223038-86c51ed26bb4 // indirect
	golang.org/x/tools v0.1.12 // indirect
)

replace gvisor.dev/gvisor => github.com/sagernet/gvisor v0.0.0-20220402114650-763d12dc953e

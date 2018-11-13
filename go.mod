module github.com/raedahgroup/dcrcli

require (
	github.com/btcsuite/go-flags v0.0.0-20150116065318-6c288d648c1c
	github.com/decred/dcrd/dcrutil v1.1.1
	github.com/go-chi/chi v3.3.3+incompatible
	github.com/golang/protobuf v1.2.0 // indirect
	github.com/raedahgroup/dcrcli/walletrpcclient v0.0.1
	github.com/skip2/go-qrcode v0.0.0-20171229120447-cf5f9fa2f0d8
	golang.org/x/crypto v0.0.0-20180718160520-a2144134853f
	golang.org/x/sys v0.0.0-20180928133829-e4b3c5e90611 // indirect
	google.golang.org/genproto v0.0.0-20180928223349-c7e5094acea1 // indirect
)

replace github.com/raedahgroup/dcrcli/walletrpcclient => ./walletrpcclient

module payment-service

go 1.22

require (
	build/proto v0.0.0
	github.com/google/uuid v1.6.0
	google.golang.org/protobuf v1.33.0
	kafkaclient v0.0.0-00010101000000-000000000000
)

require github.com/google/go-cmp v0.6.0 // indirect

replace (
	build/proto => ./build/proto
	kafkaclient => ./kafkaclient
)

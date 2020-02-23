assets::
	rm -f */*~
	go run github.com/shuLhan/go-bindata/cmd/go-bindata -o assets.go assets/*
	
run: assets
	go run . . 

install:
	go install ./...
	

go build -o c cmd/client/main.go
go build -o c cmd/server/main.go


go run cmd/server/main.go
go run cmd/client/main.go
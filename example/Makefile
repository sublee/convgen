all: pb/job.pb.go db/models.go api/api.go

pb/job.pb.go: job.proto
	buf generate

db/models.go: schema/000000_job.up.sql
	sqlc generate

api/api.go: api.yaml
	oapi-codegen -o api/api.go ./api.yaml

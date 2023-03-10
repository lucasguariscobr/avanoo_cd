module github.com/avanoo/avanoo_cd

go 1.15

require (
	github.com/avanoo/avanoo_cd/.docs v0.0.0-00010101000000-000000000000 // indirect
	github.com/avanoo/avanoo_cd/deploy v0.0.0-00010101000000-000000000000
	github.com/avanoo/avanoo_cd/environments v0.0.0-00010101000000-000000000000 // indirect
	github.com/avanoo/avanoo_cd/server v0.0.0-00010101000000-000000000000
	github.com/avanoo/avanoo_cd/utils v0.0.0-00010101000000-000000000000
	github.com/go-redis/redis/v8 v8.7.1 // indirect
	github.com/google/go-cmp v0.5.5 // indirect
	github.com/gorilla/mux v1.8.0 // indirect
	github.com/justinas/alice v1.2.0 // indirect
	github.com/satori/go.uuid v1.2.0 // indirect
	github.com/sendgrid/rest v2.6.3+incompatible // indirect
	github.com/sendgrid/sendgrid-go v3.8.0+incompatible // indirect
	github.com/swaggo/http-swagger v1.0.0 // indirect
)

replace github.com/avanoo/avanoo_cd/deploy => ./deploy

replace github.com/avanoo/avanoo_cd/server => ./server

replace github.com/avanoo/avanoo_cd/utils => ./utils

replace github.com/avanoo/avanoo_cd/environments => ./environments

replace github.com/avanoo/avanoo_cd/.docs => ./.docs

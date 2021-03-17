module github.com/avanoo/avanoo_cd

go 1.15

require (
	github.com/alecthomas/template v0.0.0-20190718012654-fb15b899a751
	github.com/go-redis/redis v6.15.9+incompatible
	github.com/google/go-cmp v0.5.5
	github.com/gorilla/mux v1.8.0
	github.com/justinas/alice v1.2.0
	github.com/satori/go.uuid v1.2.0
	github.com/sendgrid/rest v2.6.3+incompatible
	github.com/sendgrid/sendgrid-go v3.8.0+incompatible
	github.com/stretchr/testify v1.6.1
	github.com/swaggo/http-swagger v1.0.0
	github.com/swaggo/swag v1.7.0
    github.com/avanoo/avanoo_cd/deploy v0.0.0-00010101000000-000000000000
    github.com/avanoo/avanoo_cd/server v0.0.0-00010101000000-000000000000
    github.com/avanoo/avanoo_cd/utils v0.0.0-00010101000000-000000000000
    github.com/avanoo/avanoo_cd/environments v0.0.0-00010101000000-000000000000
    github.com/avanoo/avanoo_cd/.docs v0.0.0-00010101000000-000000000000
)

replace github.com/avanoo/avanoo_cd/deploy          => ./deploy
replace github.com/avanoo/avanoo_cd/server          => ./server
replace github.com/avanoo/avanoo_cd/utils           => ./utils
replace github.com/avanoo/avanoo_cd/environments    => ./environments
replace github.com/avanoo/avanoo_cd/.docs           => ./.docs
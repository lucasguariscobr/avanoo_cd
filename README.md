`go run avanoo_cd`
`./avanoo_cd cd_config.json`
`go test ./... -coverpkg=$(go list ./... | grep -v .doc | tr '\n' ',') -coverprofile=cp.out -args -config $(pwd)/conf/configuration_test.json`
`go tool cover -html=cp.out`
`go tool cover -func=cp.out | grep total`

# Push application
`../../bin/swag init --output ./.docs`
`go install avanoo_cd`
`mv ../../bin/avanoo_cd .`
`sftp lucas@cd.placeboapp.com`
`put avanoo_cd`

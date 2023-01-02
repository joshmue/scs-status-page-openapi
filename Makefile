generate:
	oapi-codegen --package api --generate types,server,spec openapi.yaml > pkg/api/api.gen.go
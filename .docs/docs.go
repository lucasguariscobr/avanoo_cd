// GENERATED BY THE COMMAND ABOVE; DO NOT EDIT
// This file was generated by swaggo/swag at
// 2019-12-02 15:29:59.067166 -0300 -03 m=+0.052952651

package docs

import (
	"bytes"
	"encoding/json"
	"strings"

	"github.com/alecthomas/template"
	"github.com/swaggo/swag"
)

var doc = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{.Description}}",
        "title": "{{.Title}}",
        "contact": {
            "name": "Lucas",
            "email": "lucas@avanoo.com"
        },
        "license": {
            "name": "GNU GPLv3"
        },
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/builds": {
            "get": {
                "description": "Returns the list of builds executed by the service.\nOnly branches that are connected to a domain have their docker images generated.\nThe build information is kept for one week.",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "builds"
                ],
                "summary": "List all builds",
                "parameters": [
                    {
                        "enum": [
                            "queued",
                            "running",
                            "completed",
                            "successful",
                            "failed"
                        ],
                        "type": "string",
                        "description": "Filter Builds",
                        "name": "filter",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/deploy.Build"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/utils.JSONErrror"
                        }
                    },
                    "404": {},
                    "405": {}
                }
            }
        },
        "/domain": {
            "post": {
                "description": "Creates or updates avanoo domains.\nWhen the domain already exists, its content will be replaced by the params provided on the request.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "domains"
                ],
                "summary": "Manage Avanoo Domains",
                "parameters": [
                    {
                        "description": "Domain",
                        "name": "domain",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "type": "object",
                            "$ref": "#/definitions/deploy.Domain"
                        }
                    }
                ],
                "responses": {
                    "204": {},
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/utils.JSONErrror"
                        }
                    },
                    "404": {},
                    "405": {}
                }
            }
        },
        "/domain/{domain}": {
            "get": {
                "description": "Return attributes of an existing domain",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "domains"
                ],
                "summary": "Return attributes of an existing domain",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Name of the domain",
                        "name": "domain",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/deploy.Domain"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/utils.JSONErrror"
                        }
                    },
                    "404": {},
                    "405": {}
                }
            },
            "delete": {
                "description": "Delete an existing domain's definition.\nThis operation does not stop the service or modifies the exisitng domain.",
                "tags": [
                    "domains"
                ],
                "summary": "Delete an existing domain's definition",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Name of the domain",
                        "name": "domain",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "204": {},
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/utils.JSONErrror"
                        }
                    },
                    "404": {},
                    "405": {}
                }
            }
        },
        "/domains": {
            "get": {
                "description": "Return all domains and their parameters.",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "domains"
                ],
                "summary": "List all domains",
                "responses": {
                    "200": {
                        "description": "List of Domains",
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/deploy.Domain"
                            }
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/utils.JSONErrror"
                        }
                    },
                    "404": {},
                    "405": {}
                }
            }
        },
        "/health": {
            "get": {
                "description": "health check",
                "tags": [
                    "health"
                ],
                "summary": "Health check",
                "responses": {
                    "200": {},
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/utils.JSONErrror"
                        }
                    },
                    "404": {},
                    "405": {}
                }
            }
        },
        "/updateDomainBranch": {
            "post": {
                "description": "Update the branch used by an existing domain.\nIt will update the domain after receiving the new branch value.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "domains"
                ],
                "summary": "Update a Domain's branch",
                "parameters": [
                    {
                        "description": "Update Domain Info",
                        "name": "domain",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "type": "object",
                            "$ref": "#/definitions/deploy.UpdateDomain"
                        }
                    }
                ],
                "responses": {
                    "204": {},
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/utils.JSONErrror"
                        }
                    },
                    "404": {},
                    "405": {}
                }
            }
        },
        "/webhook": {
            "post": {
                "description": "Webhook that communicates with Github.\nCurrently listens to push events from the App repository.",
                "tags": [
                    "webhook"
                ],
                "summary": "GitHub integration",
                "responses": {
                    "204": {},
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/utils.JSONErrror"
                        }
                    },
                    "404": {},
                    "405": {}
                }
            }
        }
    },
    "definitions": {
        "deploy.Build": {
            "type": "object",
            "properties": {
                "branch": {
                    "type": "string"
                },
                "date": {
                    "type": "string"
                },
                "domains": {
                    "type": "array",
                    "items": {
                        "type": "string"
                    }
                },
                "err": {
                    "type": "error"
                },
                "id": {
                    "type": "string"
                },
                "status": {
                    "type": "string"
                },
                "wg": {
                    "type": "string"
                }
            }
        },
        "deploy.Domain": {
            "type": "object",
            "properties": {
                "branch": {
                    "type": "string",
                    "example": "development"
                },
                "domain": {
                    "type": "string",
                    "example": "pre.avanoo.com"
                },
                "extra_vars": {
                    "type": "object"
                },
                "host": {
                    "type": "string"
                }
            }
        },
        "deploy.UpdateDomain": {
            "type": "object",
            "properties": {
                "branch": {
                    "type": "string",
                    "example": "development"
                },
                "domain": {
                    "type": "string",
                    "example": "pre.avanoo.com"
                }
            }
        },
        "utils.JSONErrror": {
            "type": "object",
            "properties": {
                "msg": {
                    "type": "string"
                }
            }
        }
    }
}`

type swaggerInfo struct {
	Version     string
	Host        string
	BasePath    string
	Schemes     []string
	Title       string
	Description string
}

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = swaggerInfo{
	Version:     "1.0",
	Host:        "cd.placeboapp.com",
	BasePath:    "",
	Schemes:     []string{},
	Title:       "Avanoo Continuous Delivery",
	Description: "Automating cd process.",
}

type s struct{}

func (s *s) ReadDoc() string {
	sInfo := SwaggerInfo
	sInfo.Description = strings.Replace(sInfo.Description, "\n", "\\n", -1)

	t, err := template.New("swagger_info").Funcs(template.FuncMap{
		"marshal": func(v interface{}) string {
			a, _ := json.Marshal(v)
			return string(a)
		},
	}).Parse(doc)
	if err != nil {
		return doc
	}

	var tpl bytes.Buffer
	if err := t.Execute(&tpl, sInfo); err != nil {
		return doc
	}

	return tpl.String()
}

func init() {
	swag.Register(swag.Name, &s{})
}

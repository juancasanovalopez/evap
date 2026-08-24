// Command api is the Lambda entrypoint. Locally it can also run as a plain
// HTTP server when LOCAL_HTTP_ADDR is set (useful for manual testing).
package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/awslabs/aws-lambda-go-api-proxy/httpadapter"

	appauth "evap-backend/internal/auth"
	appconfig "evap-backend/internal/config"
	"evap-backend/internal/router"
	"evap-backend/internal/store"
)

func main() {
	ctx := context.Background()

	cfg, err := appconfig.Load(ctx)
	if err != nil {
		log.Fatalf("failed to load configuration: %v", err)
	}

	awsCfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Fatalf("failed to load AWS SDK configuration: %v", err)
	}

	repo := store.NewDynamoDBUserRepository(dynamodb.NewFromConfig(awsCfg), cfg.DynamoTableName)
	issuer := appauth.NewTokenIssuer(cfg.JWTSigningKey, router.DefaultJWTTTL)

	handler := router.New(router.Deps{
		Config:        cfg,
		Users:         repo,
		Issuer:        issuer,
		SecureCookies: true,
	})

	if addr := os.Getenv("LOCAL_HTTP_ADDR"); addr != "" {
		log.Printf("listening locally on %s", addr)
		log.Fatal(http.ListenAndServe(addr, handler))
		return
	}

	lambda.Start(httpadapter.NewV2(handler).ProxyWithContext)
}

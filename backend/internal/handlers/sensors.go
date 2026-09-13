package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iot"
	"github.com/aws/aws-sdk-go-v2/service/iot/types"

	"evap-backend/internal/i18n"
	"evap-backend/internal/middleware"
)

type IoTClient interface {
	CreateThing(context.Context, *iot.CreateThingInput, ...func(*iot.Options)) (*iot.CreateThingOutput, error)
	CreateKeysAndCertificate(context.Context, *iot.CreateKeysAndCertificateInput, ...func(*iot.Options)) (*iot.CreateKeysAndCertificateOutput, error)
	AttachThingPrincipal(context.Context, *iot.AttachThingPrincipalInput, ...func(*iot.Options)) (*iot.AttachThingPrincipalOutput, error)
	AttachPolicy(context.Context, *iot.AttachPolicyInput, ...func(*iot.Options)) (*iot.AttachPolicyOutput, error)
}

type createSensorRequest struct {
	Name string `json:"name"`
}

type createSensorResponse struct {
	ThingName      string `json:"thing_name"`
	OwnerUserID    string `json:"owner_user_id"`
	Endpoint       string `json:"endpoint"`
	Topic          string `json:"topic"`
	CertificatePEM string `json:"certificate_pem"`
	PrivateKey     string `json:"private_key"`
}

var sensorNamePattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_-]{0,63}$`)

func CreateSensorHandler(client IoTClient, policyName, endpoint string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims, ok := middleware.ClaimsFromContext(r.Context())
		if !ok {
			writeLocalizedError(w, r, http.StatusUnauthorized, i18n.AuthUnauthorized)
			return
		}
		var request createSensorRequest
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096)).Decode(&request); err != nil || !sensorNamePattern.MatchString(request.Name) {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "sensor name must use 1-64 letters, numbers, hyphens, or underscores"})
			return
		}

		ownerID := ownerSlug(claims.Subject)
		thingName := fmt.Sprintf("%s-%s", ownerID, request.Name)
		thing, err := client.CreateThing(r.Context(), &iot.CreateThingInput{
			ThingName:        aws.String(thingName),
			AttributePayload: &types.AttributePayload{Attributes: map[string]string{"owner_user_id": ownerID}},
		})
		if err != nil {
			writeJSON(w, http.StatusBadGateway, map[string]string{"error": "could not create sensor"})
			return
		}
		keys, err := client.CreateKeysAndCertificate(r.Context(), &iot.CreateKeysAndCertificateInput{SetAsActive: true})
		if err != nil {
			writeJSON(w, http.StatusBadGateway, map[string]string{"error": "could not create sensor credentials"})
			return
		}
		_, err = client.AttachThingPrincipal(r.Context(), &iot.AttachThingPrincipalInput{
			ThingName:          thing.ThingName,
			Principal:          keys.CertificateArn,
			ThingPrincipalType: types.ThingPrincipalTypeExclusiveThing,
		})
		if err == nil {
			_, err = client.AttachPolicy(r.Context(), &iot.AttachPolicyInput{PolicyName: aws.String(policyName), Target: keys.CertificateArn})
		}
		if err != nil {
			writeJSON(w, http.StatusBadGateway, map[string]string{"error": "could not finish sensor setup"})
			return
		}
		writeJSON(w, http.StatusCreated, createSensorResponse{
			ThingName:      aws.ToString(thing.ThingName),
			OwnerUserID:    ownerID,
			Endpoint:       endpoint,
			Topic:          "evap/" + ownerID + "/" + aws.ToString(thing.ThingName) + "/readings",
			CertificatePEM: aws.ToString(keys.CertificatePem),
			PrivateKey:     aws.ToString(keys.KeyPair.PrivateKey),
		})
	}
}

func ownerSlug(owner string) string {
	return strings.NewReplacer("#", "-", "/", "-", ":", "-").Replace(owner)
}

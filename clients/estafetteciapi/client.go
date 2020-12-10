package estafetteciapi

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	contracts "github.com/estafette/estafette-ci-contracts"
	foundation "github.com/estafette/estafette-foundation"
	"github.com/opentracing-contrib/go-stdlib/nethttp"
	"github.com/opentracing/opentracing-go"
	"github.com/rs/zerolog/log"
	"github.com/sethgrid/pester"
)

type Client interface {
	GetToken(ctx context.Context) (token string, err error)
	GetCatalogEntities(ctx context.Context, parentKey, parentValue, entityKey string) (entities []*contracts.CatalogEntity, err error)
	CreateCatalogEntity(ctx context.Context, entity *contracts.CatalogEntity) (err error)
	UpdateCatalogEntity(ctx context.Context, entity *contracts.CatalogEntity) (err error)
	DeleteCatalogEntity(ctx context.Context, entity *contracts.CatalogEntity) (err error)
}

// NewClient returns a new estafetteciapi.Client
func NewClient(apiBaseURL, clientID, clientSecret string) Client {
	return &client{
		apiBaseURL:   apiBaseURL,
		clientID:     clientID,
		clientSecret: clientSecret,
	}
}

type client struct {
	apiBaseURL   string
	clientID     string
	clientSecret string
	token        string
}

func (c *client) GetToken(ctx context.Context) (token string, err error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "ApiClient::GetToken")
	defer span.Finish()

	log.Debug().Msgf("Retrieving JWT token")

	clientObject := contracts.Client{
		ClientID:     c.clientID,
		ClientSecret: c.clientSecret,
	}

	bytes, err := json.Marshal(clientObject)
	if err != nil {
		return
	}

	getTokenURL := fmt.Sprintf("%v/api/auth/client/login", c.apiBaseURL)
	headers := map[string]string{
		"Content-Type": "application/json",
	}

	responseBody, err := c.postRequest(getTokenURL, span, strings.NewReader(string(bytes)), headers)

	tokenResponse := struct {
		Token string `json:"token"`
	}{}

	// unmarshal json body
	err = json.Unmarshal(responseBody, &tokenResponse)
	if err != nil {
		log.Error().Err(err).Str("body", string(responseBody)).Msgf("Failed unmarshalling get token response")
		return
	}

	// set token
	c.token = tokenResponse.Token

	return tokenResponse.Token, nil
}

func (c *client) GetCatalogEntities(ctx context.Context, parentKey, parentValue, entityKey string) (entities []*contracts.CatalogEntity, err error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "ApiClient::GetCatalogEntities")
	defer span.Finish()

	log.Debug().Msgf("Retrieving catalog entities of type %v with parent %v=%v", entityKey, parentKey, parentValue)

	pageNumber := 1
	pageSize := 100
	entities = make([]*contracts.CatalogEntity, 0)

	for {
		ents, pagination, err := c.getCatalogEntitiesPage(ctx, parentKey, parentValue, entityKey, pageNumber, pageSize)
		if err != nil {
			return entities, err
		}
		entities = append(entities, ents...)

		if pagination.TotalPages <= pageNumber {
			break
		}

		pageNumber++
	}

	span.LogKV("entities", len(entities))

	return entities, nil
}

func (c *client) getCatalogEntitiesPage(ctx context.Context, parentKey, parentValue, entityKey string, pageNumber, pageSize int) (entities []*contracts.CatalogEntity, pagination contracts.Pagination, err error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "ApiClient::getCatalogEntitiesPage")
	defer span.Finish()

	span.LogKV("page[number]", pageNumber, "page[size]", pageSize)

	parentFilter := ""
	if parentKey != "" {
		parentFilter = parentKey
		if parentValue != "" {
			parentFilter += "=" + parentValue
		}

		parentFilter = "&filter[parent]=" + url.QueryEscape(parentFilter)
	}

	entityFilter := ""
	if entityKey != "" {
		entityFilter = "&filter[entity]=" + url.QueryEscape(entityKey)
	}

	getCatalogEntitiesURL := fmt.Sprintf("%v/api/catalog/entities?page[number]=%v&page[size]=%v%v%v", c.apiBaseURL, pageNumber, pageSize, parentFilter, entityFilter)
	headers := map[string]string{
		"Authorization": fmt.Sprintf("Bearer %v", c.token),
		"Content-Type":  "application/json",
	}

	responseBody, err := c.getRequest(getCatalogEntitiesURL, span, nil, headers)
	if err != nil {
		log.Error().Err(err).Str("url", getCatalogEntitiesURL).Msgf("Failed retrieving get catalog entities response")
		return
	}

	var listResponse struct {
		Items      []*contracts.CatalogEntity `json:"items"`
		Pagination contracts.Pagination       `json:"pagination"`
	}

	// unmarshal json body
	err = json.Unmarshal(responseBody, &listResponse)
	if err != nil {
		log.Error().Err(err).Str("body", string(responseBody)).Str("url", getCatalogEntitiesURL).Msgf("Failed unmarshalling get catalog entities response")
		return
	}

	entities = listResponse.Items

	span.LogKV("entities", len(entities))

	return entities, listResponse.Pagination, nil
}

func (c *client) CreateCatalogEntity(ctx context.Context, entity *contracts.CatalogEntity) (err error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "ApiClient::CreateCatalogEntity")
	defer span.Finish()

	log.Debug().Msgf("Creating catalog entity %v=%v with parent %v=%v", entity.Key, entity.Value, entity.ParentKey, entity.ParentValue)

	span.LogKV("parent", fmt.Sprintf("%v=%v", entity.ParentKey, entity.ParentValue))
	span.LogKV("entity", fmt.Sprintf("%v=%v", entity.Key, entity.Value))

	bytes, err := json.Marshal(entity)
	if err != nil {
		return
	}

	createCatalogEntityURL := fmt.Sprintf("%v/api/catalog/entities", c.apiBaseURL)
	headers := map[string]string{
		"Authorization": fmt.Sprintf("Bearer %v", c.token),
		"Content-Type":  "application/json",
	}

	_, err = c.postRequest(createCatalogEntityURL, span, strings.NewReader(string(bytes)), headers, http.StatusCreated)
	if err != nil {
		return
	}

	return
}

func (c *client) UpdateCatalogEntity(ctx context.Context, entity *contracts.CatalogEntity) (err error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "ApiClient::UpdateCatalogEntity")
	defer span.Finish()

	log.Debug().Msgf("Updating catalog entity %v=%v with parent %v=%v and id=%v", entity.Key, entity.Value, entity.ParentKey, entity.ParentValue, entity.ID)

	span.LogKV("parent", fmt.Sprintf("%v=%v", entity.ParentKey, entity.ParentValue))
	span.LogKV("entity", fmt.Sprintf("%v=%v", entity.Key, entity.Value))

	bytes, err := json.Marshal(entity)
	if err != nil {
		return
	}

	updateCatalogEntityURL := fmt.Sprintf("%v/api/catalog/entities/%v", c.apiBaseURL, entity.ID)
	headers := map[string]string{
		"Authorization": fmt.Sprintf("Bearer %v", c.token),
		"Content-Type":  "application/json",
	}

	_, err = c.putRequest(updateCatalogEntityURL, span, strings.NewReader(string(bytes)), headers)
	if err != nil {
		return
	}

	return
}

func (c *client) DeleteCatalogEntity(ctx context.Context, entity *contracts.CatalogEntity) (err error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "ApiClient::DeleteCatalogEntity")
	defer span.Finish()

	log.Debug().Msgf("Deleting catalog entity %v=%v with parent %v=%v and id=%v", entity.Key, entity.Value, entity.ParentKey, entity.ParentValue, entity.ID)

	span.LogKV("parent", fmt.Sprintf("%v=%v", entity.ParentKey, entity.ParentValue))
	span.LogKV("entity", fmt.Sprintf("%v=%v", entity.Key, entity.Value))

	bytes, err := json.Marshal(entity)
	if err != nil {
		return
	}

	deleteCatalogEntityURL := fmt.Sprintf("%v/api/catalog/entities/%v", c.apiBaseURL, entity.ID)
	headers := map[string]string{
		"Authorization": fmt.Sprintf("Bearer %v", c.token),
		"Content-Type":  "application/json",
	}

	_, err = c.deleteRequest(deleteCatalogEntityURL, span, strings.NewReader(string(bytes)), headers)
	if err != nil {
		return
	}

	return
}

func (c *client) getRequest(uri string, span opentracing.Span, requestBody io.Reader, headers map[string]string, allowedStatusCodes ...int) (responseBody []byte, err error) {
	return c.makeRequest("GET", uri, span, requestBody, headers, allowedStatusCodes...)
}

func (c *client) postRequest(uri string, span opentracing.Span, requestBody io.Reader, headers map[string]string, allowedStatusCodes ...int) (responseBody []byte, err error) {
	return c.makeRequest("POST", uri, span, requestBody, headers, allowedStatusCodes...)
}

func (c *client) putRequest(uri string, span opentracing.Span, requestBody io.Reader, headers map[string]string, allowedStatusCodes ...int) (responseBody []byte, err error) {
	return c.makeRequest("PUT", uri, span, requestBody, headers, allowedStatusCodes...)
}

func (c *client) deleteRequest(uri string, span opentracing.Span, requestBody io.Reader, headers map[string]string, allowedStatusCodes ...int) (responseBody []byte, err error) {
	return c.makeRequest("DELETE", uri, span, requestBody, headers, allowedStatusCodes...)
}

func (c *client) makeRequest(method, uri string, span opentracing.Span, requestBody io.Reader, headers map[string]string, allowedStatusCodes ...int) (responseBody []byte, err error) {

	// create client, in order to add headers
	client := pester.NewExtendedClient(&http.Client{Transport: &nethttp.Transport{}})
	client.MaxRetries = 3
	client.Backoff = pester.ExponentialJitterBackoff
	client.KeepLog = true
	client.Timeout = time.Second * 10

	request, err := http.NewRequest(method, uri, requestBody)
	if err != nil {
		return nil, err
	}

	// add tracing context
	request = request.WithContext(opentracing.ContextWithSpan(request.Context(), span))

	// collect additional information on setting up connections
	request, ht := nethttp.TraceRequest(span.Tracer(), request)

	// add headers
	for k, v := range headers {
		request.Header.Add(k, v)
	}

	// perform actual request
	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	ht.Finish()

	if len(allowedStatusCodes) == 0 {
		allowedStatusCodes = []int{http.StatusOK}
	}

	if !foundation.IntArrayContains(allowedStatusCodes, response.StatusCode) {
		return nil, fmt.Errorf("%v %v responded with status code %v", method, uri, response.StatusCode)
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return
	}

	return body, nil
}

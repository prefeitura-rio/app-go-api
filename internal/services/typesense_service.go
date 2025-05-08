package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/prefeitura-rio/app-go-api/internal/config"
	"github.com/prefeitura-rio/app-go-api/internal/models"
	"github.com/typesense/typesense-go/v3/typesense"
	"github.com/typesense/typesense-go/v3/typesense/api"
)

// TypesenseService fornece operações para gerenciar coleções e documentos no Typesense
type TypesenseService struct {
	client *typesense.Client
}

// NewTypesenseService cria uma nova instância do TypesenseService
func NewTypesenseService() (*TypesenseService, error) {
	cfg := config.Get().Typesense
	
	if cfg.Host == "" || cfg.APIKey == "" {
		return nil, errors.New("configuração do Typesense incompleta")
	}

	client := typesense.NewClient(
		typesense.WithServer(fmt.Sprintf("%s://%s:%d", cfg.Protocol, cfg.Host, cfg.Port)),
		typesense.WithAPIKey(cfg.APIKey),
	)

	return &TypesenseService{
		client: client,
	}, nil
}

// CreateCollection cria uma nova coleção no Typesense
func (s *TypesenseService) CreateCollection(ctx context.Context, req models.CreateCollectionRequest) (*models.CollectionResponse, error) {
	// Converter para o formato esperado pelo Typesense
	var fields []api.Field
	for _, field := range req.Fields {
		fields = append(fields, api.Field{
			Name:     field.Name,
			Type:     field.Type,
			Facet:    &field.Facet,
			Index:    &field.Index,
			Optional: &field.Optional,
		})
	}

	// Ajustando os tipos para ponteiros onde necessário
	tokenSeparators := req.TokenSeparators
	symbolsToIndex := req.SymbolsToIndex

	schema := &api.CollectionSchema{
		Name:                   req.Name,
		Fields:                 fields,
		DefaultSortingField:    &req.DefaultSortingField,
		EnableNestedFields:     &req.EnableNestedFields,
		TokenSeparators:        &tokenSeparators,
		SymbolsToIndex:         &symbolsToIndex,
	}

	collection, err := s.client.Collections().Create(ctx, schema)
	if err != nil {
		return nil, err
	}

	// Converter resposta para o formato da API
	var tokenSeps []string
	if collection.TokenSeparators != nil {
		tokenSeps = *collection.TokenSeparators
	}
	
	var symToIndex []string
	if collection.SymbolsToIndex != nil {
		symToIndex = *collection.SymbolsToIndex
	}
	
	var createdAt int64
	if collection.CreatedAt != nil {
		createdAt = *collection.CreatedAt
	}
	
	var numDocs int
	if collection.NumDocuments != nil {
		numDocs = int(*collection.NumDocuments)
	}

	result := &models.CollectionResponse{
		Name:               collection.Name,
		DefaultSortingField: *collection.DefaultSortingField,
		EnableNestedFields: *collection.EnableNestedFields,
		TokenSeparators:    tokenSeps,
		SymbolsToIndex:     symToIndex,
		CreatedAt:          createdAt,
		NumDocuments:       numDocs,
	}

	// Converter campos
	for _, field := range collection.Fields {
		facet := false
		index := true
		optional := false
		if field.Facet != nil {
			facet = *field.Facet
		}
		if field.Index != nil {
			index = *field.Index
		}
		if field.Optional != nil {
			optional = *field.Optional
		}

		result.Fields = append(result.Fields, models.CollectionField{
			Name:     field.Name,
			Type:     field.Type,
			Facet:    facet,
			Index:    index,
			Optional: optional,
		})
	}

	return result, nil
}

// GetCollection obtém uma coleção pelo nome
func (s *TypesenseService) GetCollection(ctx context.Context, name string) (*models.CollectionResponse, error) {
	collection, err := s.client.Collection(name).Retrieve(ctx)
	if err != nil {
		return nil, err
	}

	// Extrair valores de ponteiros
	var tokenSeps []string
	if collection.TokenSeparators != nil {
		tokenSeps = *collection.TokenSeparators
	}
	
	var symToIndex []string
	if collection.SymbolsToIndex != nil {
		symToIndex = *collection.SymbolsToIndex
	}
	
	var createdAt int64
	if collection.CreatedAt != nil {
		createdAt = *collection.CreatedAt
	}
	
	var numDocs int
	if collection.NumDocuments != nil {
		numDocs = int(*collection.NumDocuments)
	}

	// Converter resposta para o formato da API
	result := &models.CollectionResponse{
		Name:               collection.Name,
		DefaultSortingField: *collection.DefaultSortingField,
		EnableNestedFields: *collection.EnableNestedFields,
		TokenSeparators:    tokenSeps,
		SymbolsToIndex:     symToIndex,
		CreatedAt:          createdAt,
		NumDocuments:       numDocs,
	}

	// Converter campos
	for _, field := range collection.Fields {
		facet := false
		index := true
		optional := false
		if field.Facet != nil {
			facet = *field.Facet
		}
		if field.Index != nil {
			index = *field.Index
		}
		if field.Optional != nil {
			optional = *field.Optional
		}

		result.Fields = append(result.Fields, models.CollectionField{
			Name:     field.Name,
			Type:     field.Type,
			Facet:    facet,
			Index:    index,
			Optional: optional,
		})
	}

	return result, nil
}

// ListCollections lista todas as coleções
func (s *TypesenseService) ListCollections(ctx context.Context) ([]models.CollectionResponse, error) {
	collections, err := s.client.Collections().Retrieve(ctx)
	if err != nil {
		return nil, err
	}

	var result []models.CollectionResponse
	for _, collection := range collections {
		var createdAt int64
		if collection.CreatedAt != nil {
			createdAt = *collection.CreatedAt
		}
		
		var numDocs int
		if collection.NumDocuments != nil {
			numDocs = int(*collection.NumDocuments)
		}
		
		collectionData := models.CollectionResponse{
			Name:          collection.Name,
			CreatedAt:     createdAt,
			NumDocuments:  numDocs,
		}

		if collection.DefaultSortingField != nil {
			collectionData.DefaultSortingField = *collection.DefaultSortingField
		}
		
		if collection.EnableNestedFields != nil {
			collectionData.EnableNestedFields = *collection.EnableNestedFields
		}

		for _, field := range collection.Fields {
			facet := false
			index := true
			optional := false
			if field.Facet != nil {
				facet = *field.Facet
			}
			if field.Index != nil {
				index = *field.Index
			}
			if field.Optional != nil {
				optional = *field.Optional
			}

			collectionData.Fields = append(collectionData.Fields, models.CollectionField{
				Name:     field.Name,
				Type:     field.Type,
				Facet:    facet,
				Index:    index,
				Optional: optional,
			})
		}

		result = append(result, collectionData)
	}

	return result, nil
}

// DeleteCollection exclui uma coleção pelo nome
func (s *TypesenseService) DeleteCollection(ctx context.Context, name string) error {
	_, err := s.client.Collection(name).Delete(ctx)
	return err
}

// UpsertDocument insere ou atualiza um documento em uma coleção
func (s *TypesenseService) UpsertDocument(ctx context.Context, collectionName string, document models.Document, action string) (models.Document, error) {
	// Criar os parâmetros de indexação sem o campo Action que não existe
	indexParams := &api.DocumentIndexParameters{}

	// Serializar o documento para JSON
	docBytes, err := json.Marshal(document)
	if err != nil {
		return nil, err
	}

	// Deserializar para map[string]interface{} para compatibilidade com a API do Typesense
	var docMap map[string]interface{}
	if err := json.Unmarshal(docBytes, &docMap); err != nil {
		return nil, err
	}

	result, err := s.client.Collection(collectionName).Documents().Upsert(ctx, docMap, indexParams)
	if err != nil {
		return nil, err
	}

	// Converter resposta de volta para Document
	resultBytes, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}

	var resultDoc models.Document
	if err := json.Unmarshal(resultBytes, &resultDoc); err != nil {
		return nil, err
	}

	return resultDoc, nil
}

// GetDocument obtém um documento pelo ID
func (s *TypesenseService) GetDocument(ctx context.Context, collectionName, docID string) (models.Document, error) {
	doc, err := s.client.Collection(collectionName).Document(docID).Retrieve(ctx)
	if err != nil {
		return nil, err
	}

	// Converter resposta de volta para Document
	docBytes, err := json.Marshal(doc)
	if err != nil {
		return nil, err
	}

	var resultDoc models.Document
	if err := json.Unmarshal(docBytes, &resultDoc); err != nil {
		return nil, err
	}

	return resultDoc, nil
}

// DeleteDocument exclui um documento pelo ID
func (s *TypesenseService) DeleteDocument(ctx context.Context, collectionName, docID string) error {
	_, err := s.client.Collection(collectionName).Document(docID).Delete(ctx)
	return err
}

// SearchDocuments busca documentos em uma coleção
func (s *TypesenseService) SearchDocuments(ctx context.Context, collectionName string, params models.SearchParameters) (*models.SearchDocumentsResponse, error) {
	// Converte os valores para ponteiros quando necessário
	q := params.Q
	queryBy := params.QueryBy
	filterBy := params.FilterBy
	sortBy := params.SortBy
	facetBy := params.FacetBy
	maxFacetValues := params.MaxFacetValues
	facetQuery := params.FacetQuery
	includeFields := params.IncludeFields
	excludeFields := params.ExcludeFields
	highlightFields := params.HighlightFields
	highlightFullFields := params.HighlightFullFields
	prefix := fmt.Sprintf("%v", params.Prefix) // Convertendo bool para string
	numTypos := fmt.Sprintf("%d", params.NumTypos) // Convertendo int para string

	searchParams := &api.SearchCollectionParams{
		Q:                   &q,
		QueryBy:             &queryBy,
		FilterBy:            &filterBy,
		SortBy:              &sortBy,
		FacetBy:             &facetBy,
		MaxFacetValues:      &maxFacetValues,
		FacetQuery:          &facetQuery,
		IncludeFields:       &includeFields,
		ExcludeFields:       &excludeFields,
		HighlightFields:     &highlightFields,
		HighlightFullFields: &highlightFullFields,
		Prefix:              &prefix,
		NumTypos:            &numTypos,
	}

	if params.Page > 0 {
		searchParams.Page = &params.Page
	}

	if params.PerPage > 0 {
		searchParams.PerPage = &params.PerPage
	}

	result, err := s.client.Collection(collectionName).Documents().Search(ctx, searchParams)
	if err != nil {
		return nil, err
	}

	// Converter resultado para o formato da resposta
	var found int
	if result.Found != nil {
		found = *result.Found
	}
	
	response := &models.SearchDocumentsResponse{
		Found: found,
		Page:  *result.Page,
	}

	// Converter os hits
	if result.Hits != nil {
		for _, hit := range *result.Hits {
			hitData := make(map[string]interface{})
			// Converter o documento
			docBytes, err := json.Marshal(hit.Document)
			if err != nil {
				return nil, err
			}
			
			if err := json.Unmarshal(docBytes, &hitData); err != nil {
				return nil, err
			}

			// Adicionar outros campos do hit
			hitData["highlights"] = hit.Highlights
			response.Hits = append(response.Hits, hitData)
		}
	}

	// Converter parâmetros da requisição
	paramsBytes, err := json.Marshal(result.RequestParams)
	if err != nil {
		return nil, err
	}
	
	if err := json.Unmarshal(paramsBytes, &response.RequestParams); err != nil {
		return nil, err
	}

	return response, nil
}

// ImportDocuments importa múltiplos documentos em batch
func (s *TypesenseService) ImportDocuments(ctx context.Context, collectionName string, documents []models.Document, action string) (int, error) {
	// Criar parâmetros sem o campo Action inválido
	options := &api.ImportDocumentsParams{}

	// Preparar os documentos como array de interface{}
	var docsInterface []interface{}
	for _, doc := range documents {
		docBytes, err := json.Marshal(doc)
		if err != nil {
			return 0, err
		}
		var docMap map[string]interface{}
		if err := json.Unmarshal(docBytes, &docMap); err != nil {
			return 0, err
		}
		docsInterface = append(docsInterface, docMap)
	}

	result, err := s.client.Collection(collectionName).Documents().Import(ctx, docsInterface, options)
	if err != nil {
		return 0, err
	}

	// Contar documentos importados com sucesso
	successCount := 0
	for _, item := range result {
		if item.Success {
			successCount++
		}
	}

	return successCount, nil
}

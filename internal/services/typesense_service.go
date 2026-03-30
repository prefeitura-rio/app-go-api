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

// TypesenseServiceInterface define os métodos do serviço de busca Typesense
type TypesenseServiceInterface interface {
	SearchCursos(ctx context.Context, params models.CursoSearchParameters) (*models.SearchDocumentsResponse, error)
	SearchEmpregos(ctx context.Context, params models.EmpregoSearchParameters) (*models.SearchDocumentsResponse, error)
	SearchDocuments(ctx context.Context, collectionName string, params models.SearchParameters) (*models.SearchDocumentsResponse, error)
	SearchMultiCollection(ctx context.Context, params models.MultiCollectionSearchParameters) (*models.MultiCollectionSearchResponse, error)
}

// TypesenseService fornece operações de busca no Typesense
type TypesenseService struct {
	client *typesense.Client
}

// NewTypesenseService cria uma nova instância do TypesenseService
func NewTypesenseService() (*TypesenseService, error) {
	cfg, err := config.Get()
	if err != nil {
		return nil, fmt.Errorf("erro ao obter configurações: %w", err)
	}

	if cfg.TypeSense.Host == "" || cfg.TypeSense.APIKey == "" {
		return nil, errors.New("configuração do Typesense incompleta")
	}

	client := typesense.NewClient(
		typesense.WithServer(fmt.Sprintf("%s://%s:%d", cfg.TypeSense.Protocol, cfg.TypeSense.Host, cfg.TypeSense.Port)),
		typesense.WithAPIKey(cfg.TypeSense.APIKey),
	)

	return &TypesenseService{
		client: client,
	}, nil
}

// SearchCursos busca cursos usando parâmetros específicos e otimizados para a estrutura aninhada
func (s *TypesenseService) SearchCursos(ctx context.Context, params models.CursoSearchParameters) (*models.SearchDocumentsResponse, error) {
	// Converter para parâmetros padrão
	searchParams := params.ToSearchParameters()

	// Executar busca na collection de cursos
	response, err := s.SearchDocuments(ctx, "cursos", searchParams)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar cursos: %w", err)
	}

	// Processar hits para extrair apenas dados do record
	if response.Hits != nil {
		for i, hit := range response.Hits {
			if record, exists := hit["record"]; exists {
				// Mover dados do record para o nível principal do hit
				if recordMap, ok := record.(map[string]interface{}); ok {
					// Preservar campos importantes do documento original
					recordMap["document_id"] = hit["id"]
					recordMap["action"] = hit["action"]

					// Preservar highlights se existirem
					if highlights, exists := hit["highlights"]; exists {
						recordMap["highlights"] = highlights
					}

					// Substituir o hit pelos dados do record
					response.Hits[i] = recordMap
				}
			}
		}
	}

	return response, nil
}

// SearchEmpregos busca empregos usando parâmetros específicos e otimizados para a estrutura aninhada
func (s *TypesenseService) SearchEmpregos(ctx context.Context, params models.EmpregoSearchParameters) (*models.SearchDocumentsResponse, error) {
	// Converter para parâmetros padrão
	searchParams := params.ToSearchParameters()

	// Executar busca na collection de empregos
	response, err := s.SearchDocuments(ctx, "empregos", searchParams)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar empregos: %w", err)
	}

	// Processar hits para extrair apenas dados do record
	if response.Hits != nil {
		for i, hit := range response.Hits {
			if record, exists := hit["record"]; exists {
				// Mover dados do record para o nível principal do hit
				if recordMap, ok := record.(map[string]interface{}); ok {
					// Preservar campos importantes do documento original
					recordMap["document_id"] = hit["id"]
					recordMap["action"] = hit["action"]

					// Preservar highlights se existirem
					if highlights, exists := hit["highlights"]; exists {
						recordMap["highlights"] = highlights
					}

					// Substituir o hit pelos dados do record
					response.Hits[i] = recordMap
				}
			}
		}
	}

	return response, nil
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
	prefix := fmt.Sprintf("%v", params.Prefix)     // Convertendo bool para string
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

// SearchMultiCollection busca documentos em múltiplas coleções
func (s *TypesenseService) SearchMultiCollection(ctx context.Context, params models.MultiCollectionSearchParameters) (*models.MultiCollectionSearchResponse, error) {
	if len(params.Collections) == 0 {
		return nil, errors.New("nenhuma coleção especificada para busca")
	}

	// Inicializa resposta com mapa de resultados
	response := &models.MultiCollectionSearchResponse{
		Results: make(map[string]models.SearchDocumentsResponse),
	}

	// Contador de total de documentos encontrados em todas as coleções
	totalFound := 0

	// Busca em cada coleção
	for _, collectionName := range params.Collections {
		// Executa a busca na coleção atual
		result, err := s.SearchDocuments(ctx, collectionName, params.Params)
		if err != nil {
			// Se houver erro em uma coleção, continua com as próximas
			// mas registra o erro nos resultados
			response.Results[collectionName] = models.SearchDocumentsResponse{
				Found: 0,
				Hits:  []map[string]interface{}{},
				RequestParams: map[string]interface{}{
					"error": err.Error(),
				},
			}
			continue
		}

		// Adiciona resultado ao mapa de resultados
		response.Results[collectionName] = *result

		// Incrementa o contador total
		totalFound += result.Found
	}

	// Atualiza o total encontrado
	response.TotalFound = totalFound

	return response, nil
}

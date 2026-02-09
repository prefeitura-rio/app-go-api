package empregabilidade

import (
	"github.com/prefeitura-rio/app-go-api/internal/handlers/common"
	"github.com/prefeitura-rio/app-go-api/internal/models/empregabilidade"
	services "github.com/prefeitura-rio/app-go-api/internal/services/empregabilidade"
)

type ModeloTrabalhoHandler = common.CRUDHandlerUUID[empregabilidade.ModeloTrabalho]

func NewModeloTrabalhoHandler(service *services.ModeloTrabalhoService) *ModeloTrabalhoHandler {
	return common.NewCRUDHandlerUUID[empregabilidade.ModeloTrabalho](service, "Modelo de trabalho")
}

type TipoPCDHandler = common.CRUDHandlerUUID[empregabilidade.TipoPCD]

func NewTipoPCDHandler(service *services.TipoPCDService) *TipoPCDHandler {
	return common.NewCRUDHandlerUUID[empregabilidade.TipoPCD](service, "Tipo PCD")
}

type IdiomaHandler = common.CRUDHandlerUUID[empregabilidade.Idioma]

func NewIdiomaHandler(service *services.IdiomaService) *IdiomaHandler {
	return common.NewCRUDHandlerUUID[empregabilidade.Idioma](service, "Idioma")
}

type NivelIdiomaHandler = common.CRUDHandlerUUID[empregabilidade.NivelIdioma]

func NewNivelIdiomaHandler(service *services.NivelIdiomaService) *NivelIdiomaHandler {
	return common.NewCRUDHandlerUUID[empregabilidade.NivelIdioma](service, "Nível de idioma")
}

type EscolaridadeHandler = common.CRUDHandlerUUID[empregabilidade.Escolaridade]

func NewEscolaridadeHandler(service *services.EscolaridadeService) *EscolaridadeHandler {
	return common.NewCRUDHandlerUUID[empregabilidade.Escolaridade](service, "Escolaridade")
}

type TipoConquistaHandler = common.CRUDHandlerUUID[empregabilidade.TipoConquista]

func NewTipoConquistaHandler(service *services.TipoConquistaService) *TipoConquistaHandler {
	return common.NewCRUDHandlerUUID[empregabilidade.TipoConquista](service, "Tipo de conquista")
}

type SituacaoAtualHandler = common.CRUDHandlerUUID[empregabilidade.SituacaoAtual]

func NewSituacaoAtualHandler(service *services.SituacaoAtualService) *SituacaoAtualHandler {
	return common.NewCRUDHandlerUUID[empregabilidade.SituacaoAtual](service, "Situação atual")
}

type DisponibilidadeHandler = common.CRUDHandlerUUID[empregabilidade.Disponibilidade]

func NewDisponibilidadeHandler(service *services.DisponibilidadeService) *DisponibilidadeHandler {
	return common.NewCRUDHandlerUUID[empregabilidade.Disponibilidade](service, "Disponibilidade")
}
